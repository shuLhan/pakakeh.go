// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// SSECallback define the handler for Server-Sent Events (SSE).
//
// SSECallback type pass SSEConn that contains original HTTP request.
// This allow the server to check for header "Last-Event-ID" and/or for
// authentication.
// Remember that "the original Request.Body must not be used" according to
// [http.Hijacker] documentation.
type SSECallback func(sse *SSEConn)

// SSEConn define the connection when the SSE request accepted by server.
type SSEConn struct {
	HttpRequest *http.Request //revive:disable-line

	bufrw *bufio.ReadWriter
	conn  net.Conn
}

// WriteEvent write message with optional event type and id to client.
//
// The "event" parameter is optional.
// If its empty, no "event:" line will be send to client.
//
// The "data" parameter must not be empty, otherwise no message will be
// send.
// If "data" value contains new line character ('\n'), the message will be
// split into multiple "data:".
//
// The id parameter is optional.
// If its nil, it will be ignored.
// if its non-nil and empty, it will be send as empty ID.
//
// It will return an error if its failed to write to peer connection.
func (ep *SSEConn) WriteEvent(event, data string, id *string) (err error) {
	event = strings.TrimSpace(event)
	if len(data) == 0 {
		return nil
	}

	var buf bytes.Buffer

	if len(event) != 0 {
		buf.WriteString(`event:`)
		buf.WriteString(event)
		buf.WriteByte('\n')
	}

	ep.writeData(&buf, data, id)

	_, err = ep.bufrw.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf(`WriteEvent: %w`, err)
	}
	ep.bufrw.Flush()
	return nil
}

// WriteRaw write raw event message directly, without any parsing.
func (ep *SSEConn) WriteRaw(msg []byte) (err error) {
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
func (ep *SSEConn) WriteRetry(retry time.Duration) (err error) {
	_, err = fmt.Fprintf(ep.bufrw, "retry:%d\n\n", retry.Milliseconds())
	if err != nil {
		return fmt.Errorf(`WriteRetry: %w`, err)
	}
	ep.bufrw.Flush()
	return nil
}

// workerKeepAlive periodically send an empty message to client to keep the
// connection alive.
func (ep *SSEConn) workerKeepAlive(interval time.Duration) {
	var (
		ticker   = time.NewTicker(interval)
		emptyMsg = []byte(":\n\n")

		err error
	)
	for _ = range ticker.C {
		err = ep.WriteRaw(emptyMsg)
		if err != nil {
			// Write failed, probably connection has been
			// closed.
			ticker.Stop()
			return
		}
	}
}

func (ep *SSEConn) writeData(buf *bytes.Buffer, data string, id *string) {
	var (
		lines = strings.Split(data, "\n")
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

// handshake write the last HTTP response to indicate the connection is
// accepted.
func (ep *SSEConn) handshake() {
	ep.bufrw.WriteString("HTTP/1.1 200 OK\r\n")
	ep.bufrw.WriteString("content-type: text/event-stream\r\n")
	ep.bufrw.WriteString("cache-control: no-cache\r\n")
	ep.bufrw.WriteString("\r\n")
	ep.bufrw.Flush()
}
