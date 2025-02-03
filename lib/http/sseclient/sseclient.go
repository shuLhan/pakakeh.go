// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package sseclient implement HTTP client for Server-Sent Events (SSE).
//
// # Notes on implementation
//
// The SSE specification have inconsistent state when dispatching empty
// data.
// In the "9.2.6 Interpreting an event stream", if the data buffer is empty
// it would return; but in the third example it can dispatch an empty
// string.
// In this implement we ignore an empty string in server and client.
//
// References,
//   - [whatwg.org Server-sent events]
//
// [whatwg.org Server-sent events]: https://html.spec.whatwg.org/multipage/server-sent-events.html
package sseclient

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	pakakeh "git.sr.ht/~shulhan/pakakeh.go"
	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

const defTimeout = 10 * time.Second

// defEventBuffer define maximum event buffered in channel.
const defEventBuffer = 1024

// Client for SSE.
// Once the Client filled, user need only to call Connect to start receiving
// message from channel C.
type Client struct {
	C     <-chan Event
	event chan Event

	serverURL *url.URL
	header    http.Header

	conn   net.Conn
	closeq chan struct{}

	// Endpoint define the HTTP server URL to connect.
	Endpoint string

	// LastEventID define the last event ID to be sent during handshake.
	// Once the handshake success, this field will be reset and may set
	// with next ID from server.
	// This field is optional.
	LastEventID string

	// Timeout define the read and write timeout when reading and
	// writing from/to server.
	// This field is optional default to 10 seconds.
	Timeout time.Duration

	// Retry define how long, in milliseconds, the client should wait
	// before reconnecting back to server after disconnect.
	// Zero or negative value disable it.
	//
	// This field is optional, default to 0 (not retrying).
	Retry time.Duration

	// Insecure allow connect to HTTPS endpoint with invalid
	// certificate.
	Insecure bool
}

// Close the connection and release all resources.
func (cl *Client) Close() (err error) {
	// Close the connection, wait until it catched by consume goroutine.
	if cl.conn != nil {
		err = cl.conn.Close()

		var timeWait = time.NewTimer(50 * time.Millisecond)
		select {
		case cl.closeq <- struct{}{}:
			// Tell the consume goroutine we initiate the close.
		case <-timeWait.C:
			// The consume goroutine may already end or end at
			// the same time.
		}
		cl.conn = nil
	}
	return err
}

// Connect to server and start consume the message and propagate to each
// registered handlers.
//
// The header parameter define custom, optional HTTP header to be sent
// during handshake.
// The following header cannot be set: Host, User-Agent, and Accept.
func (cl *Client) Connect(header http.Header) (err error) {
	var logp = `Connect`

	err = cl.init(header)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	err = cl.connect()
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	// Reset the ID to store the ID from server.
	cl.LastEventID = ``

	go cl.consume()

	return nil
}

func (cl *Client) connect() (err error) {
	err = cl.dial()
	if err != nil {
		return err
	}

	var packet []byte

	packet, err = cl.handshake()
	if err != nil {
		return err
	}

	select {
	case cl.event <- Event{Type: EventTypeOpen}:
	default:
	}

	// The HTTP response may contains events in the body,
	// consume it.
	cl.parseEvent(packet)

	return nil
}

// init validate and set default field values.
func (cl *Client) init(header http.Header) (err error) {
	cl.serverURL, err = url.Parse(cl.Endpoint)
	if err != nil {
		return err
	}

	var host, port string

	host, port, err = net.SplitHostPort(cl.serverURL.Host)
	if err != nil {
		return err
	}
	if len(port) == 0 {
		switch cl.serverURL.Scheme {
		case `http`:
			port = `80`
		case `https`:
			port = `443`
		default:
			return fmt.Errorf(`unknown scheme %q`, cl.serverURL.Scheme)
		}
	}
	cl.serverURL.Host = net.JoinHostPort(host, port)

	cl.header = header
	if cl.header == nil {
		cl.header = http.Header{}
	}
	cl.header.Set(libhttp.HeaderHost, cl.serverURL.Host)
	cl.header.Set(libhttp.HeaderUserAgent, `libhttp/`+pakakeh.Version)
	cl.header.Set(libhttp.HeaderAccept, libhttp.ContentTypeEventStream)

	if cl.Timeout <= 0 {
		cl.Timeout = defTimeout
	}

	cl.event = make(chan Event, defEventBuffer)
	cl.C = cl.event
	cl.closeq = make(chan struct{})

	return nil
}

func (cl *Client) dial() (err error) {
	if cl.serverURL.Scheme == `https` {
		var tlsConfig = &tls.Config{
			InsecureSkipVerify: cl.Insecure,
		}
		cl.conn, err = tls.Dial(`tcp`, cl.serverURL.Host, tlsConfig)
	} else {
		cl.conn, err = net.Dial(`tcp`, cl.serverURL.Host)
	}
	if err != nil {
		return err
	}
	return nil
}

