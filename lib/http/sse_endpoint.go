// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	liberrors "github.com/shuLhan/share/lib/errors"
)

// SSECallback define the handler for Server-Sent Events (SSE).
//
// SSECallback type pass original HTTP request.
// This allow the server to check for header "Last-Event-ID" and/or for
// authentication.
// Remember that "the original Request.Body must not be used" according to
// [http.Hijacker] documentation.
type SSECallback func(sse *SSEEndpoint, req *http.Request)

// SSEEndpoint endpoint to create Server-Sent Events (SSE) on server.
//
// For creating the SSE client see subpackage [sseclient].
type SSEEndpoint struct {
	bufrw *bufio.ReadWriter
	conn  net.Conn

	// Path where server accept the request for SSE.
	Path string

	// Call handler that will called when request to Path accepted.
	Call SSECallback
}

// WriteEvent write message with event type to client.
//
// The event parameter must not be empty, otherwise it will not be sent.
//
// The msg parameter must not be empty, otherwise it will not be sent
// If msg contains new line character ('\n'), the message will be split into
// multiple "data:".
//
// The id parameter is optional.
// If its nil, it will be ignored.
// if its non-nil and empty, it will be send as empty ID.
//
// It will return an error if its failed to write to peer connection.
func (ep *SSEEndpoint) WriteEvent(event, msg string, id *string) (err error) {
	event = strings.TrimSpace(event)
	if len(event) == 0 {
		return nil
	}
	if len(msg) == 0 {
		return nil
	}

	var buf bytes.Buffer

	buf.WriteString(`event:`)
	buf.WriteString(event)
	buf.WriteByte('\n')

	ep.writeData(&buf, msg, id)

	_, err = ep.bufrw.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf(`WriteMessage: %w`, err)
	}
	ep.bufrw.Flush()
	return nil
}

// WriteMessage write a message with optional id to client.
//
// The msg parameter must not be empty, otherwise it will not be sent
// If msg contains new line character ('\n'), the message will be split into
// multiple "data:".
//
// The id parameter is optional.
// If its nil, it will be ignored.
// if its non-nil and empty, it will be send as empty ID.
//
// It will return an error if its failed to write to peer connection.
func (ep *SSEEndpoint) WriteMessage(msg string, id *string) (err error) {
	if len(msg) == 0 {
		return nil
	}

	var buf bytes.Buffer

	ep.writeData(&buf, msg, id)

	_, err = ep.bufrw.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf(`WriteMessage: %w`, err)
	}
	ep.bufrw.Flush()
	return nil
}

// WriteRaw write raw event message directly, without any parsing.
func (ep *SSEEndpoint) WriteRaw(msg []byte) (err error) {
	_, err = ep.bufrw.Write(msg)
	if err != nil {
		return fmt.Errorf(`WriteRaw: %w`, err)
	}
	ep.bufrw.Flush()
	return nil
}

// WriteRetry inform user how long they should wait, after disconnect,
// before re-connecting back to server.
//
// The duration must be in millisecond.
func (ep *SSEEndpoint) WriteRetry(retry time.Duration) (err error) {
	_, err = fmt.Fprintf(ep.bufrw, "retry:%d\n\n", retry.Milliseconds())
	if err != nil {
		return fmt.Errorf(`WriteRetry: %w`, err)
	}
	ep.bufrw.Flush()
	return nil
}

func (ep *SSEEndpoint) writeData(buf *bytes.Buffer, msg string, id *string) {
	var (
		lines = strings.Split(msg, "\n")
		line  string
	)
	for _, line = range lines {
		buf.WriteString(`data:`)
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if id != nil {
		buf.WriteString(`id:`)
		buf.WriteString(*id)
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')
}

func (ep *SSEEndpoint) call(
	res http.ResponseWriter,
	req *http.Request,
	evaluators []Evaluator,
	vals map[string]string,
) {
	var err error

	err = req.ParseForm()
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	// Fill the form with path binding.
	if len(vals) > 0 {
		if req.Form == nil {
			req.Form = make(url.Values, len(vals))
		}
		var k, v string
		for k, v = range vals {
			if len(k) > 0 && len(v) > 0 {
				req.Form.Set(k, v)
			}
		}
	}

	err = ep.doEvals(res, req, evaluators)
	if err != nil {
		return
	}

	err = ep.hijack(res)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	ep.handshake()
	ep.Call(ep, req)
	ep.conn.Close()
}

func (ep *SSEEndpoint) doEvals(
	res http.ResponseWriter,
	req *http.Request,
	evaluators []Evaluator,
) (err error) {
	var eval Evaluator

	for _, eval = range evaluators {
		err = eval(req, nil)
		if err != nil {
			var errInternal = &liberrors.E{}
			if !errors.As(err, &errInternal) {
				errInternal.Code = http.StatusUnprocessableEntity
			}
			http.Error(res, err.Error(), errInternal.Code)
			return err
		}
	}
	return nil
}

func (ep *SSEEndpoint) hijack(res http.ResponseWriter) (err error) {
	var (
		hijack http.Hijacker
		ok     bool
	)

	hijack, ok = res.(http.Hijacker)
	if !ok {
		return errors.New(`http.ResponseWriter is not http.Hijacker`)
	}

	ep.conn, ep.bufrw, err = hijack.Hijack()
	if err != nil {
		return err
	}

	return nil
}

// handshake write the last HTTP response to indicate the connection is
// accepted.
func (ep *SSEEndpoint) handshake() {
	ep.bufrw.WriteString("HTTP/1.1 200 OK\r\n")
	ep.bufrw.WriteString("content-type: text/event-stream\r\n")
	ep.bufrw.WriteString("cache-control: no-cache\r\n")
	ep.bufrw.WriteString("\r\n")
	ep.bufrw.Flush()
}
