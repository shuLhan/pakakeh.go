package ini

import (
	"io"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseVariable(t *testing.T) {
	cases := []struct {
		desc       string
		in         []byte
		expErr     error
		expMode    varMode
		expFormat  []byte
		expComment []byte
		expKey     []byte
		expValue   []byte
	}{{
		desc:   "Empty",
		expErr: errVarNameInvalid,
	}, {
		desc: "Empty with space",
		in: []byte("  	"),
		expErr: errVarNameInvalid,
	}, {
		desc:   "Digit at start",
		in:     []byte("0name"),
		expErr: errVarNameInvalid,
	}, {
		desc:      "Digit at end",
		in:        []byte("name0"),
		expErr:    io.EOF,
		expMode:   varModeSingle,
		expFormat: []byte("%s"),
		expKey:    []byte("name0"),
		expValue:  varValueTrue,
	}, {
		desc:      "Digit at middle",
		in:        []byte("na0me"),
		expErr:    io.EOF,
		expMode:   varModeSingle,
		expFormat: []byte("%s"),
		expKey:    []byte("na0me"),
		expValue:  varValueTrue,
	}, {
		desc:   "Hyphen at start",
		in:     []byte("-name"),
		expErr: errVarNameInvalid,
	}, {
		desc:      "Hyphen at end",
		in:        []byte("name-"),
		expErr:    io.EOF,
		expMode:   varModeSingle,
		expFormat: []byte("%s"),
		expKey:    []byte("name-"),
		expValue:  varValueTrue,
	}, {
		desc:      "hyphen at middle",
		in:        []byte("na-me"),
		expErr:    io.EOF,
		expMode:   varModeSingle,
		expFormat: []byte("%s"),
		expKey:    []byte("na-me"),
		expValue:  varValueTrue,
	}, {
		desc:   "Non alnumhyp at start",
		in:     []byte("!name"),
		expErr: errVarNameInvalid,
	}, {
		desc:   "Non alnumhyp at end",
		in:     []byte("name!"),
		expErr: errVarNameInvalid,
	}, {
		desc:   "Non alnumhyp at middle",
		in:     []byte("na!me"),
		expErr: errVarNameInvalid,
	}, {
		desc:   "With escaped char \\",
		in:     []byte(`na\me`),
		expErr: errVarNameInvalid,
	}, {
		desc:       "With comment #1",
		in:         []byte(`name; comment`),
		expErr:     io.EOF,
		expMode:    varModeSingle | varModeComment,
		expKey:     []byte("name"),
		expComment: []byte("; comment"),
		expFormat:  []byte("%s%s"),
		expValue:   varValueTrue,
	}, {
		desc:       "With comment #2",
		in:         []byte(`name ; comment`),
		expErr:     io.EOF,
		expMode:    varModeSingle | varModeComment,
		expKey:     []byte("name"),
		expComment: []byte("; comment"),
		expFormat:  []byte("%s %s"),
		expValue:   varValueTrue,
	}, {
		desc:      "With empty value #1",
		in:        []byte(`name=`),
		expErr:    io.EOF,
		expMode:   varModeSingle,
		expKey:    []byte("name"),
		expFormat: []byte("%s="),
		expValue:  varValueTrue,
	}, {
		desc:      "With empty value #2",
		in:        []byte(`name =`),
		expErr:    io.EOF,
		expMode:   varModeSingle,
		expKey:    []byte("name"),
		expFormat: []byte("%s ="),
		expValue:  varValueTrue,
	}, {
		desc:       "With empty value and comment",
		in:         []byte(`name =# a comment`),
		expErr:     io.EOF,
		expMode:    varModeSingle | varModeComment,
		expKey:     []byte("name"),
		expFormat:  []byte("%s =%s"),
		expComment: []byte("# a comment"),
		expValue:   varValueTrue,
	}}

	reader := NewReader()

	for _, c := range cases {
		t.Log(c)
		reader.reset(c.in)

		err := reader.parseVariable()
		if err != nil {
			test.Assert(t, "error", c.expErr, err, true)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "mode", c.expMode, reader._var.mode, true)
		test.Assert(t, "format", c.expFormat, reader._var.format, true)
		test.Assert(t, "key", c.expKey, reader._var.key, true)
		test.Assert(t, "value", c.expValue, reader._var.value, true)
		test.Assert(t, "comment", c.expComment, reader._var.others, true)
	}
}

func TestParseVarValue(t *testing.T) {
	cases := []struct {
		desc       string
		in         []byte
		expErr     error
		expFormat  []byte
		expValue   []byte
		expComment []byte
	}{{
		desc:     `Empty input`,
		expErr:   io.EOF,
		expValue: varValueTrue,
	}, {
		desc:      `Input with spaces`,
		in:        []byte(`   `),
		expErr:    io.EOF,
		expFormat: []byte(`   `),
		expValue:  varValueTrue,
	}, {
		desc: `Input with tab`,
		in: []byte(`	`),
		expErr: io.EOF,
		expFormat: []byte(`	`),
		expValue: varValueTrue,
	}, {
		desc: `Input with newline`,
		in: []byte(`
`),
		expErr: nil,
		expFormat: []byte(`
`),
		expValue: varValueTrue,
	}, {
		desc:      `Double quoted with spaces`,
		in:        []byte(`"   "`),
		expErr:    io.EOF,
		expFormat: []byte(`"   "`),
		expValue:  []byte("   "),
	}, {
		desc:   `Double quote at start only`,
		in:     []byte(`"\\ value`),
		expErr: errValueInvalid,
	}, {
		desc:   `Double quote at end only`,
		in:     []byte(`\\ value "`),
		expErr: errValueInvalid,
	}, {
		desc:      `Double quoted at start only`,
		in:        []byte(`"\\" value`),
		expErr:    io.EOF,
		expFormat: []byte(`"\\" value`),
		expValue:  []byte(`\ value`),
	}, {
		desc:      `Double quoted at end only`,
		in:        []byte(`value "\""`),
		expErr:    io.EOF,
		expFormat: []byte(`value "\""`),
		expValue:  []byte(`value "`),
	}, {
		desc:      `Double quoted at start and end`,
		in:        []byte(`"\\" value "\""`),
		expErr:    io.EOF,
		expFormat: []byte(`"\\" value "\""`),
		expValue:  []byte(`\ value "`),
	}, {
		desc:       `With comment #`,
		in:         []byte(`value # comment`),
		expErr:     io.EOF,
		expFormat:  []byte(`value %s`),
		expValue:   []byte("value"),
		expComment: []byte("# comment"),
	}, {
		desc:       `With comment ;`,
		in:         []byte(`value ; comment`),
		expErr:     io.EOF,
		expFormat:  []byte("value %s"),
		expValue:   []byte("value"),
		expComment: []byte("; comment"),
	}, {
		desc:      `With comment # inside double-quote`,
		in:        []byte(`"value # comment"`),
		expErr:    io.EOF,
		expFormat: []byte(`"value # comment"`),
		expValue:  []byte(`value # comment`),
	}, {
		desc:      `With comment ; inside double-quote`,
		in:        []byte(`"value ; comment"`),
		expErr:    io.EOF,
		expFormat: []byte(`"value ; comment"`),
		expValue:  []byte(`value ; comment`),
	}, {
		desc:       `Double quote and comment #1`,
		in:         []byte(`val" "#ue`),
		expErr:     io.EOF,
		expFormat:  []byte(`val" "%s`),
		expValue:   []byte(`val `),
		expComment: []byte(`#ue`),
	}, {
		desc:       `Double quote and comment #2`,
		in:         []byte(`val" " #ue`),
		expErr:     io.EOF,
		expFormat:  []byte(`val" " %s`),
		expValue:   []byte(`val `),
		expComment: []byte(`#ue`),
	}, {
		desc:       `Double quote and comment #3`,
		in:         []byte(`val " " #ue`),
		expErr:     io.EOF,
		expFormat:  []byte(`val " " %s`),
		expValue:   []byte(`val  `),
		expComment: []byte(`#ue`),
	}, {
		desc:      `Escaped chars`,
		in:        []byte(`value \"escaped\" here`),
		expErr:    io.EOF,
		expFormat: []byte(`value \"escaped\" here`),
		expValue:  []byte(`value "escaped" here`),
	}}

	reader := NewReader()
	for _, c := range cases {
		t.Log(c.desc)
		reader.reset(c.in)

		err := reader.parseVarValue()
		if err != nil {
			test.Assert(t, "error", c.expErr, err, true)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "format", c.expFormat, reader._var.format, true)
		test.Assert(t, "value", c.expValue, reader._var.value, true)
		test.Assert(t, "comment", c.expComment, reader._var.others, true)
	}

}
