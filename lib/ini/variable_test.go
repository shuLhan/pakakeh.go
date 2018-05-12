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
		v    *Variable
		exp  string
	}{{
		desc: "With mode empty #1",
		v: &Variable{
			mode: varModeEmpty,
		},
	}, {
		desc: "With mode empty #2",
		v: &Variable{
			mode: varModeEmpty,
			format: "	",
		},
		exp: "	",
	}, {
		desc: "With mode comment #1",
		v: &Variable{
			mode:   varModeComment,
			format: "  %s",
			others: "; comment",
		},
		exp: "  ; comment",
	}, {
		desc: "With mode comment #2",
		v: &Variable{
			mode:   varModeComment,
			others: "; comment",
		},
		exp: "; comment\n",
	}, {
		desc: "With mode section",
		v: &Variable{
			mode:    varModeSection,
			secName: "section",
		},
		exp: "[section]\n",
	}, {
		desc: "With mode section and comment #1",
		v: &Variable{
			mode:    varModeSection | varModeComment,
			secName: "section",
			others:  "; comment",
		},
		exp: "[section] ; comment\n",
	}, {
		desc: "With mode section and comment #2",
		v: &Variable{
			mode:    varModeSection | varModeComment,
			format:  " [%s]   %s",
			secName: "section",
			others:  "; comment",
		},
		exp: " [section]   ; comment",
	}, {
		desc: "With mode section and subsection",
		v: &Variable{
			mode:    varModeSection | varModeSubsection,
			secName: "section",
			subName: "subsection",
		},
		exp: `[section "subsection"]\n`,
	}, {
		desc: "With mode section, subsection, and comment",
		v: &Variable{
			mode:    varModeSection | varModeSubsection | varModeComment,
			secName: "section",
			subName: "subsection",
			others:  "; comment",
		},
		exp: `[section "subsection"] ; comment\n`,
	}, {
		desc: "With mode single",
		v: &Variable{
			mode: varModeSingle,
			Key:  "name",
		},
		exp: "name = true\n",
	}, {
		desc: "With mode single and comment",
		v: &Variable{
			mode:   varModeSingle | varModeComment,
			Key:    "name",
			others: "; comment",
		},
		exp: "name = true ; comment\n",
	}, {
		desc: "With mode value",
		v: &Variable{
			mode:  varModeValue,
			Key:   "name",
			Value: "value",
		},
		exp: "name = value\n",
	}, {
		desc: "With mode value and comment",
		v: &Variable{
			mode:   varModeValue | varModeComment,
			Key:    "name",
			Value:  "value",
			others: "; comment",
		},
		exp: "name = value ; comment\n",
	}, {
		desc: "With mode multi",
		v: &Variable{
			mode:  varModeMulti,
			Key:   "name",
			Value: "value",
		},
		exp: "name = value\n",
	}, {
		desc: "With mode multi and comment",
		v: &Variable{
			mode:   varModeMulti | varModeComment,
			Key:    "name",
			Value:  "value",
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
