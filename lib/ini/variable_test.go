// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestVariableIsValueBoolTrue(t *testing.T) {
	cases := []struct {
		desc string
		v    *Variable
		exp  bool
	}{{
		desc: "With empty value",
		v:    &Variable{},
		exp:  true,
	}, {
		desc: "With value in all caps",
		v: &Variable{
			Value: "TRUE",
		},
		exp: true,
	}, {
		desc: "With value is yes",
		v: &Variable{
			Value: "YES",
		},
		exp: true,
	}, {
		desc: "With value is ya",
		v: &Variable{
			Value: "yA",
		},
		exp: true,
	}, {
		desc: "With value is 1",
		v: &Variable{
			Value: "1",
		},
		exp: true,
	}, {
		desc: "With value is 11",
		v: &Variable{
			Value: "11",
		},
		exp: false,
	}, {
		desc: "With value is tru (typo)",
		v: &Variable{
			Value: "tru",
		},
		exp: false,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.v.IsValueBoolTrue()

		test.Assert(t, "", c.exp, got, true)
	}
}

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
