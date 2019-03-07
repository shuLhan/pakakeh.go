// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
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
		endpoint: _testWSAddr + "?" + _qKeyTicket + "=" + _testExternalJWT,
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

		_, err := NewClient(c.endpoint, c.headers)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}
	}
}

func TestClientPing(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	endpoint := _testWSAddr + "?" + _qKeyTicket + "=" + _testExternalJWT

	testClient, err := NewClient(endpoint, nil)
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
		if frames.IsClosed() {
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
}

func TestClientText(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	endpoint := _testWSAddr + "?" + _qKeyTicket + "=" + _testExternalJWT

	testClient, err := NewClient(endpoint, nil)
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
		req: []byte{
			0x81, 0x05,
			'H', 'e', 'l', 'l', 'o',
		},
		exp: NewFrameClose(false, StatusBadRequest, nil),
	}, {
		desc:      "Small payload, masked",
		reconnect: true,
		req: []byte{
			0x81, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
		exp: NewFrameText(false, []byte("Hello")),
	}, {
		desc: "Medium payload 256, unmasked",
		req:  libbytes.Concat([]byte{0x81, 0x7E, 0x01, 0x00}, _dummyPayload256),
		exp:  NewFrameClose(false, StatusBadRequest, nil),
	}, {
		desc:      "Medium payload 256, masked",
		reconnect: true,
		req: libbytes.Concat([]byte{
			0x81, 0xFE, 0x01, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload256Masked),
		exp: libbytes.Concat([]byte{
			0x81, 0x7E, 0x01, 0x00,
		}, _dummyPayload256),
	}, {
		desc: "Large payload 65536, unmasked",
		req: libbytes.Concat([]byte{
			0x81, 0x7F,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}, _dummyPayload65536),
		exp: NewFrameClose(false, StatusBadRequest, nil),
	}, {
		desc:      "Large payload 65536, masked",
		reconnect: true,
		req: libbytes.Concat([]byte{
			0x81, 0xFF,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload65536Masked),
		exp: libbytes.Concat([]byte{
			0x81, 0x7F,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}, _dummyPayload65536),
	}}

	recvHandler := func(ctx context.Context, resp []byte) (err error) {
		exp := ctx.Value(ctxKeyBytes).([]byte)

		test.Assert(t, "", exp, resp, true)

		frames := Unpack(resp)
		if frames.IsClosed() {
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

		ctx := context.WithValue(context.Background(), ctxKeyBytes, c.exp)
		err := testClient.send(ctx, c.req, recvHandler)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestClientFragmentation(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	endpoint := _testWSAddr + "?" + _qKeyTicket + "=" + _testExternalJWT

	testClient, err := NewClient(endpoint, nil)
	if err != nil {
		t.Fatal("TestClientFragmentation: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		frames    []Frame
		exps      [][]byte
	}{{
		desc: "Two text frames, unmasked",
		frames: []Frame{{
			fin:     0,
			opcode:  opcodeText,
			payload: []byte{'H', 'e', 'l'},
		}, {
			fin:     frameIsFinished,
			opcode:  opcodeCont,
			payload: []byte{'l', 'o'},
		}},
		exps: [][]byte{
			NewFrameClose(false, StatusBadRequest, nil),
		},
	}, {
		desc:      "Three text frames, unmasked",
		reconnect: true,
		frames: []Frame{{
			fin:     0,
			opcode:  opcodeText,
			payload: []byte("Hel"),
		}, {
			fin:     0,
			opcode:  opcodeCont,
			payload: []byte("lo, "),
		}, {
			fin:     frameIsFinished,
			opcode:  opcodeCont,
			payload: []byte("Shulhan"),
		}},
		exps: [][]byte{
			NewFrameClose(false, StatusBadRequest, nil),
		},
	}, {
		desc:      "Three text frames with control message in the middle",
		reconnect: true,
		frames: []Frame{{
			fin:     0,
			opcode:  opcodeText,
			masked:  frameIsMasked,
			payload: []byte("Hel"),
		}, {
			fin:     0,
			opcode:  opcodeCont,
			masked:  frameIsMasked,
			payload: []byte("lo, "),
		}, {
			fin:     frameIsFinished,
			opcode:  opcodePing,
			masked:  frameIsMasked,
			payload: []byte("PING"),
		}, {
			fin:     frameIsFinished,
			opcode:  opcodeCont,
			masked:  frameIsMasked,
			payload: []byte("Shulhan"),
		}},
		exps: [][]byte{
			{0x8A, 0x04, 'P', 'I', 'N', 'G'}, // PONG with payload PING.
			{
				0x81, 0x0E,
				'H', 'e', 'l', 'l', 'o', ',', ' ',
				'S', 'h', 'u', 'l', 'h', 'a', 'n',
			},
		},
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
			req := c.frames[x].pack(true)

			err := testClient.send(context.Background(), req, nil)
			if err != nil {
				t.Fatal(err)
			}
		}

		for x := 0; x < len(c.exps); x++ {
			res, err := testClient.recv()
			if err != nil {
				t.Fatal(err)
			}

			test.Assert(t, "res", c.exps[x], res, true)

			frames := Unpack(res)
			if frames.IsClosed() {
				testClient.SendClose(false)
				break
			}
		}
	}
}

func TestClientSendBin(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	endpoint := _testWSAddr + "?" + _qKeyTicket + "=" + _testExternalJWT

	testClient, err := NewClient(endpoint, nil)
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
			opcode:  opcodeBin,
			len:     5,
			payload: []byte("Hello"),
		},
	}}

	checkBinResponse := func(ctx context.Context, frames *Frames) error {
		exp := ctx.Value(ctxKeyFrame).(*Frame)

		test.Assert(t, "SendBin response", exp, frames.v[0], true)

		if frames.IsClosed() {
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
}

func TestClientSendPing(t *testing.T) {
	if _testServer == nil {
		runTestServer()
	}

	endpoint := _testWSAddr + "?" + _qKeyTicket + "=" + _testExternalJWT

	testClient, err := NewClient(endpoint, nil)
	if err != nil {
		t.Fatal("TestSendBin: NewClient: " + err.Error())
	}

	cases := []struct {
		desc      string
		reconnect bool
		payload   []byte
		exp       *Frame
	}{{
		desc: "Without payload",
		exp: &Frame{
			fin:    frameIsFinished,
			opcode: opcodePong,
			len:    0,
		},
	}, {
		desc:    "With payload",
		payload: []byte("Test"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodePong,
			len:     4,
			payload: []byte("Test"),
		},
	}}

	handlePing := func(ctx context.Context, packet []byte) error {
		frames := Unpack(packet)

		exp := ctx.Value(ctxKeyFrame).(*Frame)

		test.Assert(t, "SendPing response", exp, frames.v[0], true)

		if frames.IsClosed() {
			testClient.SendClose(false)
		}

		return nil
	}

	testClient.handlePing = handlePing

	for _, c := range cases {
		t.Log(c.desc)

		if c.reconnect {
			err := testClient.connect()
			if err != nil {
				t.Fatal(err)
			}
		}

		ctx := context.WithValue(context.Background(), ctxKeyFrame, c.exp)

		err := testClient.SendPing(ctx, c.payload)
		if err != nil {
			t.Fatal("TestSendPing: " + err.Error())
		}
	}
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
	// When client accepted by server, send ping immediately and expect to
	// receive PONG response.
	//
	_testServer.HandleClientAdd = func(ctx context.Context, conn int) {
		framePing := NewFramePing(false, expPayload)
		err := Send(conn, framePing)
		if err != nil {
			cleanupServePing()
			t.Fatal("TestClientServePing: handleClientAdd: Send: " + err.Error())
		}
	}

	_testServer.handlePong = func(conn int, frame *Frame) {
		cleanupServePing()
		test.Assert(t, "TestClientServePing", expPayload, frame.payload, true)
		wg.Done()
	}

	endpoint := _testWSAddr + "?" + _qKeyTicket + "=" + _testExternalJWT

	testClient, err := NewClient(endpoint, nil)
	if err != nil {
		cleanupServePing()
		t.Fatal("TestClientServePing: NewClient: " + err.Error())
	}

	packet, err := testClient.recv()
	if err != nil {
		cleanupServePing()
		t.Fatal("TestClientServePing: Recv: " + err.Error())

	}

	frame, _ := frameUnpack(packet)
	test.Assert(t, "Client receive", expPayload, frame.payload, true)

	wg.Add(1)
	testClient.pingQueue <- frame
	wg.Done()

	cleanupServePing()
}
