// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
		expMode    lineMode
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
		expMode:    lineModeSection,
		expSecName: "section-.",
		expFormat:  "[%s]",
	}, {
		desc:       "With valid name and comment",
		in:         `[section-.] ; a comment`,
		expErr:     io.EOF,
		expMode:    lineModeSection | lineModeComment,
		expSecName: "section-.",
		expFormat:  "[%s] %s",
		expComment: "; a comment",
	}, {
		desc:       "With valid name and sub",
		in:         `[section-. "su\bsec\tio\n"]`,
		expErr:     io.EOF,
		expMode:    lineModeSection | lineModeSubsection,
		expSecName: "section-.",
		expSubName: "subsection",
		expFormat:  `[%s "%s"]`,
	}, {
		desc:       "With valid name, sub, and comment",
		in:         `[section-. "su\bsec\tio\n"]   # comment`,
		expErr:     io.EOF,
		expMode:    lineModeSection | lineModeSubsection | lineModeComment,
		expSecName: "section-.",
		expSubName: "subsection",
		expFormat:  `[%s "%s"]   %s`,
		expComment: "# comment",
	}}

	reader := newReader()

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
		test.Assert(t, "format", c.expFormat, reader._var.format, true)
		test.Assert(t, "section", c.expSecName, reader._var.secName, true)
		test.Assert(t, "subsection", c.expSubName, reader._var.subName, true)
		test.Assert(t, "comment", c.expComment, reader._var.others, true)
	}
}

func TestParseSubsection(t *testing.T) {
	cases := []struct {
		desc       string
		in         string
		expErr     error
		expMode    lineMode
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
		expMode:   lineModeSubsection,
		expFormat: `"%s"]`,
		expSub:    ` `,
	}, {
		desc:      "With valid subsection",
		in:        `"subsection\""]`,
		expErr:    io.EOF,
		expMode:   lineModeSubsection,
		expFormat: `"%s"]`,
		expSub:    `subsection"`,
	}}

	reader := newReader()

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
		test.Assert(t, "format", c.expFormat, reader._var.format, true)
		test.Assert(t, "subsection", c.expSub, reader._var.subName, true)
		test.Assert(t, "comment", c.expComment, reader._var.others, true)
	}
}

