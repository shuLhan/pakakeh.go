// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"testing"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/test"
)

//
// TestNewClient this test require a websocket server to be run.
//
func TestNewClient(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	cases := []struct {
		desc     string
		endpoint string
		headers  http.Header
		expErr   string
	}{{
		desc:   "With empty endpoint",
		expErr: "websocket: NewClient: parse : empty url",
	}, {
		desc:     "With custom header",
		endpoint: _testEndpointAuth,
		headers: http.Header{
			"Host":   []string{"myhost"},
			"Origin": []string{"localhost"},
		},
	}, {
		desc:     "Without credential",
		endpoint: _testWSAddr,
		expErr:   "websocket: NewClient: 400 Missing authorization",
	}, {
		desc:     "With closed connection",
		endpoint: "ws://127.0.0.1:4444",
		expErr:   "websocket: NewClient: dial tcp 127.0.0.1:4444: connect: connection refused",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		testClient, err := NewClient(c.endpoint, c.headers)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		testClient.SendClose(true)
	}
}

func TestClientPing(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	testClient, err := NewClient(_testEndpointAuth, nil)
	if err != nil {
		t.Fatal("TestClientPing: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		req       []byte
		exp       []byte
	}{{
		desc: "Without payload, unmasked",
		req:  NewFramePing(false, nil),
		exp:  NewFrameClose(false, StatusBadRequest, nil),
	}, {
		desc:      "With payload, unmasked",
		reconnect: true,
		req:       []byte{0x89, 0x05, 'H', 'e', 'l', 'l', 'o'},
		exp:       NewFrameClose(false, StatusBadRequest, nil),
	}, {
		desc:      "With payload, masked",
		reconnect: true,
		req: []byte{
			0x89, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
		exp: NewFramePong(false, []byte("Hello")),
	}}

	recvHandler := func(ctx context.Context, resp []byte) (err error) {
		exp := ctx.Value(ctxKeyBytes).([]byte)

		test.Assert(t, "resp", exp, resp, true)

		frames := Unpack(resp)
		if frames.isClosed() {
			testClient.SendClose(false)
		}

		return
	}

	for _, c := range cases {
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.WithValue(context.Background(), ctxKeyBytes, c.exp)
		err := testClient.send(ctx, c.req, recvHandler)
		if err != nil {
			t.Fatal(err)
		}
	}

	testClient.Quit()
}

func TestClientText(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	testClient, err := NewClient(_testEndpointAuth, nil)
	if err != nil {
		t.Fatal("TestClientText: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		req       []byte
		exp       []byte
	}{{
		desc: "Small payload, unmasked",
		req:  NewFrameText(false, []byte("Hello")),
		exp:  NewFrameClose(false, StatusBadRequest, nil),
	}, {
		desc:      "Small payload, masked",
		reconnect: true,
		req:       NewFrameText(true, []byte("Hello")),
		exp:       []byte("Hello"),
	}, {
		desc: "Medium payload 256, unmasked",
		req:  NewFrameText(false, _dummyPayload256),
	}, {
		desc:      "Medium payload 256, masked",
		reconnect: true,
		req:       NewFrameText(true, _dummyPayload256),
		exp:       _dummyPayload256,
	}, {
		desc: "Large payload 65536, unmasked",
		req:  NewFrameText(false, _dummyPayload65536),
	}, {
		desc:      "Large payload 65536, masked",
		reconnect: true,
		req:       NewFrameText(true, _dummyPayload65536),
		exp:       _dummyPayload65536,
	}}

	recvHandler := func(ctx context.Context, resp []byte) (err error) {
		exp := ctx.Value(ctxKeyBytes).([]byte)

		if len(exp) != len(resp) {
			t.Logf("recvHandler first 4 bytes: % x\n", resp[:4])
			t.Logf("recvHandler last  4 bytes: % x\n", resp[len(resp)-4:])
		}

		frames := Unpack(resp)
		if frames.isClosed() {
			t.Log("sending close...")
			testClient.SendClose(false)
		} else {
			got := frames.payload()
			test.Assert(t, "TestClientText len", len(exp), len(got), true)
			test.Assert(t, "TestClientText", exp, got, true)
		}

		return nil
	}

	for _, c := range cases {
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.WithValue(context.Background(), ctxKeyBytes, c.exp)
		err := testClient.send(ctx, c.req, recvHandler)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Log("Quit ...")
	testClient.Quit()
}

func TestClientFragmentation(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	testClient, err := NewClient(_testEndpointAuth, nil)
	if err != nil {
		t.Fatal("TestClientFragmentation: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		frames    []Frame
		exp       []byte
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
		exp: NewFrameClose(false, StatusBadRequest, nil),
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
		exp: NewFrameClose(false, StatusBadRequest, nil),
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
		exp: NewFrameText(false, []byte("Hello, Shulhan")),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		for x := 0; x < len(c.frames); x++ {
			req := c.frames[x].Pack(true)

			err := testClient.send(context.Background(), req, nil)
			if err != nil {
				t.Fatal(err)
			}
		}

		res, err := testClient.recv()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "res", c.exp, res, true)

		frames := Unpack(res)
		if frames.isClosed() {
			testClient.SendClose(false)
			break
		}
	}

	testClient.Quit()
}

func TestClientFragmentation2(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	testClient, err := NewClient(_testEndpointAuth, nil)
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

	exps := [][]byte{{
		0x8A, 0x04, 'P', 'I', 'N', 'G',
	}, {
		0x81, 0x0E, 'H', 'e', 'l', 'l', 'o', ',', ' ',
		'S', 'h', 'u', 'l', 'h', 'a', 'n',
	}}

	// Server may send PONG and data frame in one packet.
	expMulti := [][]byte{
		libbytes.Concat(exps[0], exps[1]),
		libbytes.Concat(exps[1], exps[0]),
	}

	for x := 0; x < len(frames); x++ {
		req := frames[x].Pack(true)

		err := testClient.send(context.Background(), req, nil)
		if err != nil {
			t.Fatal(err)
		}
	}

	got, err := testClient.recv()
	if err != nil {
		t.Fatal("TestClientFragmentation2: recv: " + err.Error())
	}

	var foundMulti bool
	for _, exp := range expMulti {
		if bytes.Equal(exp, got) {
			foundMulti = true
			break
		}
	}
	if foundMulti {
		return
	}
	var foundPong, foundText bool
	switch {
	case bytes.Equal(exps[0], got):
		foundPong = true
	case bytes.Equal(exps[1], got):
		foundText = true
	}

	got, err = testClient.recv()
	if err != nil {
		t.Fatal("TestClientFragmentation2: recv: " + err.Error())
	}
	switch {
	case bytes.Equal(exps[0], got):
		foundPong = true
	case bytes.Equal(exps[1], got):
		foundText = true
	}

	if foundPong && foundText {
		return
	}

	if foundPong {
		t.Fatal("No text response received")
	} else {
		t.Fatal("No pong response received")
	}

	testClient.SendClose(true)
}

func TestClientSendBin(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	testClient, err := NewClient(_testEndpointAuth, nil)
	if err != nil {
		t.Fatal("TestSendBin: NewClient: " + err.Error())
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
			fin:     frameIsFinished,
			opcode:  OpcodeBin,
			len:     5,
			payload: []byte("Hello"),
		},
	}}

	checkBinResponse := func(ctx context.Context, frames *Frames) error {
		exp := ctx.Value(ctxKeyFrame).(*Frame)

		test.Assert(t, "SendBin response", exp, frames.v[0], true)

		if frames.isClosed() {
			testClient.SendClose(false)
		}

		return nil
	}

	for _, c := range cases {
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.WithValue(context.Background(), ctxKeyFrame, c.exp)

		err := testClient.SendBin(ctx, c.payload, checkBinResponse)
		if err != nil {
			t.Fatal("TestSendBin: " + err.Error())
		}
	}

	testClient.Quit()
}

func TestClientSendPing(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	testClient, err := NewClient(_testEndpointAuth, nil)
	if err != nil {
		t.Fatal("TestSendBin: NewClient: " + err.Error())
	}

	testHandlePing := func(ctx context.Context, packet []byte) error {
		frames := Unpack(packet)

		exp := ctx.Value(ctxKeyFrame).(*Frame)

		test.Assert(t, "SendPing response", exp, frames.v[0], true)

		if frames.isClosed() {
			t.Log("TestClientSendPing closing ...")
			testClient.SendClose(false)
		}

		return nil
	}

	cases := []struct {
		desc      string
		reconnect bool
		handler   clientRawHandler
		payload   []byte
		exp       *Frame
	}{{
		desc:    "Without payload",
		handler: testHandlePing,
		exp: &Frame{
			fin:    frameIsFinished,
			opcode: OpcodePong,
			len:    0,
		},
	}, {
		desc:    "With payload",
		handler: testHandlePing,
		payload: []byte("Test"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodePong,
			len:     4,
			payload: []byte("Test"),
		},
	}, {
		desc:    "With default handler",
		handler: handlePing,
		payload: []byte("Test"),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		testClient.handlePing = c.handler

		ctx := context.WithValue(context.Background(), ctxKeyFrame, c.exp)

		err := testClient.SendPing(ctx, c.payload)
		if err != nil {
			t.Fatal("TestSendPing: " + err.Error())
		}
	}

	testClient.Quit()
}

func cleanupServePing() {
	_testServer.HandleClientAdd = nil
	_testServer.handlePong = nil
}

func TestClientServePing(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	var wg sync.WaitGroup
	expPayload := []byte("ping from server")

	//
	// When client accepted by server, send control PING immediately and
	// expect to receive PONG response.
	//
	_testServer.HandleClientAdd = func(ctx context.Context, conn int) {
		framePing := NewFramePing(false, expPayload)
		for x := 0; x < 3; x++ {
			err := Send(conn, framePing)
			if err != nil {
				t.Fatal("TestClientServePing: HandleClientAdd: Send: " + err.Error())
			}
		}
		t.Logf("TestClientServePing: HandleClientAdd: PING\n")
	}

	_testServer.handlePong = func(conn int, frame *Frame) {
		t.Logf("TestClientServePing: handlePong: % x\n", frame.payload)
		test.Assert(t, "TestClientServePing", expPayload, frame.payload, true)
		wg.Done()
	}

	defer cleanupServePing()

	testClient, err := NewClient(_testEndpointAuth, nil)
	if err != nil {
		t.Fatal("TestClientServePing: NewClient: " + err.Error())
	}

	// Read PING from server.HandleClientAdd.
	var frames *Frames
	for frames == nil {
		packet, err := testClient.recv()
		if err != nil {
			t.Fatal("TestClientServePing: Recv: " + err.Error())
		}

		frames = Unpack(packet)
	}

	test.Assert(t, "Client receive", expPayload, frames.v[0].payload, true)

	wg.Add(1)
	testClient.pingQueue <- frames.v[0]
	wg.Wait()

	testClient.Quit()
}

func TestClientSendClose(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	testClient, err := NewClient(_testEndpointAuth, nil)
	if err != nil {
		t.Fatal("TestClientSendClose: NewClient: " + err.Error())
	}

	err = testClient.SendClose(true)
	if err != nil {
		t.Fatal("TestClientSendClose: " + err.Error())
	}

	test.Assert(t, "client.conn", nil, testClient.conn, true)

	err = testClient.SendPing(context.Background(), nil)

	test.Assert(t, "error", errConnClosed, err, true)
}

func TestClientQuit(t *testing.T) {
	var wg sync.WaitGroup

	if _testServer == nil {
		runTestServer()
	}

	_testServer.HandleClientRemove = func(ctx context.Context, conn int) {
		gotUID := ctx.Value(CtxKeyUID).(uint64)
		test.Assert(t, "context uid", _testUID, gotUID, true)
		wg.Done()
	}

	defer func() {
		_testServer.HandleClientRemove = nil
	}()

	testClient, err := NewClient(_testEndpointAuth, nil)
	if err != nil {
		t.Fatal("TestClientSendClose: NewClient: " + err.Error())
	}

	wg.Add(1)
	testClient.Quit()

	test.Assert(t, "client.conn", nil, testClient.conn, true)

	err = testClient.SendPing(context.Background(), nil)

	test.Assert(t, "error", errConnClosed, err, true)

	wg.Wait()
}
