// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestVariableString(t *testing.T) {
	cases := []struct {
		desc string
		v    *variable
		exp  string
	}{{
		desc: "With mode empty #1",
		v: &variable{
			mode: varModeEmpty,
		},
	}, {
		desc: "With mode empty #2",
		v: &variable{
			mode: varModeEmpty,
			format: "	",
		},
		exp: "	",
	}, {
		desc: "With mode comment #1",
		v: &variable{
			mode:   varModeComment,
			format: "  %s",
			others: "; comment",
		},
		exp: "  ; comment",
	}, {
		desc: "With mode comment #2",
		v: &variable{
			mode:   varModeComment,
			others: "; comment",
		},
		exp: "; comment\n",
	}, {
		desc: "With mode section",
		v: &variable{
			mode:    varModeSection,
			secName: "section",
		},
		exp: "[section]\n",
	}, {
		desc: "With mode section and comment #1",
		v: &variable{
			mode:    varModeSection | varModeComment,
			secName: "section",
			others:  "; comment",
		},
		exp: "[section] ; comment\n",
	}, {
		desc: "With mode section and comment #2",
		v: &variable{
			mode:    varModeSection | varModeComment,
			format:  " [%s]   %s",
			secName: "section",
			others:  "; comment",
		},
		exp: " [section]   ; comment",
	}, {
		desc: "With mode section and subsection",
		v: &variable{
			mode:    varModeSection | varModeSubsection,
			secName: "section",
			subName: "subsection",
		},
		exp: `[section "subsection"]\n`,
	}, {
		desc: "With mode section, subsection, and comment",
		v: &variable{
			mode:    varModeSection | varModeSubsection | varModeComment,
			secName: "section",
			subName: "subsection",
			others:  "; comment",
		},
		exp: `[section "subsection"] ; comment\n`,
	}, {
		desc: "With mode single",
		v: &variable{
			mode: varModeSingle,
			key:  "name",
		},
		exp: "name = true\n",
	}, {
		desc: "With mode single and comment",
		v: &variable{
			mode:   varModeSingle | varModeComment,
			key:    "name",
			others: "; comment",
		},
		exp: "name = true ; comment\n",
	}, {
		desc: "With mode value",
		v: &variable{
			mode:  varModeValue,
			key:   "name",
			value: "value",
		},
		exp: "name = value\n",
	}, {
		desc: "With mode value and comment",
		v: &variable{
			mode:   varModeValue | varModeComment,
			key:    "name",
			value:  "value",
			others: "; comment",
		},
		exp: "name = value ; comment\n",
	}, {
		desc: "With mode multi",
		v: &variable{
			mode:  varModeMulti,
			key:   "name",
			value: "value",
		},
		exp: "name = value\n",
	}, {
		desc: "With mode multi and comment",
		v: &variable{
			mode:   varModeMulti | varModeComment,
			key:    "name",
			value:  "value",
			others: "; comment",
		},
		exp: "name = value ; comment\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.v.String()

		test.Assert(t, "", c.exp, got, true)
	}
}
