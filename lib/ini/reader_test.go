package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseVarName(t *testing.T) {
	cases := []struct {
		desc  string
		in    []byte
		exp   []byte
		expOK bool
	}{{
		desc: "Empty",
	}, {
		desc: "Empty with space",
		in: []byte("  	"),
	}, {
		desc: "Digit at start",
		in:   []byte("0name"),
	}, {
		desc:  "Digit at end",
		in:    []byte("name0"),
		exp:   []byte("name0"),
		expOK: true,
	}, {
		desc:  "Digit at middle",
		in:    []byte("na0me"),
		exp:   []byte("na0me"),
		expOK: true,
	}, {
		desc: "Hyphen at start",
		in:   []byte("-name"),
	}, {
		desc:  "Hyphen at end",
		in:    []byte("name-"),
		exp:   []byte("name-"),
		expOK: true,
	}, {
		desc:  "hyphen at middle",
		in:    []byte("na-me"),
		exp:   []byte("na-me"),
		expOK: true,
	}, {
		desc: "Non alnumhyp at start",
		in:   []byte("!name"),
	}, {
		desc: "Non alnumhyp at end",
		in:   []byte("name!"),
	}, {
		desc: "Non alnumhyp at middle",
		in:   []byte("na!me"),
	}, {
		desc: "With escaped char \\",
		in:   []byte(`na\me`),
	}}

	reader := &Reader{}

	for _, c := range cases {
		t.Log(c.desc)

		got, ok := reader.parseVarName(c.in)
		if !ok {
			test.Assert(t, c.expOK, ok, true)
		}

		test.Assert(t, c.exp, got, true)
	}
}

func TestParseVarValue(t *testing.T) {
	cases := []struct {
		desc   string
		in     []byte
		expval []byte
		expcom []byte
		expok  bool
	}{{
		desc:   `Empty input`,
		expval: varValueTrue,
		expok:  true,
	}, {
		desc:   `Input with spaces`,
		in:     []byte(`   `),
		expval: varValueTrue,
		expok:  true,
	}, {
		desc:   `Double quoted with spaces`,
		in:     []byte(`"   "`),
		expval: []byte(`   `),
		expok:  true,
	}, {
		desc:  `Double quote at start only`,
		in:    []byte(`"\\ value`),
		expok: false,
	}, {
		desc:  `Double quote at end only`,
		in:    []byte(`\\ value "`),
		expok: false,
	}, {
		desc:   `Double quoted at start only`,
		in:     []byte(`"\\" value`),
		expval: []byte(`\ value`),
		expok:  true,
	}, {
		desc:   `Double quoted at end only`,
		in:     []byte(`value "\""`),
		expval: []byte(`value "`),
		expok:  true,
	}, {
		desc:   `Double quoted at start and end`,
		in:     []byte(`"\\" value "\""`),
		expval: []byte(`\ value "`),
		expok:  true,
	}, {
		desc:   `With comment #`,
		in:     []byte(`value # comment`),
		expval: []byte(`value`),
		expcom: []byte(` # comment`),
		expok:  true,
	}, {
		desc:   `With comment ;`,
		in:     []byte(`value ; comment`),
		expval: []byte(`value`),
		expcom: []byte(` ; comment`),
		expok:  true,
	}, {
		desc:   `With comment # inside double-quote`,
		in:     []byte(`"value # comment"`),
		expval: []byte(`value # comment`),
		expok:  true,
	}, {
		desc:   `With comment ; inside double-quote`,
		in:     []byte(`"value ; comment"`),
		expval: []byte(`value ; comment`),
		expok:  true,
	}, {
		desc:   `Double quote and comment #1`,
		in:     []byte(`val" "#ue`),
		expval: []byte(`val `),
		expcom: []byte(`#ue`),
		expok:  true,
	}, {
		desc:   `Double quote and comment #2`,
		in:     []byte(`val" " #ue`),
		expval: []byte(`val `),
		expcom: []byte(` #ue`),
		expok:  true,
	}, {
		desc:   `Double quote and comment #3`,
		in:     []byte(`val " " #ue`),
		expval: []byte(`val  `),
		expcom: []byte(` #ue`),
		expok:  true,
	}, {
		desc:   `Escaped chars`,
		in:     []byte(`value \"escaped\" here`),
		expval: []byte(`value "escaped" here`),
		expok:  true,
	}}

	reader := &Reader{}

	for _, c := range cases {
		t.Log(c.desc)

		gotval, gotcom, ok := reader.parseVarValue(c.in)
		if !ok {
			test.Assert(t, c.expok, ok, true)
		}

		test.Assert(t, c.expval, gotval, true)
		test.Assert(t, c.expcom, gotcom, true)
	}

}
