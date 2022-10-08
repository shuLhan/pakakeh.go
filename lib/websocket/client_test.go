// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

// TestConnect this test require a websocket server to be run.
func TestConnect(t *testing.T) {
	type testCase struct {
		headers  http.Header
		desc     string
		endpoint string
		expErr   string
	}

	if _testServer == nil {
		runTestServer()
	}

	var cases = []testCase{{
		desc:     "With custom header",
		endpoint: _testEndpointAuth,
		headers: http.Header{
			"Host":   []string{"myhost"},
			"Origin": []string{"localhost"},
		},
	}, {
		desc:     "Without credential",
		endpoint: _testWSAddr,
		expErr:   "websocket: Connect: 400 Missing authorization",
	}, {
		desc:     "With closed connection",
		endpoint: "ws://127.0.0.1:4444",
		expErr:   "websocket: Connect: dial tcp 127.0.0.1:4444: connect: connection refused",
	}}

	var (
		c      testCase
		client *Client
		err    error
	)

	for _, c = range cases {
		t.Log(c.desc)

		client = &Client{
			Endpoint: c.endpoint,
			Headers:  c.headers,
		}

		err = client.Connect()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		client.sendClose(StatusNormal, nil)
	}
}

func TestClient_parseURI(t *testing.T) {
	type testCase struct {
		expTLSConfig     *tls.Config
		endpoint         string
		expRemoteAddress string
		expError         string
	}

	var cases = []testCase{{
		endpoint:         "ws://127.0.0.1:8080",
		expRemoteAddress: "127.0.0.1:8080",
	}, {
		endpoint:         "wss://127.0.0.1",
		expRemoteAddress: "127.0.0.1:443",
		expTLSConfig:     new(tls.Config),
	}, {
		endpoint:         "wss://127.0.0.1:8000",
		expRemoteAddress: "127.0.0.1:8000",
		expTLSConfig:     new(tls.Config),
	}, {
		endpoint:         "http://127.0.0.1",
		expRemoteAddress: "127.0.0.1:80",
	}, {
		endpoint:         "https://127.0.0.1",
		expRemoteAddress: "127.0.0.1:443",
		expTLSConfig:     new(tls.Config),
	}, {
		endpoint:         "https://127.0.0.1:8443",
		expRemoteAddress: "127.0.0.1:8443",
		expTLSConfig:     new(tls.Config),
	}}

	var (
		cl = &Client{}

		c   testCase
		err error
	)

	for _, c = range cases {
		t.Log("parseURI", c.endpoint)

		cl.remoteAddr = ""
		cl.TLSConfig = nil
		cl.Endpoint = c.endpoint

		err = cl.parseURI()
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		test.Assert(t, "remote address", c.expRemoteAddress, cl.remoteAddr)
		test.Assert(t, "TLS config", c.expTLSConfig, cl.TLSConfig)
	}
}

func TestClientPing(t *testing.T) {
	type testCase struct {
		exp  *Frame
		desc string
		req  []byte
	}

	if _testServer == nil {
		runTestServer()
	}

	var cases = []testCase{{
		desc: `Without payload, unmasked`,
		req:  NewFramePing(false, nil),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc: `With payload, unmasked`,
		req:  NewFramePing(false, []byte(`Hello`)),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc: `With payload, masked`,
		req:  NewFramePing(true, []byte(`Hello`)),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePong,
			len:        5,
			payload:    []byte(`Hello`),
			isComplete: true,
		},
	}}

	var (
		gotFrame = make(chan *Frame)

		cl  *Client
		got *Frame
		c   testCase
		err error
	)

	for _, c = range cases {
		t.Log(c.desc)

		cl = &Client{
			Endpoint: _testEndpointAuth,
			handleClose: func(cl *Client, f *Frame) error {
				cl.sendClose(f.closeCode, nil)
				cl.Quit()
				gotFrame <- f
				return nil
			},
			handlePong: func(cl *Client, f *Frame) (err error) {
				gotFrame <- f
				return nil
			},
		}

		err = cl.Connect()
		if err != nil {
			t.Fatal(err)
		}

		cl.Lock()
		err = cl.send(c.req)
		cl.Unlock()
		if err != nil {
			t.Fatal(err)
		}

		got = <-gotFrame
		test.Assert(t, `response`, c.exp, got)

		if got.opcode != OpcodeClose {
			cl.Close()
		}
	}
}

