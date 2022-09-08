// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"crypto/tls"
	"net/http"
	"strings"
	"sync"
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
		exp       *Frame
		expClose  *Frame
		desc      string
		req       []byte
		reconnect bool
	}

	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}

		wg  sync.WaitGroup
		err error
	)

	err = testClient.Connect()
	if err != nil {
		t.Fatal("TestClientPing: " + err.Error())
	}

	var cases = []testCase{{
		desc: "Without payload, unmasked",
		req:  NewFramePing(false, nil),
		expClose: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc:      "With payload, unmasked",
		reconnect: true,
		req:       NewFramePing(false, []byte("Hello")),
		expClose: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc:      "With payload, masked",
		reconnect: true,
		req:       NewFramePing(true, []byte("Hello")),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePong,
			len:        5,
			payload:    []byte("Hello"),
			isComplete: true,
		},
	}}

	var (
		c testCase
	)

	for _, c = range cases {
		var c testCase = c
		t.Log(c.desc)

		if c.reconnect {
			err = testClient.Connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		testClient.handleClose = func(cl *Client, got *Frame) error {
			var exp *Frame = c.expClose

			test.Assert(t, "close", exp, got)

			if len(got.payload) >= 2 {
				got.payload = got.payload[2:]
			}

			cl.sendClose(got.closeCode, got.payload)
			cl.Quit()
			wg.Done()
			return nil
		}

		testClient.handlePong = func(cl *Client, got *Frame) (err error) {
			var exp *Frame = c.exp

			test.Assert(t, "handlePong", exp, got)

			wg.Done()
			return nil
		}

		wg.Add(1)
		testClient.Lock()
		err = testClient.send(c.req)
		testClient.Unlock()
		if err != nil {
			t.Fatal(err)
		}

		wg.Wait()
	}

	testClient.Quit()
}

func TestClientText(t *testing.T) {
	type testCase struct {
		exp       *Frame
		expClose  *Frame
		desc      string
		req       []byte
		reconnect bool
	}

	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}

		wg  sync.WaitGroup
		err error
	)

	err = testClient.Connect()
	if err != nil {
		t.Fatal("TestClientText: " + err.Error())
	}

	var cases = []testCase{{
		desc: "Small payload, unmasked",
		req:  NewFrameText(false, []byte("Hello")),
		expClose: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc:      "Small payload, masked",
		reconnect: true,
		req:       NewFrameText(true, []byte("Hello")),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        5,
			payload:    []byte("Hello"),
			isComplete: true,
		},
	}, {
		desc: "Medium payload 256, unmasked",
		req:  NewFrameText(false, _dummyPayload256),
		expClose: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc:      "Medium payload 256, masked",
		reconnect: true,
		req:       NewFrameText(true, _dummyPayload256),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        uint64(len(_dummyPayload256)),
			payload:    _dummyPayload256,
			isComplete: true,
		},
	}, {
		desc: "Large payload 65536, unmasked",
		req:  NewFrameText(false, _dummyPayload65536),
		expClose: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusBadRequest,
			len:        2,
			payload:    []byte{0x03, 0xEA},
			isComplete: true,
		},
	}, {
		desc:      "Large payload 65536, masked",
		reconnect: true,
		req:       NewFrameText(true, _dummyPayload65536),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        uint64(len(_dummyPayload65536)),
			payload:    _dummyPayload65536,
			isComplete: true,
		},
	}}

	var (
		c testCase
	)

	for _, c = range cases {
		var c testCase = c
		t.Log(c.desc)

		if c.reconnect {
			err = testClient.Connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		testClient.handleClose = func(cl *Client, got *Frame) error {
			var exp *Frame = c.expClose
			test.Assert(t, "close", exp, got)
			cl.sendClose(got.closeCode, got.payload)
			cl.Quit()
			wg.Done()
			return nil
		}

		testClient.HandleText = func(cl *Client, got *Frame) error {
			var exp *Frame = c.exp
			test.Assert(t, "text", exp, got)
			wg.Done()
			return nil
		}

		wg.Add(1)
		err = testClient.send(c.req)
		if err != nil {
			t.Fatal(err)
		}

		wg.Wait()
	}
}

