// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"net/http"
	"testing"

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
		req:  concatBytes([]byte{0x81, 0x7E, 0x01, 0x00}, _dummyPayload256...),
		exp:  NewFrameClose(false, StatusBadRequest, nil),
	}, {
		desc:      "Medium payload 256, masked",
		reconnect: true,
		req: concatBytes([]byte{
			0x81, 0xFE, 0x01, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload256Masked...),
		exp: concatBytes([]byte{
			0x81, 0x7E, 0x01, 0x00,
		}, _dummyPayload256...),
	}, {
		desc: "Large payload 65536, unmasked",
		req: concatBytes([]byte{
			0x81, 0x7F,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}, _dummyPayload65536...),
		exp: NewFrameClose(false, StatusBadRequest, nil),
	}, {
		desc:      "Large payload 65536, masked",
		reconnect: true,
		req: concatBytes([]byte{
			0x81, 0xFF,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload65536Masked...),
		exp: concatBytes([]byte{
			0x81, 0x7F,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}, _dummyPayload65536...),
	}}

	recvHandler := func(ctx context.Context, resp []byte) (err error) {
		exp := ctx.Value(ctxKeyBytes).([]byte)
		test.Assert(t, "", exp, resp, true)
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

func TestClientFragmentation(t *testing.T) {
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
		frames    []Frame
		exps      [][]byte
	}{{
		desc: "Two text frames, unmasked",
		frames: []Frame{{
			Fin:     0,
			opcode:  opcodeText,
			Payload: []byte{'H', 'e', 'l'},
		}, {
			Fin:     frameIsFinished,
			opcode:  opcodeCont,
			Payload: []byte{'l', 'o'},
		}},
		exps: [][]byte{
			NewFrameClose(false, StatusBadRequest, nil),
		},
	}, {
		desc:      "Three text frames, unmasked",
		reconnect: true,
		frames: []Frame{{
			Fin:     0,
			opcode:  opcodeText,
			Payload: []byte("Hel"),
		}, {
			Fin:     0,
			opcode:  opcodeCont,
			Payload: []byte("lo, "),
		}, {
			Fin:     frameIsFinished,
			opcode:  opcodeCont,
			Payload: []byte("Shulhan"),
		}},
		exps: [][]byte{
			NewFrameClose(false, StatusBadRequest, nil),
		},
	}, {
		desc:      "Three text frames with control message in the middle",
		reconnect: true,
		frames: []Frame{{
			Fin:     0,
			opcode:  opcodeText,
			Masked:  frameIsMasked,
			Payload: []byte("Hel"),
		}, {
			Fin:     0,
			opcode:  opcodeCont,
			Masked:  frameIsMasked,
			Payload: []byte("lo, "),
		}, {
			Fin:     frameIsFinished,
			opcode:  opcodePing,
			Masked:  frameIsMasked,
			Payload: []byte("PING"),
		}, {
			Fin:     frameIsFinished,
			opcode:  opcodeCont,
			Masked:  frameIsMasked,
			Payload: []byte("Shulhan"),
		}},
		exps: [][]byte{
			{0x8A, 0x04, 'P', 'I', 'N', 'G'},
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
			req := c.frames[x].Pack(true)

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
		}
	}
}
