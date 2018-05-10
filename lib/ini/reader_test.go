package ini

import (
	"io"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseSectionHeader(t *testing.T) {
	cases := []struct {
		desc       string
		in         string
		expErr     error
		expMode    varMode
		expFormat  string
		expSecName string
		expSubName string
		expComment string
	}{{
		desc:   "With empty input",
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #0 (no name)",
		in:     `[`,
		expErr: errBadConfig,
	}, {

		desc:   "With invalid section #1 (no section start)",
		in:     `nosection start]`,
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #2 (with leading space)",
		in:     `[ section]`,
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #3 (not closed]",
		in:     `[section`,
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #4 (no sub)",
		in:     `[section  ]`,
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #5 (invalid char)",
		in:     `[section!]`,
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #6 (trailing char)",
		in:     `[section] not`,
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #7 (trailing char)",
		in:     `[section "subsection"]    not`,
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #8 (trailing space)",
		in:     `[section "subsection" ] # comment`,
		expErr: errBadConfig,
	}, {
		desc:   "With invalid section #9 (not closed)",
		in:     `[section "subsection" # comment`,
		expErr: errBadConfig,
	}, {
		desc:       "With valid name",
		in:         `[section-.]`,
		expErr:     io.EOF,
		expMode:    varModeSection,
		expSecName: "section-.",
		expFormat:  "[%s]",
	}, {
		desc:       "With valid name and comment",
		in:         `[section-.] ; a comment`,
		expErr:     io.EOF,
		expMode:    varModeSection | varModeComment,
		expSecName: "section-.",
		expFormat:  "[%s] %s",
		expComment: "; a comment",
	}, {
		desc:       "With valid name and sub",
		in:         `[section-. "su\bsec\tio\n"]`,
		expErr:     io.EOF,
		expMode:    varModeSection | varModeSubsection,
		expSecName: "section-.",
		expSubName: "subsection",
		expFormat:  `[%s "%s"]`,
	}, {
		desc:       "With valid name, sub, and comment",
		in:         `[section-. "su\bsec\tio\n"]   # comment`,
		expErr:     io.EOF,
		expMode:    varModeSection | varModeSubsection | varModeComment,
		expSecName: "section-.",
		expSubName: "subsection",
		expFormat:  `[%s "%s"]   %s`,
		expComment: "# comment",
	}}

	reader := NewReader()

	for _, c := range cases {
		t.Log(c.desc)

		reader.reset([]byte(c.in))

		err := reader.parseSectionHeader()
		if err != nil {
			test.Assert(t, "error", c.expErr, err, true)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "mode", c.expMode, reader._var.mode, true)
		test.Assert(t, "format", c.expFormat, string(reader._var.format), true)
		test.Assert(t, "section", c.expSecName, string(reader._var.secName), true)
		test.Assert(t, "subsection", c.expSubName, string(reader._var.subName), true)
		test.Assert(t, "comment", c.expComment, string(reader._var.others), true)
	}

}

func TestParseSubsection(t *testing.T) {
	cases := []struct {
		desc       string
		in         string
		expErr     error
		expMode    varMode
		expFormat  string
		expSub     string
		expComment string
	}{{
		desc:   "With empty input",
		expErr: errBadConfig,
	}, {
		desc:   "With invalid format",
		in:     `" ]`,
		expErr: errBadConfig,
	}, {
		desc:      "With leading space",
		in:        `" "]`,
		expErr:    io.EOF,
		expMode:   varModeSubsection,
		expFormat: `"%s"]`,
		expSub:    ` `,
	}, {
		desc:      "With valid subsection",
		in:        `"subsection\""]`,
		expErr:    io.EOF,
		expMode:   varModeSubsection,
		expFormat: `"%s"]`,
		expSub:    `subsection"`,
	}}

	reader := NewReader()

	for _, c := range cases {
		t.Log(c.desc)

		reader.reset([]byte(c.in))

		err := reader.parseSubsection()
		if err != nil {
			test.Assert(t, "error", c.expErr, err, true)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "mode", c.expMode, reader._var.mode, true)
		test.Assert(t, "format", c.expFormat, string(reader._var.format), true)
		test.Assert(t, "subsection", c.expSub, string(reader._var.subName), true)
		test.Assert(t, "comment", c.expComment, string(reader._var.others), true)
	}
}

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
	}, {
		desc:      "With empty value #3",
		in:        []byte(`name     `),
		expErr:    io.EOF,
		expMode:   varModeSingle,
		expKey:    []byte("name"),
		expFormat: []byte("%s     "),
		expValue:  varValueTrue,
	}, {
		desc: "With newline",
		in: []byte(`name 
`),
		expErr:    io.EOF,
		expMode:   varModeSingle,
		expKey:    []byte("name"),
		expFormat: []byte("%s \n"),
		expValue:  varValueTrue,
	}, {
		desc:   "With invalid char",
		in:     []byte(`name 1`),
		expErr: errVarNameInvalid,
	}}

	reader := NewReader()

	for _, c := range cases {
		t.Log(c.desc)

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
		desc:      `Escaped chars #1`,
		in:        []byte(`value \"escaped\" here`),
		expErr:    io.EOF,
		expFormat: []byte(`value \"escaped\" here`),
		expValue:  []byte(`value "escaped" here`),
	}, {
		desc:      `Escaped chars #2`,
		in:        []byte(`"value\b\n\t\"escaped\" here"`),
		expErr:    io.EOF,
		expFormat: []byte(`"value\b\n\t\"escaped\" here"`),
		expValue:  []byte("value\b\n\t\"escaped\" here"),
	}, {
		desc:   `Invalid escaped chars`,
		in:     []byte(`"value\b\n\x\"escaped\" here"`),
		expErr: errValueInvalid,
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