func TestClientFragmentation(t *testing.T) {
	type testCase struct {
		exp      *Frame
		expClose *Frame
		desc     string
		frames   []Frame
	}

	if _testServer == nil {
		runTestServer()
	}

	var (
		wg sync.WaitGroup
	)

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
		expClose: &Frame{
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
		expClose: &Frame{
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
		c          testCase
		testClient *Client
		err        error
		req        []byte
		x          int
		brokenPipe bool
	)

	for _, c = range cases {
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}

		err = testClient.Connect()
		if err != nil {
			t.Fatal(err)
		}

		testClient.handleClose = func(desc string, exp *Frame) ClientHandler {
			return func(cl *Client, got *Frame) (err error) {
				test.Assert(t, desc+": close", exp, got)
				cl.sendClose(got.closeCode, got.payload)
				cl.Quit()
				wg.Done()
				return nil
			}
		}(c.desc, c.expClose)

		testClient.HandleText = func(desc string, exp *Frame) ClientHandler {
			return func(cl *Client, got *Frame) error {
				test.Assert(t, desc+": text", exp, got)
				wg.Done()
				return nil
			}
		}(c.desc, c.exp)

		wg.Add(1)
		for x = 0; x < len(c.frames); x++ {
			req = c.frames[x].pack()

			testClient.Lock()
			err = testClient.send(req)
			testClient.Unlock()
			if err != nil {
				// If the client send unmasked frame, the
				// server may close the connection before we
				// can test send the second frame.
				brokenPipe = strings.Contains(err.Error(), "write: broken pipe")
				if !brokenPipe {
					t.Fatalf("expecting broken pipe, got %s", err)
				}
				break
			}
		}
		wg.Wait()
	}
}

func TestClientFragmentation2(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}

		wg  sync.WaitGroup
		err error
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

	testClient.handlePong = func(cl *Client, got *Frame) error {
		var exp = &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePong,
			len:        4,
			payload:    []byte("PING"),
			isComplete: true,
		}
		test.Assert(t, "handlePong", exp, got)
		wg.Done()
		return nil
	}

	testClient.HandleText = func(cl *Client, got *Frame) error {
		var exp = &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        14,
			payload:    []byte("Hello, Shulhan"),
			isComplete: true,
		}
		test.Assert(t, "handlePong", exp, got)
		wg.Done()
		return nil
	}

	wg.Add(2)

	var (
		x   int
		req []byte
	)

	for x = 0; x < len(frames); x++ {
		req = frames[x].pack()

		testClient.Lock()
		err = testClient.send(req)
		testClient.Unlock()
		if err != nil {
			t.Fatal(err)
		}
	}

	wg.Wait()
}

func TestClientSendBin(t *testing.T) {
	type testCase struct {
		exp       *Frame
		desc      string
		payload   []byte
		reconnect bool
	}

	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}

		c   testCase
		wg  sync.WaitGroup
		err error
	)

	err = testClient.Connect()
	if err != nil {
		t.Fatal("TestSendBin: Connect: " + err.Error())
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

	for _, c = range cases {
		var cc testCase = c

		t.Log(cc.desc)

		if cc.reconnect {
			err = testClient.Connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		testClient.HandleBin = func(cl *Client, got *Frame) error {
			var exp *Frame = cc.exp
			test.Assert(t, "HandleBin", exp, got)
			wg.Done()
			return nil
		}

		wg.Add(1)
		err = testClient.SendBin(cc.payload)
		if err != nil {
			t.Fatal("TestSendBin: " + err.Error())
		}

		wg.Wait()
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

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}

		wg  sync.WaitGroup
		err error
	)

	err = testClient.Connect()
	if err != nil {
		t.Fatal("TestSendBin: Connect: " + err.Error())
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
		c testCase
	)

	for _, c = range cases {
		var cc testCase = c
		t.Log(cc.desc)

		testClient.handlePong = func(cl *Client, got *Frame) error {
			var exp *Frame = cc.exp
			test.Assert(t, "handlePong", exp, got)
			wg.Done()
			return nil
		}

		wg.Add(1)
		err = testClient.SendPing(cc.payload)
		if err != nil {
			t.Fatal("TestSendPing: " + err.Error())
		}

		wg.Wait()
	}
}

func TestClient_sendClose(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}
		wg sync.WaitGroup

		err error
	)

	err = testClient.Connect()
	if err != nil {
		t.Fatal("TestClient_sendClose: Connect: " + err.Error())
	}

	testClient.handleClose = func(cl *Client, got *Frame) error {
		var exp = &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusNormal,
			len:        8,
			payload:    []byte{0x03, 0xE8, 'n', 'o', 'r', 'm', 'a', 'l'},
			isComplete: true,
		}
		test.Assert(t, "handleClose", exp, got)
		cl.Quit()
		wg.Done()
		return nil
	}

	wg.Add(1)
	err = testClient.sendClose(StatusNormal, []byte("normal"))
	if err != nil {
		t.Fatal("TestClient_sendClose: " + err.Error())
	}

	wg.Wait()

	err = testClient.SendPing(nil)

	test.Assert(t, "error", ErrConnClosed, err)
}