func TestClient_send_FrameText(t *testing.T) {
	type testCase struct {
		exp  *Frame
		desc string
		req  []byte
	}

	if _testServer == nil {
		runTestServer()
	}

	var cases = []testCase{{
		desc: `Small payload, unmasked`,
		req:  NewFrameText(false, []byte(`Hello`)),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc: `Small payload, masked`,
		req:  NewFrameText(true, []byte(`Hello`)),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        5,
			payload:    []byte(`Hello`),
			isComplete: true,
		},
	}, {
		desc: `Medium payload 256, unmasked`,
		req:  NewFrameText(false, _dummyPayload256),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc: `Medium payload 256, masked`,
		req:  NewFrameText(true, _dummyPayload256),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        uint64(len(_dummyPayload256)),
			payload:    _dummyPayload256,
			isComplete: true,
		},
	}, {
		desc: `Large payload 65536, unmasked`,
		req:  NewFrameText(false, _dummyPayload65536),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc: `Large payload 65536, masked`,
		req:  NewFrameText(true, _dummyPayload65536),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        uint64(len(_dummyPayload65536)),
			payload:    _dummyPayload65536,
			isComplete: true,
		},
	}}

	var (
		gotFrame = make(chan *Frame)

		cl  *Client
		got *Frame
		c   testCase
		err error
	)

	for _, c = range cases {
		t.Log(c.desc)

		cl = &Client{
			Endpoint: _testEndpointAuth,
			handleClose: func(cl *Client, f *Frame) error {
				cl.sendClose(f.closeCode, nil)
				cl.Quit()
				gotFrame <- f
				return nil
			},
			HandleText: func(cl *Client, f *Frame) error {
				gotFrame <- f
				return nil
			},
		}

		err = cl.Connect()
		if err != nil {
			t.Fatal(err.Error())
		}

		cl.Lock()
		err = cl.send(c.req)
		if err != nil {
			t.Fatal(err)
		}
		cl.Unlock()

		got = <-gotFrame
		test.Assert(t, `response`, c.exp, got)

		if got.opcode != OpcodeClose {
			cl.Close()
		}
	}
}

func TestClientFragmentation(t *testing.T) {
	type testCase struct {
		exp    *Frame
		desc   string
		frames []Frame
	}

	if _testServer == nil {
		runTestServer()
	}

	var cases = []testCase{{
		desc: "Two text frames, unmasked",
		frames: []Frame{{
			fin:     0,
			opcode:  OpcodeText,
			payload: []byte{'H', 'e', 'l'},
		}, {
			fin:     frameIsFinished,
			opcode:  OpcodeCont,
			payload: []byte{'l', 'o'},
		}},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc: "Three text frames, unmasked",
		frames: []Frame{{
			fin:     0,
			opcode:  OpcodeText,
			payload: []byte("Hel"),
		}, {
			fin:     0,
			opcode:  OpcodeCont,
			payload: []byte("lo, "),
		}, {
			fin:     frameIsFinished,
			opcode:  OpcodeCont,
			payload: []byte("Shulhan"),
		}},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc: "Three text frames, masked",
		frames: []Frame{{
			fin:     0,
			opcode:  OpcodeText,
			masked:  frameIsMasked,
			payload: []byte("Hel"),
		}, {
			fin:     0,
			opcode:  OpcodeCont,
			masked:  frameIsMasked,
			payload: []byte("lo, "),
		}, {
			fin:     frameIsFinished,
			opcode:  OpcodeCont,
			masked:  frameIsMasked,
			payload: []byte("Shulhan"),
		}},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        14,
			payload:    []byte("Hello, Shulhan"),
			isComplete: true,
		},
	}}

	var (
		gotFrame = make(chan *Frame)

		cl    *Client
		frame Frame
		got   *Frame
		c     testCase
		err   error
		x     int
		req   []byte
	)

	for _, c = range cases {
		t.Log(c.desc)

		cl = &Client{
			Endpoint: _testEndpointAuth,
			handleClose: func(cl *Client, f *Frame) error {
				cl.sendClose(f.closeCode, nil)
				cl.Quit()
				gotFrame <- f
				return nil
			},
			HandleText: func(cl *Client, f *Frame) error {
				gotFrame <- f
				return nil
			},
		}

		err = cl.Connect()
		if err != nil {
			t.Fatal(err)
		}

		for x, frame = range c.frames {
			req = frame.pack()

			cl.Lock()
			err = cl.send(req)
			cl.Unlock()
			if err != nil {
				// If the client send unmasked frame,
				// server may close the connection before we
				// can test send the second frame.
				t.Logf(`send frame %d: %s`, x, err)
			}
		}

		got = <-gotFrame
		test.Assert(t, `response`, c.exp, got)

		if got.opcode != OpcodeClose {
			cl.Close()
		}
	}
}

