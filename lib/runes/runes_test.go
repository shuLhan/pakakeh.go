// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package runes

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestDiff(t *testing.T) {
	cases := []struct {
		l   []rune
		r   []rune
		exp []rune
	}{{
		l:   []rune{'a', 'b', 'a', 'b', 'c', 'd'},
		r:   []rune{'d', 'c', 'a', 'e'},
		exp: []rune{'b', 'e'},
	}, {
		l:   []rune{'a', 'b', 'a', 'b', 'c', 'd'},
		r:   []rune{'d', 'c', 'a', 'e'},
		exp: []rune{'b', 'e'},
	}, {
		l:   []rune{'a', 'b', 'a', 'b', 'c', 'd'},
		r:   []rune{'d', 'c', 'a', 'b', 'a', 'b', 'e'},
		exp: []rune{'e'},
	}, {
		l:   []rune{'d', 'c', 'a', 'b', 'a', 'b', 'e'},
		r:   []rune{'a', 'b', 'f', 'a', 'b', 'c', 'd'},
		exp: []rune{'e', 'f'},
	},
	}
	for _, c := range cases {
		got := Diff(c.l, c.r)

		test.Assert(t, "", string(c.exp), string(got), true)
	}
}

//nolint:dupl
func TestEncloseRemove(t *testing.T) {
	line := []rune(`// Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`)

	cases := []struct {
		line     []rune
		leftcap  []rune
		rightcap []rune
		exp      string
	}{{
		line:     line,
		leftcap:  []rune("<"),
		rightcap: []rune(">"),
		exp:      `// Copyright 2016-2018 "Shulhan ". All rights reserved.`,
	}, {
		line:     line,
		leftcap:  []rune(`"`),
		rightcap: []rune(`"`),
		exp:      `// Copyright 2016-2018 . All rights reserved.`,
	}, {
		line:     line,
		leftcap:  []rune(`/`),
		rightcap: []rune(`/`),
		exp:      ` Copyright 2016-2018 "Shulhan <ms@kilabit.info>". All rights reserved.`,
	}, {
		line:     []rune(`/* TEST */`),
		leftcap:  []rune(`/*`),
		rightcap: []rune(`*/`),
		exp:      "",
	}}

	for _, c := range cases {
		got, _ := EncloseRemove(c.line, c.leftcap, c.rightcap)

		test.Assert(t, "", c.exp, string(got), true)
	}
}
