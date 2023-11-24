// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sseclient

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	libhttp "github.com/shuLhan/share/lib/http"
	libnet "github.com/shuLhan/share/lib/net"
	"github.com/shuLhan/share/lib/test"
)

func TestClient(t *testing.T) {
	type testCase struct {
		kind string
		data string
		id   func() *string
		exp  Event
	}

	var fnoid = func() *string { return nil }

	var cases = []testCase{{
		kind: EventTypeOpen,
		exp: Event{
			Type: EventTypeOpen,
		},
	}, {
		kind: EventTypeMessage,
		data: `Hello, world`,
		id:   fnoid,
		exp: Event{
			Type: EventTypeMessage,
			Data: `Hello, world`,
		},
	}, {
		kind: EventTypeMessage,
		data: "Hello\nmulti\nline\nworld",
		id:   fnoid,
		exp: Event{
			Type: EventTypeMessage,
			Data: "Hello\nmulti\nline\nworld",
		},
	}, {
		kind: `join`,
		data: `John join the event`,
		id:   fnoid,
		exp: Event{
			Type: `join`,
			Data: `John join the event`,
		},
	}, {
		kind: `join`,
		data: `Jane join the event`,
		id:   func() *string { id := `1`; return &id },
		exp: Event{
			Type: `join`,
			Data: `Jane join the event`,
			ID:   `1`,
		},
	}}

	var expq = make(chan Event)

	var servercb = func(ep *libhttp.SSEEndpoint, _ *http.Request) {
		var (
			c   testCase
			err error
			x   int
		)
		for x, c = range cases {
			switch c.kind {
			case EventTypeOpen:
				// NO-OP, this is sent during connect.
			case EventTypeError:
				// NO-OP, this is sent when client has an
				// error.
			case EventTypeMessage:
				err = ep.WriteMessage(c.data, c.id())
				if err != nil {
					t.Fatal(`WriteMessage`, err)
				}
			default:
				// Named type.
				err = ep.WriteEvent(c.kind, c.data, c.id())
				if err != nil {
					t.Fatalf(`WriteEvent #%d: %s`, x, err)
				}
			}
			expq <- c.exp
		}
	}

	var (
		serverAddress string
		err           error
	)

	serverAddress, err = testRunSSEServer(t, servercb)
	if err != nil {
		t.Fatal(`testRunSSEServer:`, err)
	}

	var cl = Client{
		Endpoint: fmt.Sprintf(`http://%s/sse`, serverAddress),
	}

	err = cl.Connect(nil)
	if err != nil {
		t.Fatal(`Connect:`, err)
	}

	var (
		timeout = 3 * time.Second
		ticker  = time.NewTicker(timeout)

		expEvent Event
		gotEvent Event
		tag      string
		x        int
	)
	for x = range cases {
		tag = fmt.Sprintf(`expEvent #%d`, x)
		select {
		case <-ticker.C:
			t.Fatalf(`%s: timeout`, tag)

		case gotEvent = <-cl.C:
			expEvent = <-expq
			test.Assert(t, tag, expEvent, gotEvent)
		}
		ticker.Reset(timeout)
	}

	_ = cl.Close()

	test.Assert(t, `LastEventID`, cl.LastEventID, `1`)
}

// testGenerateAddress generate random port for server address.
func testGenerateAddress() (addr string) {
	var port = rand.Int() % 60000
	if port < 1024 {
		port += 1024
	}
	return fmt.Sprintf(`127.0.0.1:%d`, port)
}

func testRunSSEServer(t *testing.T, cb libhttp.SSECallback) (address string, err error) {
	address = testGenerateAddress()

	var (
		serverOpts = &libhttp.ServerOptions{
			Address: address,
		}
		srv *libhttp.Server
	)
	srv, err = libhttp.NewServer(serverOpts)
	if err != nil {
		return ``, err
	}

	var sse = &libhttp.SSEEndpoint{
		Path: `/sse`,
		Call: cb,
	}

	err = srv.RegisterSSE(sse)
	if err != nil {
		return ``, err
	}

	go srv.Start()

	err = libnet.WaitAlive(`tcp`, address, 1*time.Second)
	if err != nil {
		// Server may not run due to address already in use.
		t.Skip(err)
	}

	t.Cleanup(func() { srv.Stop(1 * time.Second) })

	return address, nil
}
