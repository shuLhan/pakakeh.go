// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/test"
)

func TestFrameUnpack(t *testing.T) {
	cases := []struct {
		desc string
		in   []byte
		exp  *Frame
	}{{
		desc: "With empty input",
	}, {
		desc: "A single-frame unmasked text message",
		in:   []byte{0x81, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			masked:     0,
			payload:    []byte{'H', 'e', 'l', 'l', 'o'},
			len:        5,
			isComplete: true,
		},
	}, {
		desc: "A single-frame masked text message",
		in: []byte{
			0x81, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeText,
			masked:     frameIsMasked,
			payload:    []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey:    _testMaskKey,
			len:        5,
			isComplete: true,
		},
	}, {
		desc: "A fragmented unmasked text message",
		in:   []byte{0x01, 0x03, 0x48, 0x65, 0x6c},
		exp: &Frame{
			fin:        0,
			opcode:     OpcodeText,
			masked:     0,
			payload:    []byte{'H', 'e', 'l'},
			len:        3,
			isComplete: true,
		},
	}, {
		desc: "A fragmented unmasked text message",
		in:   []byte{0x80, 0x02, 0x6c, 0x6f},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeCont,
			masked:     0,
			payload:    []byte{'l', 'o'},
			len:        2,
			isComplete: true,
		},
	}, {
		desc: `Unmasked Ping request (contains a body of "Hello")`,
		in:   []byte{0x89, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePing,
			masked:     0,
			payload:    []byte{'H', 'e', 'l', 'l', 'o'},
			len:        5,
			isComplete: true,
		},
	}, {
		desc: `Pong without payload`,
		in: []byte{
			0x8A, 0x80,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePong,
			masked:     frameIsMasked,
			maskKey:    _testMaskKey,
			isComplete: true,
		},
	}, {
		desc: `Pong with payload`,
		in: []byte{
			0x8a, 0x85,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
			0x7f, 0x9f, 0x4d, 0x51, 0x58,
		},
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodePong,
			masked:     frameIsMasked,
			payload:    []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey:    _testMaskKey,
			len:        5,
			isComplete: true,
		},
	}, {
		desc: `256 bytes binary message in a single unmasked frame`,
		in:   libbytes.Concat([]byte{0x82, 0x7E, 0x01, 0x00}, _dummyPayload256),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeBin,
			masked:     0,
			payload:    _dummyPayload256,
			len:        256,
			isComplete: true,
		},
	}, {
		desc: `256 bytes binary message in a single masked frame`,
		in: libbytes.Concat([]byte{
			0x82, 0xFE, 0x01, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload256Masked),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeBin,
			masked:     frameIsMasked,
			payload:    _dummyPayload256,
			maskKey:    _testMaskKey,
			len:        256,
			isComplete: true,
		},
	}, {
		desc: `65536 binary message in a single unmasked frame`,
		in: libbytes.Concat([]byte{
			0x82, 0x7F,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		}, _dummyPayload65536),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeBin,
			masked:     0,
			payload:    _dummyPayload65536,
			len:        65536,
			isComplete: true,
		},
	}, {
		desc: `65536 binary message in a single masked frame`,
		in: libbytes.Concat([]byte{
			0x82, 0xFF,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload65536Masked),
		exp: &Frame{
			fin:        frameIsFinished,
			opcode:     OpcodeBin,
			masked:     frameIsMasked,
			payload:    _dummyPayload65536,
			maskKey:    _testMaskKey,
			len:        65536,
			isComplete: true,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		gots := Unpack(c.in)

		if gots != nil && len(gots.v) > 0 {
			test.Assert(t, "", c.exp, gots.v[0], true)
		}
	}
}

func TestFramesAppend(t *testing.T) {
	frames := &Frames{}

	cases := []struct {
		desc       string
		f          *Frame
		expLen     int
		expPayload string
	}{{
		desc:       "With nil frame",
		expLen:     0,
		expPayload: "",
	}, {
		desc: "With non nil frames",
		f: &Frame{
			opcode:  OpcodeText,
			payload: []byte("A"),
		},
		expLen:     1,
		expPayload: "A",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		frames.Append(c.f)

		test.Assert(t, "Frames.Len", c.expLen, len(frames.v), true)
		test.Assert(t, "Frames.payload", c.expPayload,
			string(frames.payload()), true)
	}
}

func TestFramesIsClosed(t *testing.T) {
	cases := []struct {
		desc   string
		frames *Frames
		exp    bool
	}{{
		desc:   "With empty frames",
		frames: &Frames{},
	}, {
		desc: "With no close frames",
		frames: &Frames{
			v: []*Frame{{
				opcode: OpcodeText,
			}},
		},
	}, {
		desc: "With close frames at the end",
		frames: &Frames{
			v: []*Frame{{
				opcode: OpcodeText,
			}, {
				opcode: OpcodeText,
			}, {
				opcode: OpcodeClose,
			}},
		},
		exp: true,
	}}

	for _, c := range cases {
		t.Log(c.desc)
		got := c.frames.isClosed()
		test.Assert(t, "Frames.isClosed", c.exp, got, true)
	}
}

func TestFramesPayload(t *testing.T) {
	cases := []struct {
		desc string
		fs   *Frames
		exp  string
	}{{
		desc: "With empty frames",
		fs:   &Frames{},
	}, {
		desc: "With the first frame is CLOSE",
		fs: &Frames{
			v: []*Frame{{
				fin:     frameIsFinished,
				opcode:  OpcodeClose,
				payload: []byte{0, 0},
			}},
		},
	}, {
		desc: "With data frame",
		fs: &Frames{
			v: []*Frame{{
				fin:     0,
				opcode:  OpcodeText,
				payload: []byte("Hel"),
			}, {
				fin:     0,
				opcode:  0,
				payload: []byte("lo "),
			}, {
				fin:     frameIsFinished,
				opcode:  0,
				payload: []byte("world!"),
			}},
		},
		exp: "Hello world!",
	}, {
		desc: "With interjected CLOSE frame",
		fs: &Frames{
			v: []*Frame{{
				fin:     0,
				opcode:  OpcodeText,
				payload: []byte("Hel"),
			}, {
				fin:     frameIsFinished,
				opcode:  OpcodeClose,
				payload: []byte("lo "),
			}, {
				fin:     frameIsFinished,
				opcode:  OpcodeCont,
				payload: []byte("world!"),
			}},
		},
		exp: "Hel",
	}, {
		desc: "With fin frame in the middle",
		fs: &Frames{
			v: []*Frame{{
				fin:     0,
				opcode:  OpcodeText,
				payload: []byte("Hel"),
			}, {
				fin:     frameIsFinished,
				opcode:  OpcodeCont,
				payload: []byte("lo "),
			}, {
				fin:     frameIsFinished,
				opcode:  OpcodeCont,
				payload: []byte("world!"),
			}},
		},
		exp: "Hello ",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.fs.payload()

		test.Assert(t, "Frames.payload", c.exp, string(got), true)
	}
}
