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
			format: []byte("	"),
		},
		exp: "	",
	}, {
		desc: "With mode comment #1",
		v: &Variable{
			mode:   varModeComment,
			format: []byte("  %s"),
			others: []byte("; comment"),
		},
		exp: "  ; comment",
	}, {
		desc: "With mode comment #2",
		v: &Variable{
			mode:   varModeComment,
			others: []byte("; comment"),
		},
		exp: "; comment\n",
	}, {
		desc: "With mode section",
		v: &Variable{
			mode:    varModeSection,
			secName: []byte("section"),
		},
		exp: "[section]\n",
	}, {
		desc: "With mode section and comment #1",
		v: &Variable{
			mode:    varModeSection | varModeComment,
			secName: []byte("section"),
			others:  []byte("; comment"),
		},
		exp: "[section] ; comment\n",
	}, {
		desc: "With mode section and comment #2",
		v: &Variable{
			mode:    varModeSection | varModeComment,
			format:  []byte(" [%s]   %s"),
			secName: []byte("section"),
			others:  []byte("; comment"),
		},
		exp: " [section]   ; comment",
	}, {
		desc: "With mode section and subsection",
		v: &Variable{
			mode:    varModeSection | varModeSubsection,
			secName: []byte("section"),
			subName: []byte("subsection"),
		},
		exp: `[section "subsection"]\n`,
	}, {
		desc: "With mode section, subsection, and comment",
		v: &Variable{
			mode:    varModeSection | varModeSubsection | varModeComment,
			secName: []byte("section"),
			subName: []byte("subsection"),
			others:  []byte("; comment"),
		},
		exp: `[section "subsection"] ; comment\n`,
	}, {
		desc: "With mode single",
		v: &Variable{
			mode: varModeSingle,
			key:  []byte("name"),
		},
		exp: "name = true\n",
	}, {
		desc: "With mode single and comment",
		v: &Variable{
			mode:   varModeSingle | varModeComment,
			key:    []byte("name"),
			others: []byte("; comment"),
		},
		exp: "name = true ; comment\n",
	}, {
		desc: "With mode value",
		v: &Variable{
			mode:  varModeValue,
			key:   []byte("name"),
			value: []byte("value"),
		},
		exp: "name = value\n",
	}, {
		desc: "With mode value and comment",
		v: &Variable{
			mode:   varModeValue | varModeComment,
			key:    []byte("name"),
			value:  []byte("value"),
			others: []byte("; comment"),
		},
		exp: "name = value ; comment\n",
	}, {
		desc: "With mode multi",
		v: &Variable{
			mode:  varModeMulti,
			key:   []byte("name"),
			value: []byte("value"),
		},
		exp: "name = value\n",
	}, {
		desc: "With mode multi and comment",
		v: &Variable{
			mode:   varModeMulti | varModeComment,
			key:    []byte("name"),
			value:  []byte("value"),
			others: []byte("; comment"),
		},
		exp: "name = value ; comment\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := c.v.String()

		test.Assert(t, "", c.exp, got, true)
	}
}
