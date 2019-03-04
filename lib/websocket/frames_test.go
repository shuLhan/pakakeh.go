// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

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
			fin:     frameIsFinished,
			opcode:  opcodeText,
			masked:  0,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
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
			fin:     frameIsFinished,
			opcode:  opcodeText,
			masked:  frameIsMasked,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey: _testMaskKey,
			len:     5,
		},
	}, {
		desc: "A fragmented unmasked text message",
		in:   []byte{0x01, 0x03, 0x48, 0x65, 0x6c},
		exp: &Frame{
			fin:     0,
			opcode:  opcodeText,
			masked:  0,
			payload: []byte{'H', 'e', 'l'},
			len:     3,
		},
	}, {
		desc: "A fragmented unmasked text message",
		in:   []byte{0x80, 0x02, 0x6c, 0x6f},
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodeCont,
			masked:  0,
			payload: []byte{'l', 'o'},
			len:     2,
		},
	}, {
		desc: `Unmasked Ping request (contains a body of "Hello")`,
		in:   []byte{0x89, 0x05, 0x48, 0x65, 0x6c, 0x6c, 0x6f},
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodePing,
			masked:  0,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
			len:     5,
		},
	}, {
		desc: `Pong without payload`,
		in: []byte{
			0x8A, 0x80,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		},
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodePong,
			masked:  frameIsMasked,
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
			fin:     frameIsFinished,
			opcode:  opcodePong,
			masked:  frameIsMasked,
			payload: []byte{'H', 'e', 'l', 'l', 'o'},
			maskKey: _testMaskKey,
			len:     5,
		},
	}, {
		desc: `256 bytes binary message in a single unmasked frame`,
		in:   concatBytes([]byte{0x82, 0x7E, 0x01, 0x00}, _dummyPayload256...),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  0,
			payload: _dummyPayload256,
			len:     256,
		},
	}, {
		desc: `256 bytes binary message in a single masked frame`,
		in: concatBytes([]byte{
			0x82, 0xFE, 0x01, 0x00,
			_testMaskKey[0], _testMaskKey[1], _testMaskKey[2], _testMaskKey[3],
		}, _dummyPayload256Masked...),
		exp: &Frame{
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  frameIsMasked,
			payload: _dummyPayload256,
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
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  0,
			payload: _dummyPayload65536,
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
			fin:     frameIsFinished,
			opcode:  opcodeBin,
			masked:  frameIsMasked,
			payload: _dummyPayload65536,
			maskKey: _testMaskKey,
			len:     65536,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		gots := Unpack(c.in)

		if gots != nil && gots.Len() > 0 {
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
			opcode:  opcodeText,
			payload: []byte("A"),
		},
		expLen:     1,
		expPayload: "A",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		frames.Append(c.f)

		test.Assert(t, "Frames.Len", c.expLen, frames.Len(), true)
		test.Assert(t, "Frames.payload", c.expPayload,
			string(frames.Payload()), true)
	}
}

func TestFramesGet(t *testing.T) {
	frames := &Frames{}

	f0 := &Frame{
		opcode:  opcodeText,
		payload: []byte("A"),
	}
	f1 := &Frame{
		opcode:  opcodeText,
		payload: []byte("B"),
	}
	f2 := &Frame{
		opcode:  opcodeText,
		payload: []byte("C"),
	}

	frames.Append(f0)
	frames.Append(f1)
	frames.Append(f2)

	cases := []struct {
		desc string
		x    int
		exp  *Frame
	}{{
		desc: "With negative index",
		x:    -1,
	}, {
		desc: "With out of range index",
		x:    frames.Len(),
	}, {
		desc: "With valid index",
		x:    0,
		exp:  f0,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := frames.Get(c.x)

		test.Assert(t, "Frames.Get", c.exp, got, true)
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
				opcode: opcodeText,
			}},
		},
	}, {
		desc: "With close frames at the end",
		frames: &Frames{
			v: []*Frame{{
				opcode: opcodeText,
			}, {
				opcode: opcodeText,
			}, {
				opcode: opcodeClose,
			}},
		},
		exp: true,
	}}

	for _, c := range cases {
		t.Log(c.desc)
		got := c.frames.IsClosed()
		test.Assert(t, "Frames.IsClosed", c.exp, got, true)
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
				opcode:  opcodeClose,
				payload: []byte{0, 0},
			}},
		},
	}, {
		desc: "With data frame",
		fs: &Frames{
			v: []*Frame{{
				fin:     0,
				opcode:  opcodeText,
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
				opcode:  opcodeText,
				payload: []byte("Hel"),
			}, {
				fin:     frameIsFinished,
				opcode:  opcodeClose,
				payload: []byte("lo "),
			}, {
				fin:     frameIsFinished,
				opcode:  opcodeCont,
				payload: []byte("world!"),
			}},
		},
		exp: "Hel",
	}, {
		desc: "With fin frame in the middle",
		fs: &Frames{
			v: []*Frame{{
				fin:     0,
				opcode:  opcodeText,
				payload: []byte("Hel"),
			}, {
				fin:     frameIsFinished,
				opcode:  opcodeCont,
				payload: []byte("lo "),
			}, {
				fin:     frameIsFinished,
				opcode:  opcodeCont,
				payload: []byte("world!"),
			}},
		},
		exp: "Hello ",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.fs.Payload()

		test.Assert(t, "Frames.payload", c.exp, string(got), true)
	}
}
