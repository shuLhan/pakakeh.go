// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sseclient

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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
		kind: EventTypeMessage,
		data: "line 1\n\nline2\n\n",
		id:   fnoid,
		exp: Event{
			Type: EventTypeMessage,
			Data: "line 1\n\nline2\n\n",
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

	var servercb = func(sseconn *libhttp.SSEConn) {
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
			default:
				// Named type.
				err = sseconn.WriteEvent(c.kind, c.data, c.id())
				if err != nil {
					t.Fatalf(`WriteEvent #%d: %s`, x, err)
				}
			}
			expq <- c.exp
		}
	}

	var (
		srv *libhttp.Server
		err error
	)

	srv, err = testRunSSEServer(t, servercb)
	if err != nil {
		t.Fatal(`testRunSSEServer:`, err)
	}
	t.Cleanup(func() {
		_ = srv.Stop(1 * time.Second)
	})

	var cl = Client{
		Endpoint: fmt.Sprintf(`http://%s/sse`, srv.Options.Address),
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

func TestClient_raw(t *testing.T) {
	type testCase struct {
		raw []byte
		exp []Event
	}

	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/write_raw_test.data`)
	if err != nil {
		t.Fatal(err)
	}

	var cases = []testCase{{
		raw: tdata.Input[`case/1`],
		exp: []Event{{
			Type: EventTypeOpen, // The first message always open.
		}, {
			Type: EventTypeMessage,
			Data: "YHOO\n+2\n10",
		}},
	}, {
		raw: tdata.Input[`case/2`],
		exp: []Event{{
			Type: EventTypeMessage,
			Data: `first event`,
			ID:   `1`,
		}, {
			Type: EventTypeMessage,
			Data: `second event`,
		}, {
			Type: EventTypeMessage,
			Data: ` third event`,
		}},
	}, {
		// The SSE specification have inconsistent state when
		// dispatching empty data.
		// In the "9.2.6 Interpreting an event stream", if the data
		// buffer is empty it would return; but in the example as
		// tested in this case it would dispatch empty string.
		raw: tdata.Input[`case/3`],
		exp: []Event{{
			Type: EventTypeMessage,
			Data: "\n",
		}},
	}, {
		raw: tdata.Input[`case/4`],
		exp: []Event{{
			Type: EventTypeMessage,
			Data: `test`,
		}, {
			Type: EventTypeMessage,
			Data: `test`,
		}},
	}, {
		raw: tdata.Input[`case/5`],
		exp: []Event{{
			Type: `join`,
			Data: "Named event\n with multiple\n  ID",
			ID:   `2`,
		}},
	}}

	var expq = make(chan Event)

	var servercb = func(sseconn *libhttp.SSEConn) {
		var (
			c    testCase
			ev   Event
			errw error
			x    int
		)
		for x, c = range cases {
			errw = sseconn.WriteRaw([]byte(c.raw))
			if errw != nil {
				t.Fatalf(`WriteRaw #%d: %s`, x, errw)
			}
			for _, ev = range c.exp {
				expq <- ev
			}
		}
	}

	var srv *libhttp.Server

	srv, err = testRunSSEServer(t, servercb)
	if err != nil {
		t.Fatal(`testRunSSEServer:`, err)
	}
	t.Cleanup(func() {
		_ = srv.Stop(1 * time.Second)
	})

	var cl = Client{
		Endpoint: fmt.Sprintf(`http://%s/sse`, srv.Options.Address),
	}

	err = cl.Connect(nil)
	if err != nil {
		t.Fatal(`Connect:`, err)
	}

	var (
		timeout = 1 * time.Second
		ticker  = time.NewTicker(timeout)

		c        testCase
		expEvent Event
		gotEvent Event
		tag      string
		x, y     int
	)
	for x, c = range cases {
		for y = range c.exp {
			tag = fmt.Sprintf(`Case #%d/#%d`, x, y)

			select {
			case <-ticker.C:
				t.Fatalf(`%s: timeout`, tag)

			case gotEvent = <-cl.C:
				expEvent = <-expq
				test.Assert(t, tag, expEvent, gotEvent)
			}
			ticker.Reset(timeout)
		}
	}
	_ = cl.Close()
	test.Assert(t, `LastEventID`, `2`, cl.LastEventID)
}

