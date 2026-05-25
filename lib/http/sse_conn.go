// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package http

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SSECallback defines the handler for Server-Sent Events (SSE).
//
// The SSECallback receives [SSEConn] that contains original HTTP request.
// This allow the server to check for header "Last-Event-ID" and/or for
// authentication.
// Remember that "the original [http.Request.Body] must not be used"
// according to [http.Hijacker] documentation.
type SSECallback func(sse *SSEConn)

// SSEConn defines the connection when the SSE request accepted by server.
type SSEConn struct {
	HTTPRequest *http.Request

	bufrw *bufio.ReadWriter
	conn  net.Conn

	// bufrwMtx protects the concurrent write between client and
	// workerKeepAlive.
	bufrwMtx sync.Mutex
}

// WriteEvent writes message with optional event type and id to client.
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

	err = ep.WriteRaw(buf.Bytes())
	if err != nil {
		return fmt.Errorf(`WriteEvent: %w`, err)
	}
	return nil
}

// WriteRaw writes raw event msg directly, without any parsing.
func (ep *SSEConn) WriteRaw(msg []byte) (err error) {
	ep.bufrwMtx.Lock()
	_, err = ep.bufrw.Write(msg)
	if err != nil {
		ep.bufrwMtx.Unlock()
		return fmt.Errorf(`WriteRaw: %w`, err)
	}
	ep.bufrw.Flush()
	ep.bufrwMtx.Unlock()
	return nil
}

// WriteRetry writes "retry:" event to inform client how long they should
// wait, after disconnect, before re-connecting back to server.
// The retry duration must be in millisecond.
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
	for range ticker.C {
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

// handshake writes the last HTTP response to indicate the connection is
// accepted.
func (ep *SSEConn) handshake() {
	var (
		bb  bytes.Buffer
		err error
	)

	bb.WriteString("HTTP/1.1 200 OK\r\n")
	bb.WriteString("content-type: text/event-stream\r\n")
	bb.WriteString("cache-control: no-cache\r\n")
	bb.WriteString("\r\n")

	_, err = ep.bufrw.Write(bb.Bytes())
	if err == nil {
		ep.bufrw.Flush()
	}
}
