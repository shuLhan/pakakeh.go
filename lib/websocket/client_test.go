// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"net/http"
	"sync"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

//
// TestConnect this test require a websocket server to be run.
//
func TestConnect(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	cases := []struct {
		desc     string
		endpoint string
		headers  http.Header
		expErr   string
	}{{
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

	for _, c := range cases {
		t.Log(c.desc)

		client := &Client{
			Endpoint: c.endpoint,
			Headers:  c.headers,
		}

		err := client.Connect()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		client.SendClose(StatusNormal, nil)
	}
}

func TestClientPing(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}
		wg sync.WaitGroup
	)

	err := testClient.Connect()
	if err != nil {
		t.Fatal("TestClientPing: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		req       []byte
		exp       *Frame
		expClose  *Frame
	}{{
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

	for _, c := range cases {
		c := c
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.Connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		testClient.handleClose = func(cl *Client, got *Frame) error {
			exp := c.expClose
			test.Assert(t, "close", exp, got, true)

			if len(got.payload) >= 2 {
				got.payload = got.payload[2:]
			}

			cl.SendClose(got.closeCode, got.payload)
			cl.Quit()
			wg.Done()
			return nil
		}

		testClient.handlePong = func(cl *Client, got *Frame) (err error) {
			exp := c.exp
			test.Assert(t, "handlePong", exp, got, true)
			wg.Done()
			return nil
		}

		wg.Add(1)
		err := testClient.send(c.req)
		if err != nil {
			t.Fatal(err)
		}

		wg.Wait()
	}

	testClient.Quit()
}

func TestClientText(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}
		wg sync.WaitGroup
	)

	err := testClient.Connect()
	if err != nil {
		t.Fatal("TestClientText: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		req       []byte
		exp       *Frame
		expClose  *Frame
	}{{
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

	for _, c := range cases {
		c := c
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.Connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		testClient.handleClose = func(cl *Client, got *Frame) error {
			exp := c.expClose
			test.Assert(t, "close", exp, got, true)
			cl.SendClose(got.closeCode, got.payload)
			cl.Quit()
			wg.Done()
			return nil
		}

		testClient.HandleText = func(cl *Client, got *Frame) error {
			exp := c.exp
			test.Assert(t, "text", exp, got, true)
			wg.Done()
			return nil
		}

		wg.Add(1)
		err := testClient.send(c.req)
		if err != nil {
			t.Fatal(err)
		}

		wg.Wait()
	}

	testClient.Quit()
}

func TestClientFragmentation(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}
		wg sync.WaitGroup
	)

	err := testClient.Connect()
	if err != nil {
		t.Fatal("TestClientFragmentation: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		frames    []Frame
		exp       *Frame
		expClose  *Frame
	}{{
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
		desc:      "Three text frames, unmasked",
		reconnect: true,
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
		desc:      "Three text frames, masked",
		reconnect: true,
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

	for _, c := range cases {
		c := c
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.Connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		testClient.handleClose = func(cl *Client, got *Frame) error {
			exp := c.expClose
			test.Assert(t, "close", exp, got, true)
			cl.SendClose(got.closeCode, got.payload)
			cl.Quit()
			wg.Done()
			return nil
		}

		testClient.HandleText = func(cl *Client, got *Frame) error {
			exp := c.exp
			test.Assert(t, "text", exp, got, true)
			wg.Done()
			return nil
		}

		wg.Add(1)
		for x := 0; x < len(c.frames); x++ {
			req := c.frames[x].pack()

			err := testClient.send(req)
			if err != nil {
				t.Fatal(err)
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
		wg sync.WaitGroup
	)

	err := testClient.Connect()
	if err != nil {
		t.Fatal("TestClientFragmentation2: " + err.Error())
	}

	frames := []Frame{{
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
		exp := &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePong,
			len:        4,
			payload:    []byte("PING"),
			isComplete: true,
		}
		test.Assert(t, "handlePong", exp, got, true)
		wg.Done()
		return nil
	}

	testClient.HandleText = func(cl *Client, got *Frame) error {
		exp := &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			len:        14,
			payload:    []byte("Hello, Shulhan"),
			isComplete: true,
		}
		test.Assert(t, "handlePong", exp, got, true)
		wg.Done()
		return nil
	}

	wg.Add(2)
	for x := 0; x < len(frames); x++ {
		req := frames[x].pack()

		err := testClient.send(req)
		if err != nil {
			t.Fatal(err)
		}
	}

	wg.Wait()
}

func TestClientSendBin(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}
		wg sync.WaitGroup
	)

	err := testClient.Connect()
	if err != nil {
		t.Fatal("TestSendBin: Connect: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		payload   []byte
		exp       *Frame
	}{{
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

	for _, c := range cases {
		c := c
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.Connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		testClient.HandleBin = func(cl *Client, got *Frame) error {
			exp := c.exp
			test.Assert(t, "HandleBin", exp, got, true)
			wg.Done()
			return nil
		}

		wg.Add(1)
		err := testClient.SendBin(c.payload)
		if err != nil {
			t.Fatal("TestSendBin: " + err.Error())
		}

		wg.Wait()
	}
}

func TestClientSendPing(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}
		wg sync.WaitGroup
	)

	err := testClient.Connect()
	if err != nil {
		t.Fatal("TestSendBin: Connect: " + err.Error())
	}

	cases := []struct {
		desc    string
		payload []byte
		exp     *Frame
	}{{
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

	for _, c := range cases {
		c := c
		t.Log(c.desc)

		testClient.handlePong = func(cl *Client, got *Frame) error {
			exp := c.exp
			test.Assert(t, "handlePong", exp, got, true)
			wg.Done()
			return nil
		}

		wg.Add(1)
		err := testClient.SendPing(c.payload)
		if err != nil {
			t.Fatal("TestSendPing: " + err.Error())
		}

		wg.Wait()
	}

	testClient.Quit()
}

func TestClientSendClose(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var (
		testClient = &Client{
			Endpoint: _testEndpointAuth,
		}
		wg sync.WaitGroup
	)

	err := testClient.Connect()
	if err != nil {
		t.Fatal("TestClientSendClose: Connect: " + err.Error())
	}

	testClient.handleClose = func(cl *Client, got *Frame) error {
		exp := &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeClose,
			closeCode:  StatusNormal,
			len:        8,
			payload:    []byte{0x03, 0xE8, 'n', 'o', 'r', 'm', 'a', 'l'},
			isComplete: true,
		}
		test.Assert(t, "handleClose", exp, got, true)
		cl.Quit()
		wg.Done()
		return nil
	}

	wg.Add(1)
	err = testClient.SendClose(StatusNormal, []byte("normal"))
	if err != nil {
		t.Fatal("TestClientSendClose: " + err.Error())
	}

	wg.Wait()

	err = testClient.SendPing(nil)

	test.Assert(t, "error", ErrConnClosed, err, true)
}
