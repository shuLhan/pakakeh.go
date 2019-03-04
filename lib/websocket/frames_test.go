// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

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
