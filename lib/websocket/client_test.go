// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"context"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var _wsClient *Client // nolint: gochecknoglobals

func testClientPing(t *testing.T) {
	cases := []struct {
		desc      string
		reconnect bool
		req       []byte
		exp       []byte
	}{{
		desc: "Without payload, unmasked",
		req:  ControlFramePing,
		exp:  concatBytes(ControlFrameCloseWithCode, StatusBadRequest...),
	}, {
		desc:      "With payload, unmasked",
		reconnect: true,
		req:       []byte{0x89, 0x05, 'H', 'e', 'l', 'l', 'o'},
		exp:       concatBytes(ControlFrameCloseWithCode, StatusBadRequest...),
	}, {
		desc:      "With payload, masked",
		reconnect: true,
		req: []byte{
			0x89, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
		exp: []byte{
			0x8A, 0x05,
			'H', 'e', 'l', 'l', 'o',
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		if c.reconnect {
			err := _wsClient.Reconnect()
			if err != nil {
				t.Fatal(err)
			}
		}

		c := c
		recvHandler := func(ctx context.Context, resp []byte) (err error) {
			test.Assert(t, "resp", c.exp, resp, true)
			return
		}

		err := _wsClient.Send(context.Background(), c.req, recvHandler)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testClientText(t *testing.T) {
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
		exp: concatBytes(ControlFrameCloseWithCode, StatusBadRequest...),
	}, {
		desc:      "Small payload, masked",
		reconnect: true,
		req: []byte{
			0x81, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
		exp: []byte{
			0x81, 0x05,
			'H', 'e', 'l', 'l', 'o',
		},
	}, {
		desc: "Medium payload 256, unmasked",
		req:  concatBytes([]byte{0x81, 0x7E, 0x01, 0x00}, _dummyPayload256...),
		exp:  concatBytes(ControlFrameCloseWithCode, StatusBadRequest...),
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
		exp: concatBytes(ControlFrameCloseWithCode, StatusBadRequest...),
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

	for _, c := range cases {
		t.Log(c.desc)

		if c.reconnect {
			err := _wsClient.Reconnect()
			if err != nil {
				t.Fatal(err)
			}
		}

		c := c
		recvHandler := func(ctx context.Context, resp []byte) (err error) {
			test.Assert(t, "", c.exp, resp, true)
			return
		}

		err := _wsClient.Send(context.Background(), c.req, recvHandler)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func testClientFragmentation(t *testing.T) {
	cases := []struct {
		desc      string
		reconnect bool
		frames    []Frame
		exps      [][]byte
	}{{
		desc: "Two text frames, unmasked",
		frames: []Frame{{
			Fin:     0,
			Opcode:  OpCodeText,
			Payload: []byte{'H', 'e', 'l'},
		}, {
			Fin:     FrameIsFinished,
			Opcode:  OpCodeCont,
			Payload: []byte{'l', 'o'},
		}},
		exps: [][]byte{
			concatBytes(ControlFrameCloseWithCode, StatusBadRequest...),
		},
	}, {
		desc:      "Three text frames, unmasked",
		reconnect: true,
		frames: []Frame{{
			Fin:     0,
			Opcode:  OpCodeText,
			Payload: []byte("Hel"),
		}, {
			Fin:     0,
			Opcode:  OpCodeCont,
			Payload: []byte("lo, "),
		}, {
			Fin:     FrameIsFinished,
			Opcode:  OpCodeCont,
			Payload: []byte("Shulhan"),
		}},
		exps: [][]byte{
			concatBytes(ControlFrameCloseWithCode, StatusBadRequest...),
		},
	}, {
		desc:      "Three text frames with control message in the middle",
		reconnect: true,
		frames: []Frame{{
			Fin:     0,
			Opcode:  OpCodeText,
			Masked:  FrameIsMasked,
			Payload: []byte("Hel"),
		}, {
			Fin:     0,
			Opcode:  OpCodeCont,
			Masked:  FrameIsMasked,
			Payload: []byte("lo, "),
		}, {
			Fin:     FrameIsFinished,
			Opcode:  OpCodePing,
			Masked:  FrameIsMasked,
			Payload: []byte("PING"),
		}, {
			Fin:     FrameIsFinished,
			Opcode:  OpCodeCont,
			Masked:  FrameIsMasked,
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
			err := _wsClient.Reconnect()
			if err != nil {
				t.Fatal(err)
			}
		}

		for x := 0; x < len(c.frames); x++ {
			req := c.frames[x].Pack(true)

			err := _wsClient.Send(context.Background(), req, nil)
			if err != nil {
				t.Fatal(err)
			}
		}

		for x := 0; x < len(c.exps); x++ {
			res, err := _wsClient.Recv()
			if err != nil {
				t.Fatal(err)
			}

			test.Assert(t, "res", c.exps[x], res, true)
		}
	}
}

func TestClient(t *testing.T) {
	var (
		err error
	)

	addr := _testWSAddr + "?" + _qKeyTicket + "=" + _testExternalJWT

	_wsClient, err = NewClient(addr, nil)
	if err != nil {
		t.Fatal(err)
		os.Exit(1)
	}

	if _wsClient.State != ConnStateConnected {
		t.Fatal("Client is not connected")
		os.Exit(1)
	}

	t.Run("Ping", testClientPing)
	t.Run("Text", testClientText)
	t.Run("Fragmentation", testClientFragmentation)
}
