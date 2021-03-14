// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestSpell_parseMap(t *testing.T) {
	cases := []struct {
		arg string
		exp []charsmap
	}{{
		arg: "uü",
		exp: []charsmap{
			{"u", "ü"},
		},
	}, {
		arg: "ß(ss)",
		exp: []charsmap{
			{"ß", "ss"},
		},
	}, {
		arg: "ﬁ(fi)",
		exp: []charsmap{
			{"ﬁ", "fi"},
		},
	}, {
		arg: "(ọ́)o",
		exp: []charsmap{
			{"ọ́", "o"},
		},
	}}

	spell := &Spell{}

	for _, c := range cases {
		spell.opts.charsMaps = make([]charsmap, 0, 1)

		err := spell.opts.parseMap(c.arg)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "Map", c.exp, spell.opts.charsMaps)
	}
}
