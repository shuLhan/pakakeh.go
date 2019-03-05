// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/test"
)

func TestNewFrameBin(t *testing.T) { //nolint: dupl
	cases := []struct {
		desc     string
		isMasked bool
		payload  []byte
		exp      *Frame
	}{{
		desc:    "With unmasked",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			payload: []byte("Hello!"),
		},
	}, {
		desc:     "With masked",
		isMasked: true,
		payload:  []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  frameIsMasked,
			payload: []byte("Hello!"),
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		packet := NewFrameBin(c.isMasked, c.payload)
		frames := Unpack(packet)
		frame := frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin, true)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode, true)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked, true)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload, true)
	}
}

func TestNewFrameClose(t *testing.T) {
	cases := []struct {
		desc    string
		payload []byte
		exp     *Frame
	}{{
		desc:    "With small payload",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:       frameIsFinished,
			opcode:    opcodeClose,
			closeCode: StatusBadRequest,
			masked:    frameIsMasked,
			payload:   []byte("Hello!"),
		},
	}, {
		desc:    "With overflow payload",
		payload: _dummyPayload256,
		exp: &Frame{
			fin:       frameIsFinished,
			opcode:    opcodeClose,
			closeCode: StatusBadRequest,
			masked:    frameIsMasked,
			payload:   _dummyPayload256[:123],
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		packet := NewFrameClose(true, StatusBadRequest, c.payload)
		libbytes.PrintHex("CLOSE frame unmasked", packet, 8)
		frames := Unpack(packet)
		frame := frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin, true)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode, true)
		test.Assert(t, "Frame.closeCode", c.exp.closeCode, frame.closeCode, true)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked, true)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload, true)
	}
}

func TestNewFramePing(t *testing.T) { //nolint:dupl
	cases := []struct {
		desc    string
		payload []byte
		exp     *Frame
	}{{
		desc:    "With small payload",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodePing,
			masked:  frameIsMasked,
			payload: []byte("Hello!"),
		},
	}, {
		desc:    "With overflow payload",
		payload: _dummyPayload256,
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodePing,
			masked:  frameIsMasked,
			payload: _dummyPayload256[:125],
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		packet := NewFramePing(true, c.payload)
		frames := Unpack(packet)
		frame := frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin, true)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode, true)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked, true)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload, true)
	}
}

func TestNewFramePong(t *testing.T) { //nolint: dupl
	cases := []struct {
		desc    string
		payload []byte
		exp     *Frame
	}{{
		desc:    "With small payload",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodePong,
			masked:  frameIsMasked,
			payload: []byte("Hello!"),
		},
	}, {
		desc:    "With overflow payload",
		payload: _dummyPayload256,
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodePong,
			masked:  frameIsMasked,
			payload: _dummyPayload256[:125],
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		packet := NewFramePong(true, c.payload)
		frames := Unpack(packet)
		frame := frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin, true)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode, true)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked, true)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload, true)
	}
}

func TestNewFrameText(t *testing.T) { //nolint: dupl
	cases := []struct {
		desc     string
		isMasked bool
		payload  []byte
		exp      *Frame
	}{{
		desc:    "With unmasked",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodeText,
			payload: []byte("Hello!"),
		},
	}, {
		desc:     "With masked",
		isMasked: true,
		payload:  []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodeText,
			masked:  frameIsMasked,
			payload: []byte("Hello!"),
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		packet := NewFrameText(c.isMasked, c.payload)
		frames := Unpack(packet)
		frame := frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin, true)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode, true)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked, true)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload, true)
	}
}

func TestFramePack(t *testing.T) {
	cases := []struct {
		desc string
		f    Frame
		exp  []byte
	}{{
		desc: "A single-frame unmasked text message",
		f: Frame{
			fin:     frameIsFinished,
			opcode:  opcodeText,
			masked:  0,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
		},
		exp: []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
	}, {
		desc: "A single-frame masked text message",
		f: Frame{
			fin:     frameIsFinished,
			opcode:  opcodeText,
			masked:  frameIsMasked,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
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
			fin:     0,
			opcode:  opcodeText,
			masked:  0,
			payload: []byte{'H', 'e', 'l'},
		},
		exp: []byte{0x01, 0x03, 0x48, 0x65, 0x6c},
	}, {
		desc: "A fragmented unmasked text message",
		f: Frame{
			fin:     frameIsFinished,
			opcode:  opcodeCont,
			masked:  0,
			payload: []byte{'l', 'o'},
		},
		exp: []byte{0x80, 0x02, 0x6c, 0x6f},
	}, {
		desc: `Unmasked Ping request (contains a body of "Hello")`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  opcodePing,
			masked:  0,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
		},
		exp: []byte{0x89, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
	}, {
		desc: `masked Ping response (Pong)`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  opcodePong,
			masked:  frameIsMasked,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
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
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  0,
			payload: _dummyPayload256,
		},
		exp: libbytes.Concat([]byte{0x82, 0x7E, 0x01, 0x00}, _dummyPayload256),
	}, {
		desc: `256 bytes binary message in a single masked frame`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  frameIsMasked,
			payload: _dummyPayload256,
			maskKey: _testMaskKey,
		},
		exp: libbytes.Concat([]byte{
			0x82, 0xFE,
			0x01, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload256Masked),
	}, {
		desc: `65536 binary message in a single unmasked frame`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  0,
			payload: _dummyPayload65536,
		},
		exp: libbytes.Concat([]byte{
			0x82, 0x7F,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}, _dummyPayload65536),
	}, {
		desc: `65536 binary message in a single masked frame`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  frameIsMasked,
			payload: _dummyPayload65536,
			maskKey: _testMaskKey,
			len:     65536,
		},
		exp: libbytes.Concat([]byte{
			0x82, 0xFF,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload65536Masked),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.f.pack(false)

		test.Assert(t, "", c.exp, got, true)
	}
}
