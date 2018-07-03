// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func testFramePack(t *testing.T) {

	cases := []struct {
		desc string
		f    Frame
		exp  []byte
	}{{
		desc: "A single-frame unmasked text message",
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeText,
			Masked:  0,
			Payload: []byte{'H', 'e', 'l', 'l', 'o'},
		},
		exp: []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
	}, {
		desc: "A single-frame masked text message",
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeText,
			Masked:  FrameIsMasked,
			Payload: []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey: _testMaskKey,
		},
		exp: []byte{
			0x81, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
	}, {
		desc: "A fragmented unmasked text message",
		f: Frame{
			Fin:     0,
			Opcode:  OpCodeText,
			Masked:  0,
			Payload: []byte{'H', 'e', 'l'},
		},
		exp: []byte{0x01, 0x03, 0x48, 0x65, 0x6c},
	}, {
		desc: "A fragmented unmasked text message",
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeCont,
			Masked:  0,
			Payload: []byte{'l', 'o'},
		},
		exp: []byte{0x80, 0x02, 0x6c, 0x6f},
	}, {
		desc: `Unmasked Ping request (contains a body of "Hello")`,
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodePing,
			Masked:  0,
			Payload: []byte{'H', 'e', 'l', 'l', 'o'},
		},
		exp: []byte{0x89, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
	}, {
		desc: `Masked Ping response (Pong)`,
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodePong,
			Masked:  FrameIsMasked,
			Payload: []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey: _testMaskKey,
		},
		exp: []byte{
			0x8a, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
	}, {
		desc: `256 bytes binary message in a single unmasked frame`,
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeBin,
			Masked:  0,
			Payload: _dummyPayload256,
		},
		exp: concatBytes([]byte{0x82, 0x7E, 0x01, 0x00}, _dummyPayload256...),
	}, {
		desc: `256 bytes binary message in a single masked frame`,
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeBin,
			Masked:  FrameIsMasked,
			Payload: _dummyPayload256,
			maskKey: _testMaskKey,
		},
		exp: concatBytes([]byte{
			0x82, 0xFE,
			0x01, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload256Masked...),
	}, {
		desc: `65536 binary message in a single unmasked frame`,
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeBin,
			Masked:  0,
			Payload: _dummyPayload65536,
		},
		exp: concatBytes([]byte{
			0x82, 0x7F,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}, _dummyPayload65536...),
	}, {
		desc: `65536 binary message in a single masked frame`,
		f: Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeBin,
			Masked:  FrameIsMasked,
			Payload: _dummyPayload65536,
			maskKey: _testMaskKey,
			len:     65536,
		},
		exp: concatBytes([]byte{
			0x82, 0xFF,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload65536Masked...),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.f.Pack(false)

		test.Assert(t, "", c.exp, got, true)
	}
}

func testFrameUnpack(t *testing.T) {
	cases := []struct {
		desc string
		in   []byte
		exp  *Frame
	}{{
		desc: "A single-frame unmasked text message",
		in:   []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeText,
			Masked:  0,
			Payload: []byte{'H', 'e', 'l', 'l', 'o'},
			len:     5,
		},
	}, {
		desc: "A single-frame masked text message",
		in: []byte{
			0x81, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeText,
			Masked:  FrameIsMasked,
			Payload: []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey: _testMaskKey,
			len:     5,
		},
	}, {
		desc: "A fragmented unmasked text message",
		in:   []byte{0x01, 0x03, 0x48, 0x65, 0x6c},
		exp: &Frame{
			Fin:     0,
			Opcode:  OpCodeText,
			Masked:  0,
			Payload: []byte{'H', 'e', 'l'},
			len:     3,
		},
	}, {
		desc: "A fragmented unmasked text message",
		in:   []byte{0x80, 0x02, 0x6c, 0x6f},
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeCont,
			Masked:  0,
			Payload: []byte{'l', 'o'},
			len:     2,
		},
	}, {
		desc: `Unmasked Ping request (contains a body of "Hello")`,
		in:   []byte{0x89, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodePing,
			Masked:  0,
			Payload: []byte{'H', 'e', 'l', 'l', 'o'},
			len:     5,
		},
	}, {
		desc: `Pong without payload`,
		in: []byte{
			0x8A, 0x80,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		},
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodePong,
			Masked:  FrameIsMasked,
			maskKey: _testMaskKey,
		},
	}, {
		desc: `Pong with payload`,
		in: []byte{
			0x8a, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodePong,
			Masked:  FrameIsMasked,
			Payload: []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey: _testMaskKey,
			len:     5,
		},
	}, {
		desc: `256 bytes binary message in a single unmasked frame`,
		in:   concatBytes([]byte{0x82, 0x7E, 0x01, 0x00}, _dummyPayload256...),
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeBin,
			Masked:  0,
			Payload: _dummyPayload256,
			len:     256,
		},
	}, {
		desc: `256 bytes binary message in a single masked frame`,
		in: concatBytes([]byte{
			0x82, 0xFE, 0x01, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload256Masked...),
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeBin,
			Masked:  FrameIsMasked,
			Payload: _dummyPayload256,
			maskKey: _testMaskKey,
			len:     256,
		},
	}, {
		desc: `65536 binary message in a single unmasked frame`,
		in: concatBytes([]byte{
			0x82, 0x7F,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}, _dummyPayload65536...),
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeBin,
			Masked:  0,
			Payload: _dummyPayload65536,
			len:     65536,
		},
	}, {
		desc: `65536 binary message in a single masked frame`,
		in: concatBytes([]byte{
			0x82, 0xFF,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload65536Masked...),
		exp: &Frame{
			Fin:     FrameIsFinished,
			Opcode:  OpCodeBin,
			Masked:  FrameIsMasked,
			Payload: _dummyPayload65536,
			maskKey: _testMaskKey,
			len:     65536,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		gots := Unpack(c.in)

		test.Assert(t, "", c.exp, gots[0], true)
	}
}

func TestFrame(t *testing.T) {
	t.Run("Pack", testFramePack)
	t.Run("Unpack", testFrameUnpack)
}