// handshake send the HTTP request and check for the response.
// The response must be HTTP status code 200 with Content-Type
// "text/event-stream".
//
// If the response is not empty, it contains event message, return it.
func (cl *Client) handshake() (packet []byte, err error) {
	err = cl.handshakeRequest()
	if err != nil {
		return nil, err
	}

	packet, err = libnet.Read(cl.conn, 0, cl.Timeout)
	if err != nil {
		return nil, err
	}

	var httpRes *http.Response

	httpRes, packet, err = libhttp.ParseResponseHeader(packet)
	if err != nil {
		return nil, err
	}

	if httpRes.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`handshake failed with response status %q`, httpRes.Status)
	}

	var contentType = httpRes.Header.Get(libhttp.HeaderContentType)
	if contentType != libhttp.ContentTypeEventStream {
		return nil, fmt.Errorf(`handshake failed with unknown Content-Type %q`, contentType)
	}

	return packet, nil
}

func (cl *Client) handshakeRequest() (err error) {
	var logp = `handshakeRequest`
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `GET %s`, cl.serverURL.Path)
	if len(cl.serverURL.RawQuery) != 0 {
		buf.WriteByte('?')
		buf.WriteString(cl.serverURL.RawQuery)
	}
	buf.WriteString(" HTTP/1.1\r\n")

	// Write the known values to prevent user overwrite our default
	// values.

	if len(cl.LastEventID) != 0 {
		cl.header.Set(libhttp.HeaderLastEventID, cl.LastEventID)
	}

	var (
		hkey  string
		hvals []string
		val   string
	)
	for hkey, hvals = range cl.header {
		if len(hvals) == 0 {
			continue
		}
		if len(hvals) == 1 {
			val = hvals[0]
		} else {
			val = strings.Join(hvals, `,`)
		}
		fmt.Fprintf(&buf, "%s: %s\r\n", hkey, val)
	}
	buf.WriteString("\r\n")

	var deadline = time.Now().Add(cl.Timeout)

	err = cl.conn.SetWriteDeadline(deadline)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var (
		buflen = buf.Len()
		n      int
	)

	n, err = cl.conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	if n != buflen {
		return fmt.Errorf(`handshake write error, %d out of %d`, n, buflen)
	}
	return nil
}

func (cl *Client) consume() {
	var (
		timeWait  *time.Timer
		data      []byte
		err       error
		connected bool
	)
	for {
		data, err = libnet.Read(cl.conn, 0, cl.Timeout)
		if err == nil {
			cl.parseEvent(data)
			continue
		}
		if cl.Retry <= 0 {
			// Set timeout to check if connection Close-d
			// by user.
			timeWait = time.NewTimer(100 * time.Millisecond)
		} else {
			timeWait = time.NewTimer(cl.Retry)
		}
		connected = false
		for !connected {
			select {
			case <-timeWait.C:
				if cl.Retry <= 0 {
					// Retry actually not set,
					// we close connection here.
					_ = cl.conn.Close()
					cl.conn = nil
					return
				}

				err = cl.connect()
				if err != nil {
					timeWait.Reset(cl.Retry)
					continue
				}
				connected = true
			case <-cl.closeq:
				// User initiated close.
				if !timeWait.Stop() {
					<-timeWait.C
				}
				return
			}
		}
	}
}

// parseEvent parse the raw event and publish it when ready.
func (cl *Client) parseEvent(raw []byte) {
	if len(raw) == 0 {
		return
	}

	// Normalize the line ending into "\n" only.
	var lineEnd = []byte{'\n'}
	raw = bytes.ReplaceAll(raw, []byte{'\r', '\n'}, lineEnd)
	raw = bytes.ReplaceAll(raw, []byte{'\r'}, lineEnd)

	var (
		fieldSep = []byte{':'}
		lines    = bytes.Split(raw, lineEnd)

		ev    Event
		data  bytes.Buffer
		line  []byte
		fname []byte
		fval  []byte
		err   error

		// counter count each passing "data:" event.
		// When receiving empty line, the counter will reset to 0.
		counter int
	)

	ev.reset(cl.LastEventID)

	for _, line = range lines {
		if len(line) == 0 {
			// An empty line trigger dispatching the message.
			if counter == 0 {
				// Skip continuous empty line.
				continue
			}

			ev.Data = data.String()
			if len(ev.Data) != 0 {
				select {
				case cl.event <- ev:
				default:
				}
				data.Reset()
				if ev.ID != cl.LastEventID {
					// Only set LastEventID if message
					// is complete.
					cl.LastEventID = ev.ID
				}
			}
			ev.reset(cl.LastEventID)
			counter = 0
			continue
		}

		if line[0] == ':' {
			continue
		}

		// ABNF syntax for field:
		//
		// 1*name-char [ colon [ space ] *any-char ] end-of-line
		//
		//   - There is no space in field name.
		//     So, field line like "event :E" will be ignored.
		//   - There is only one space allowed after colon.

		fname, fval, _ = bytes.Cut(line, fieldSep)

		if len(fval) != 0 {
			if fval[0] == ' ' {
				fval = fval[1:]
			}
		}

		switch string(fname) {
		case `event`:
			fval = bytes.TrimSpace(fval)
			if len(fval) != 0 {
				ev.Type = string(fval)
			}
		case `data`:
			if counter > 0 {
				data.WriteByte('\n')
			}
			data.Write(fval)
			counter++
		case `id`:
			ev.ID = string(fval)
		case `retry`:
			var retry int64
			retry, err = strconv.ParseInt(string(fval), 10, 64)
			if err == nil {
				cl.Retry = time.Duration(retry) * time.Millisecond
			}
		default:
			// Ignore the field.
		}
	}
	// Ignore incomplete event that does end with empty line.
}