func TestParseVariable(t *testing.T) {
	cases := []struct {
		desc       string
		in         []byte
		expErr     error
		expMode    lineMode
		expFormat  string
		expComment string
		expKey     string
		expValue   string
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
		expMode:   lineModeValue,
		expFormat: "%s",
		expKey:    "name0",
	}, {
		desc:      "Digit at middle",
		in:        []byte("na0me"),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expFormat: "%s",
		expKey:    "na0me",
	}, {
		desc:   "Hyphen at start",
		in:     []byte("-name"),
		expErr: errVarNameInvalid,
	}, {
		desc:      "Hyphen at end",
		in:        []byte("name-"),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expFormat: "%s",
		expKey:    "name-",
	}, {
		desc:      "hyphen at middle",
		in:        []byte("na-me"),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expFormat: "%s",
		expKey:    "na-me",
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
		expMode:    lineModeValue | lineModeComment,
		expKey:     "name",
		expComment: "; comment",
		expFormat:  "%s%s",
	}, {
		desc:       "With comment #2",
		in:         []byte(`name ; comment`),
		expErr:     io.EOF,
		expMode:    lineModeValue | lineModeComment,
		expKey:     "name",
		expComment: "; comment",
		expFormat:  "%s %s",
	}, {
		desc:      "With empty value #1",
		in:        []byte(`name=`),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expKey:    "name",
		expFormat: "%s=",
		expValue:  "",
	}, {
		desc:      "With empty value #2",
		in:        []byte(`name =`),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expKey:    "name",
		expFormat: "%s =",
		expValue:  "",
	}, {
		desc:       "With empty value and comment",
		in:         []byte(`name = # a comment`),
		expErr:     io.EOF,
		expMode:    lineModeValue | lineModeComment,
		expKey:     "name",
		expFormat:  "%s = %s%s",
		expComment: "# a comment",
		expValue:   "",
	}, {
		desc:      "With empty value #3",
		in:        []byte(`name     `),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expKey:    "name",
		expFormat: "%s     ",
	}, {
		desc: "With newline",
		in: []byte(`name 
`),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expKey:    "name",
		expFormat: "%s \n",
	}, {
		desc:   "With invalid char",
		in:     []byte(`name 1`),
		expErr: errVarNameInvalid,
	}, {
		desc:      "With dot",
		in:        []byte(`name.subname`),
		expMode:   lineModeValue,
		expErr:    io.EOF,
		expKey:    "name.subname",
		expFormat: "%s",
	}, {
		desc:      "With underscore char",
		in:        []byte(`name_subname`),
		expMode:   lineModeValue,
		expErr:    io.EOF,
		expKey:    "name_subname",
		expFormat: "%s",
	}}

	reader := newReader()

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
		expFormat  string
		expValue   string
		expComment string
	}{{
		desc:     `Empty input`,
		expErr:   io.EOF,
		expValue: "",
	}, {
		desc:      `Input with spaces`,
		in:        []byte(`   `),
		expErr:    io.EOF,
		expFormat: `   `,
		expValue:  "",
	}, {
		desc: `Input with tab`,
		in: []byte(`	`),
		expErr: io.EOF,
		expFormat: `	`,
		expValue: "",
	}, {
		desc: `Input with newline`,
		in: []byte(`
`),
		expErr: nil,
		expFormat: `
`,
		expValue: "",
	}, {
		desc:      `Double quoted with spaces`,
		in:        []byte(`"   "`),
		expErr:    io.EOF,
		expFormat: `%s`,
		expValue:  "   ",
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
		expFormat: `%s`,
		expValue:  `\ value`,
	}, {
		desc:      `Double quoted at end only`,
		in:        []byte(`value "\""`),
		expErr:    io.EOF,
		expFormat: `%s`,
		expValue:  `value "`,
	}, {
		desc:      `Double quoted at start and end`,
		in:        []byte(`"\\" value "\""`),
		expErr:    io.EOF,
		expFormat: `%s`,
		expValue:  `\ value "`,
	}, {
		desc:       `With comment #`,
		in:         []byte(`value # comment`),
		expErr:     io.EOF,
		expFormat:  `%s %s`,
		expValue:   "value",
		expComment: "# comment",
	}, {
		desc:       `With comment ;`,
		in:         []byte(`value ; comment`),
		expErr:     io.EOF,
		expFormat:  "%s %s",
		expValue:   "value",
		expComment: "; comment",
	}, {
		desc:      `With comment # inside double-quote`,
		in:        []byte(`"value # comment"`),
		expErr:    io.EOF,
		expFormat: `%s`,
		expValue:  `value # comment`,
	}, {
		desc:      `With comment ; inside double-quote`,
		in:        []byte(`"value ; comment"`),
		expErr:    io.EOF,
		expFormat: `%s`,
		expValue:  `value ; comment`,
	}, {
		desc:       `Double quote and comment #1`,
		in:         []byte(`val" "#ue`),
		expErr:     io.EOF,
		expFormat:  `%s%s`,
		expValue:   `val `,
		expComment: `#ue`,
	}, {
		desc:       `Double quote and comment #2`,
		in:         []byte(`val" " #ue`),
		expErr:     io.EOF,
		expFormat:  `%s %s`,
		expValue:   `val `,
		expComment: `#ue`,
	}, {
		desc:       `Double quote and comment #3`,
		in:         []byte(`val " " #ue`),
		expErr:     io.EOF,
		expFormat:  `%s %s`,
		expValue:   `val  `,
		expComment: `#ue`,
	}, {
		desc:      `Escaped chars #1`,
		in:        []byte(`value \"escaped\" here`),
		expErr:    io.EOF,
		expFormat: `%s`,
		expValue:  `value "escaped" here`,
	}, {
		desc:      `Escaped chars #2`,
		in:        []byte(`"value\b\n\t\"escaped\" here"`),
		expErr:    io.EOF,
		expFormat: `%s`,
		expValue:  "value\b\n\t\"escaped\" here",
	}, {
		desc:   `Invalid escaped chars`,
		in:     []byte(`"value\b\n\x\"escaped\" here"`),
		expErr: errValueInvalid,
	}}

	reader := newReader()

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