// TestClientFragmentation2 We are sending two requests, first request split
// into 3 frames in between second request (PING):
//
//	F1->F2->PING->F3
func TestClientFragmentation2(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		gotFrame   = make(chan *Frame)
		testClient = &Client{
			Endpoint: _testEndpointAuth,
			handlePong: func(cl *Client, f *Frame) error {
				gotFrame <- f
				return nil
			},
			HandleText: func(cl *Client, f *Frame) error {
				gotFrame <- f
				return nil
			},
		}

		exp *Frame
		got *Frame
		err error
		x   int
		req []byte
	)

	err = testClient.Connect()
	if err != nil {
		t.Fatal("TestClientFragmentation2: " + err.Error())
	}

	var frames = []Frame{{
		fin:     0,
		opcode:  OpcodeText,
		masked:  frameIsMasked,
		payload: []byte("Hel"),
	}, {
		fin:     0,
		opcode:  OpcodeCont,
		masked:  frameIsMasked,
		payload: []byte("lo, "),
	}, {
		fin:     frameIsFinished,
		opcode:  OpcodePing,
		masked:  frameIsMasked,
		payload: []byte("PING"),
	}, {
		fin:     frameIsFinished,
		opcode:  OpcodeCont,
		masked:  frameIsMasked,
		payload: []byte("Shulhan"),
	}}

	for x = 0; x < len(frames); x++ {
		req = frames[x].pack()

		testClient.Lock()
		err = testClient.send(req)
		testClient.Unlock()
		if err != nil {
			t.Fatal(err)
		}
	}

	// The first response should be PONG.
	exp = &Frame{
		fin:        frameIsFinished,
		opcode:     OpcodePong,
		len:        4,
		payload:    []byte(`PING`),
		isComplete: true,
	}
	got = <-gotFrame
	test.Assert(t, `response PONG`, exp, got)

	exp = &Frame{
		fin:        frameIsFinished,
		opcode:     OpcodeText,
		len:        14,
		payload:    []byte(`Hello, Shulhan`),
		isComplete: true,
	}
	got = <-gotFrame
	test.Assert(t, `response TEXT`, exp, got)
}

func TestClientSendBin(t *testing.T) {
	type testCase struct {
		exp     *Frame
		desc    string
		payload []byte
	}

	if _testServer == nil {
		runTestServer()
	}

	var cases = []testCase{{
		desc:    "Single bin frame",
		payload: []byte("Hello"),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeBin,
			len:        5,
			payload:    []byte("Hello"),
			isComplete: true,
		},
	}}

	var (
		gotFrame = make(chan *Frame)

		cl  *Client
		got *Frame
		c   testCase
		err error
	)

	for _, c = range cases {
		t.Log(c.desc)

		cl = &Client{
			Endpoint: _testEndpointAuth,
			HandleBin: func(cl *Client, f *Frame) error {
				gotFrame <- f
				return nil
			},
		}

		err = cl.Connect()
		if err != nil {
			t.Fatal(err)
		}

		err = cl.SendBin(c.payload)
		if err != nil {
			t.Fatal(err.Error())
		}

		got = <-gotFrame
		test.Assert(t, `response`, c.exp, got)

		cl.Close()
	}
}

func TestClientSendPing(t *testing.T) {
	type testCase struct {
		exp     *Frame
		desc    string
		payload []byte
	}

	if _testServer == nil {
		runTestServer()
	}

	var cases = []testCase{{
		desc: "Without payload",
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePong,
			len:        0,
			isComplete: true,
		},
	}, {
		desc:    "With payload",
		payload: []byte("Test"),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePong,
			len:        4,
			payload:    []byte("Test"),
			isComplete: true,
		},
	}}

	var (
		gotFrame   = make(chan *Frame)
		testClient = &Client{
			Endpoint: _testEndpointAuth,
			handlePong: func(cl *Client, f *Frame) error {
				gotFrame <- f
				return nil
			},
		}

		got *Frame
		err error
		c   testCase
	)

	err = testClient.Connect()
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, c = range cases {
		t.Log(c.desc)

		err = testClient.SendPing(c.payload)
		if err != nil {
			t.Fatal(err.Error())
		}

		got = <-gotFrame
		test.Assert(t, `response`, c.exp, got)
	}
}

func TestClient_sendClose(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		gotFrame = make(chan *Frame)
		cl       = &Client{
			Endpoint: _testEndpointAuth,
			handleClose: func(cl *Client, f *Frame) error {
				cl.sendClose(f.closeCode, nil)
				cl.Quit()
				gotFrame <- f
				return nil
			},
		}

		got *Frame
		err error
	)

	err = cl.Connect()
	if err != nil {
		t.Fatal("TestClient_sendClose: Connect: " + err.Error())
	}

	err = cl.sendClose(StatusNormal, []byte("normal"))
	if err != nil {
		t.Fatal("TestClient_sendClose: " + err.Error())
	}

	got = <-gotFrame
	var exp = &Frame{
		fin:        frameIsFinished,
		opcode:     OpcodeClose,
		closeCode:  StatusNormal,
		len:        8,
		payload:    []byte{0x03, 0xE8, 'n', 'o', 'r', 'm', 'a', 'l'},
		isComplete: true,
	}
	test.Assert(t, `sendClose response`, exp, got)

	err = cl.SendPing(nil)
	test.Assert(t, `SendPing should error`, ErrConnClosed, err)
}