func TestClientRetry(t *testing.T) {
	const testKindClose = `close`

	type testCase struct {
		kind string
		raw  []byte
		exp  []Event
	}

	var cases = []testCase{{
		raw: []byte("retry: 100\n\n"),
		exp: []Event{{
			Type: EventTypeOpen, // The first message always open.
		}},
	}, {
		// This is where server close its handler.
		kind: testKindClose,
	}, {
		raw: []byte("data: after close\n\n"),
		exp: []Event{{
			Type: EventTypeOpen, // The first message after reconnect.
		}, {
			Type: EventTypeMessage,
			Data: `after close`,
		}},
	}}

	var expq = make(chan Event)

	// Counter for test case between open and close in server
	// callback.
	var casenum atomic.Int64

	var servercb = func(sseconn *libhttp.SSEConn) {
		var (
			x   = int(casenum.Load())
			c   testCase
			ev  Event
			err error
		)
		for x < len(cases) {
			c = cases[x]
			x++
			casenum.Store(int64(x))

			switch c.kind {
			case testKindClose:
				// Close the connection here, continue
				// later.
				return
			default:
				err = sseconn.WriteRaw([]byte(c.raw))
				if err != nil {
					t.Fatalf(`WriteRaw #%d: %s`, x, err)
				}
			}
			for _, ev = range c.exp {
				expq <- ev
			}
		}
	}

	var (
		srv *libhttp.Server
		err error
	)

	srv, err = testRunSSEServer(t, servercb)
	if err != nil {
		t.Fatal(`testRunSSEServer:`, err)
	}
	t.Cleanup(func() {
		_ = srv.Stop(1 * time.Second)
	})

	var cl = Client{
		Endpoint: fmt.Sprintf(`http://%s/sse`, srv.Options.Address),
	}
	err = cl.Connect(nil)
	if err != nil {
		t.Fatal(`Connect:`, err)
	}

	var (
		timeout = 1 * time.Second
		ticker  = time.NewTicker(timeout)

		c        testCase
		expEvent Event
		gotEvent Event
		tag      string
		x, y     int
	)
	for x, c = range cases {
		for y = range c.exp {
			tag = fmt.Sprintf(`Case #%d/#%d`, x, y)

			select {
			case <-ticker.C:
				t.Fatalf(`%s: timeout`, tag)

			case gotEvent = <-cl.C:
				expEvent = <-expq
				test.Assert(t, tag, expEvent, gotEvent)
			}
			ticker.Reset(timeout)
		}
	}
	_ = cl.Close()

}

// testGenerateAddress generate random port for server address.
func testGenerateAddress() (addr string) {
	var (
		logp = `testGenerateAddress`
		max  = big.NewInt(60000)

		randInt *big.Int
		err     error
	)

	randInt, err = rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatalf(`%s: %s`, logp, err)
	}

	var port = randInt.Int64()
	if port < 1024 {
		port += 1024
	}
	return fmt.Sprintf(`127.0.0.1:%d`, port)
}

func testRunSSEServer(t *testing.T, cb libhttp.SSECallback) (srv *libhttp.Server, err error) {
	var address = testGenerateAddress()

	var (
		serverOpts = libhttp.ServerOptions{
			Address: address,
		}
	)
	srv, err = libhttp.NewServer(serverOpts)
	if err != nil {
		return nil, err
	}

	var sse = libhttp.SSEEndpoint{
		Path: `/sse`,
		Call: cb,
	}

	err = srv.RegisterSSE(sse)
	if err != nil {
		return nil, err
	}

	go func() {
		var errStart = srv.Start()
		if errStart != nil {
			log.Fatal(errStart)
		}
	}()

	err = libnet.WaitAlive(`tcp`, address, 1*time.Second)
	if err != nil {
		// Server may not run due to address already in use.
		t.Skip(err)
	}

	return srv, nil
}
