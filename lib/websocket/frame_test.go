// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestNewFrameBin(t *testing.T) {
	type testCase struct {
		exp      *Frame
		desc     string
		payload  []byte
		isMasked bool
	}

	var cases = []testCase{{
		desc:    "With unmasked",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeBin,
			payload: []byte("Hello!"),
		},
	}, {
		desc:     "With masked",
		isMasked: true,
		payload:  []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeBin,
			masked:  frameIsMasked,
			payload: []byte("Hello!"),
		},
	}}

	var (
		c      testCase
		packet []byte
		frames *Frames
		frame  *Frame
	)

	for _, c = range cases {
		t.Log(c.desc)

		packet = NewFrameBin(c.isMasked, c.payload)
		frames = Unpack(packet)
		frame = frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload)
	}
}

func TestNewFrameClose(t *testing.T) {
	type testCase struct {
		exp     *Frame
		desc    string
		payload []byte
	}

	var cases = []testCase{{
		desc:    "With small payload",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:       frameIsFinished,
			opcode:    OpcodeClose,
			closeCode: StatusBadRequest,
			masked:    frameIsMasked,
			payload: libbytes.Concat([]byte{0x03, 0xEA},
				[]byte("Hello!")),
		},
	}, {
		desc:    "With overflow payload",
		payload: _dummyPayload256,
		exp: &Frame{
			fin:       frameIsFinished,
			opcode:    OpcodeClose,
			closeCode: StatusBadRequest,
			masked:    frameIsMasked,
			payload: libbytes.Concat([]byte{0x03, 0xEA},
				_dummyPayload256[:123]),
		},
	}}

	var (
		c      testCase
		packet []byte
		frames *Frames
		frame  *Frame
	)

	for _, c = range cases {
		t.Log(c.desc)

		packet = NewFrameClose(true, StatusBadRequest, c.payload)
		frames = Unpack(packet)
		frame = frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode)
		test.Assert(t, "Frame.closeCode", c.exp.closeCode, frame.closeCode)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload)
	}
}

func TestNewFramePing(t *testing.T) {
	type testCase struct {
		exp     *Frame
		desc    string
		payload []byte
	}

	var cases = []testCase{{
		desc:    "With small payload",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodePing,
			masked:  frameIsMasked,
			payload: []byte("Hello!"),
		},
	}, {
		desc:    "With overflow payload",
		payload: _dummyPayload256,
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodePing,
			masked:  frameIsMasked,
			payload: _dummyPayload256[:125],
		},
	}}

	var (
		c      testCase
		packet []byte
		frames *Frames
		frame  *Frame
	)

	for _, c = range cases {
		t.Log(c.desc)

		packet = NewFramePing(true, c.payload)
		frames = Unpack(packet)
		frame = frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload)
	}
}

func TestNewFramePong(t *testing.T) {
	type testCase struct {
		exp     *Frame
		desc    string
		payload []byte
	}

	var cases = []testCase{{
		desc:    "With small payload",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodePong,
			masked:  frameIsMasked,
			payload: []byte("Hello!"),
		},
	}, {
		desc:    "With overflow payload",
		payload: _dummyPayload256,
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodePong,
			masked:  frameIsMasked,
			payload: _dummyPayload256[:125],
		},
	}}

	var (
		c      testCase
		frames *Frames
		frame  *Frame
		packet []byte
	)

	for _, c = range cases {
		t.Log(c.desc)

		packet = NewFramePong(true, c.payload)
		frames = Unpack(packet)
		frame = frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload)
	}
}

func TestNewFrameText(t *testing.T) {
	type testCase struct {
		exp      *Frame
		desc     string
		payload  []byte
		isMasked bool
	}

	var cases = []testCase{{
		desc:    "With unmasked",
		payload: []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeText,
			payload: []byte("Hello!"),
		},
	}, {
		desc:     "With masked",
		isMasked: true,
		payload:  []byte("Hello!"),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeText,
			masked:  frameIsMasked,
			payload: []byte("Hello!"),
		},
	}}

	var (
		c      testCase
		frames *Frames
		frame  *Frame
		packet []byte
	)

	for _, c = range cases {
		t.Log(c.desc)

		packet = NewFrameText(c.isMasked, c.payload)
		frames = Unpack(packet)
		frame = frames.v[0]

		test.Assert(t, "Frame.fin", c.exp.fin, frame.fin)
		test.Assert(t, "Frame.opcode", c.exp.opcode, frame.opcode)
		test.Assert(t, "Frame.masked", c.exp.masked, frame.masked)
		test.Assert(t, "Frame.payload", c.exp.payload, frame.payload)
	}
}

func TestFramePack(t *testing.T) {
	type testCase struct {
		desc string
		exp  []byte
		f    Frame
	}

	var cases = []testCase{{
		desc: "A single-frame unmasked text message",
		f: Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeText,
			masked:  0,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
		},
		exp: []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
	}, {
		desc: "A single-frame masked text message",
		f: Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeText,
			masked:  frameIsMasked,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey: _testMaskKey,
		},
		exp: []byte{
			0x81, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2],
			_testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
	}, {
		desc: "A fragmented unmasked text message",
		f: Frame{
			fin:     0,
			opcode:  OpcodeText,
			masked:  0,
			payload: []byte{'H', 'e', 'l'},
		},
		exp: []byte{0x01, 0x03, 0x48, 0x65, 0x6c},
	}, {
		desc: "A fragmented unmasked text message",
		f: Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeCont,
			masked:  0,
			payload: []byte{'l', 'o'},
		},
		exp: []byte{0x80, 0x02, 0x6c, 0x6f},
	}, {
		desc: `Unmasked Ping request (contains a body of "Hello")`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  OpcodePing,
			masked:  0,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
		},
		exp: []byte{0x89, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
	}, {
		desc: `masked Ping response (Pong)`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  OpcodePong,
			masked:  frameIsMasked,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey: _testMaskKey,
		},
		exp: []byte{
			0x8a, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2],
			_testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
	}, {
		desc: `256 bytes binary message in a single unmasked frame`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeBin,
			masked:  0,
			payload: _dummyPayload256,
		},
		exp: libbytes.Concat([]byte{0x82, 0x7E, 0x01, 0x00},
			_dummyPayload256),
	}, {
		desc: `256 bytes binary message in a single masked frame`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeBin,
			masked:  frameIsMasked,
			payload: _dummyPayload256,
			maskKey: _testMaskKey,
		},
		exp: libbytes.Concat([]byte{
			0x82, 0xFE,
			0x01, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2],
			_testMaskKey[3],
		}, _dummyPayload256Masked),
	}, {
		desc: `65536 binary message in a single unmasked frame`,
		f: Frame{
			fin:     frameIsFinished,
			opcode:  OpcodeBin,
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
			opcode:  OpcodeBin,
			masked:  frameIsMasked,
			payload: _dummyPayload65536,
			maskKey: _testMaskKey,
			len:     65536,
		},
		exp: libbytes.Concat([]byte{
			0x82, 0xFF,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2],
			_testMaskKey[3],
		}, _dummyPayload65536Masked),
	}}

	var (
		c   testCase
		got []byte
	)

	for _, c = range cases {
		t.Log(c.desc)

		got = c.f.pack()

		test.Assert(t, "", c.exp, got)
	}
}
