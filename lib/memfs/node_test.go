// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"io"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNode_Read(t *testing.T) {
	node := &Node{
		size: 3,
		V:    []byte("123"),
	}

	p := make([]byte, 1)

	cases := []struct {
		desc     string
		p        []byte
		exp      []byte
		expN     int
		expError error
	}{{
		desc: "With empty p",
	}, {
		desc: "With buffer 1 byte (1)",
		p:    p,
		exp:  []byte(`1`),
		expN: 1,
	}, {
		desc: "With buffer 1 byte (2)",
		p:    p,
		exp:  []byte(`2`),
		expN: 1,
	}, {
		desc: "With buffer 1 byte (3)",
		p:    p,
		exp:  []byte(`3`),
		expN: 1,
	}, {
		desc:     "With buffer 1 byte (4)",
		p:        p,
		exp:      []byte(`3`),
		expN:     0,
		expError: io.EOF,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		n, err := node.Read(c.p)

		test.Assert(t, "p", c.exp, c.p, true)
		test.Assert(t, "n", c.expN, n, true)
		test.Assert(t, "error", c.expError, err, true)
	}
}

func TestNode_Seek(t *testing.T) {
	node := &Node{
		size: 3,
		V:    []byte("123"),
	}

	cases := []struct {
		desc     string
		offset   int64
		whence   int
		exp      int64
		expError error
	}{{
		desc:     "With invalid whence",
		offset:   5,
		whence:   3,
		expError: errWhence,
	}, {
		desc:     "With invalid offset",
		offset:   -5,
		whence:   io.SeekStart,
		expError: errOffset,
	}, {
		desc:   "SeekStart",
		offset: 5,
		whence: io.SeekStart,
		exp:    5,
	}, {
		desc:   "SeekCurrent",
		offset: 5,
		whence: io.SeekCurrent,
		exp:    10,
	}, {
		desc:   "SeekEnd",
		offset: 5,
		whence: io.SeekEnd,
		exp:    8,
	}}
	for _, c := range cases {
		t.Log(c.desc)

		got, err := node.Seek(c.offset, c.whence)

		test.Assert(t, "Seek", c.exp, got, true)
		test.Assert(t, "Seek error", c.expError, err, true)
	}
}
