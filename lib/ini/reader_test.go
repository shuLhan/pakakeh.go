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
		expErr     error
		expFormat  string
		expSecName string
		expSubName string

		desc    string
		in      string
		expMode lineMode
	}{{
		desc:   `With no section name`,
		in:     ``,
		expErr: errBadConfig,
	}, {
		desc:   `With leading space`,
		in:     ` section]`,
		expErr: errBadConfig,
	}, {
		desc:   `Without closed`,
		in:     `section`,
		expErr: errBadConfig,
	}, {
		desc:   `With trailing space, no sub`,
		in:     `section  ]`,
		expErr: errBadConfig,
	}, {
		desc:   `With invalid char`,
		in:     `section!]`,
		expErr: errBadConfig,
	}, {
		desc:   `With trailing string`,
		in:     `section] not`,
		expErr: errBadConfig,
	}, {
		desc:   `With sub and trailing string`,
		in:     `section "subsection"]    not`,
		expErr: errBadConfig,
	}, {
		desc:   `With sub and trailing space`,
		in:     `section "subsection" ] # comment`,
		expErr: errBadConfig,
	}, {
		desc:   `With subsction and not closed`,
		in:     `section "subsection" # comment`,
		expErr: errBadConfig,
	}, {
		desc:       `With valid name`,
		in:         `section-.]`,
		expErr:     io.EOF,
		expMode:    lineModeSection,
		expSecName: `section-.`,
		expFormat:  `[%s]`,
	}, {
		desc:       `With valid name and comment`,
		in:         `section-.] ; a comment`,
		expErr:     io.EOF,
		expMode:    lineModeSection,
		expSecName: `section-.`,
		expFormat:  `[%s] ; a comment`,
	}, {
		desc:       `With valid name and sub`,
		in:         `section-. "su\bsec\tio\n"]`,
		expErr:     io.EOF,
		expMode:    lineModeSection | lineModeSubsection,
		expSecName: `section-.`,
		expSubName: `subsection`,
		expFormat:  `[%s "%s"]`,
	}, {
		desc:       `With valid name, sub, and comment`,
		in:         `section-. "su\bsec\tio\n"]   # comment`,
		expErr:     io.EOF,
		expMode:    lineModeSection | lineModeSubsection,
		expSecName: `section-.`,
		expSubName: `subsection`,
		expFormat:  `[%s "%s"]   # comment`,
	}}

	reader := newReader()

	for _, c := range cases {
		t.Log(c.desc)

		reader.reset([]byte(c.in))

		err := reader.parseSectionHeader()
		if err != nil {
			test.Assert(t, "error", c.expErr, err)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "mode", c.expMode, reader._var.mode)
		test.Assert(t, "format", c.expFormat, reader._var.format)
		test.Assert(t, "section", c.expSecName, reader._var.secName)
		test.Assert(t, "subsection", c.expSubName, reader._var.subName)
	}
}

func TestParseSubsection(t *testing.T) {
	cases := []struct {
		expErr    error
		expFormat string
		expSub    string

		desc string
		in   string

		expMode lineMode
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
			test.Assert(t, "error", c.expErr, err)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "mode", c.expMode, reader._var.mode)
		test.Assert(t, "format", c.expFormat, reader._var.format)
		test.Assert(t, "subsection", c.expSub, reader._var.subName)
	}
}

func TestParseVariable(t *testing.T) {
	cases := []struct {
		expErr    error
		expFormat string
		expKey    string
		expValue  string

		desc    string
		in      []byte
		expMode lineMode
	}{{
		desc:   "Empty",
		expErr: errVarNameInvalid,
	}, {
		desc:   "Empty with space",
		in:     []byte("  	"),
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
		desc:      `Without space before comment`,
		in:        []byte(`name; comment`),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expKey:    `name`,
		expFormat: `%s; comment`,
	}, {
		desc:      `With space before comment`,
		in:        []byte(`name ; comment`),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expKey:    `name`,
		expFormat: `%s ; comment`,
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
		desc:      `With empty value and comment`,
		in:        []byte(`name = # a comment`),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expKey:    `name`,
		expFormat: `%s = %s# a comment`,
		expValue:  ``,
	}, {
		desc:      "With empty value #3",
		in:        []byte(`name     `),
		expErr:    io.EOF,
		expMode:   lineModeValue,
		expKey:    "name",
		expFormat: "%s     ",
	}, {
		desc:      `With newline`,
		in:        []byte("name \n"),
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
			test.Assert(t, "error", c.expErr, err)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "mode", c.expMode, reader._var.mode)
		test.Assert(t, "format", c.expFormat, reader._var.format)
		test.Assert(t, "key", c.expKey, reader._var.key)
		test.Assert(t, "value", c.expValue, reader._var.value)
	}
}

func TestParseVarValue(t *testing.T) {
	cases := []struct {
		expErr    error
		expFormat string
		expValue  string

		desc string
		in   []byte
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
		desc:      `Input with tab`,
		in:        []byte(`	`),
		expErr:    io.EOF,
		expFormat: `	`,
		expValue:  "",
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
		desc:      `With comment #`,
		in:        []byte(`value # comment`),
		expErr:    io.EOF,
		expFormat: `%s # comment`,
		expValue:  `value`,
	}, {
		desc:      `With comment ;`,
		in:        []byte(`value ; comment`),
		expErr:    io.EOF,
		expFormat: `%s ; comment`,
		expValue:  `value`,
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
		desc:      `Double quote and comment #1`,
		in:        []byte(`val" "#ue`),
		expErr:    io.EOF,
		expFormat: `%s#ue`,
		expValue:  `val `,
	}, {
		desc:      `Double quote and comment #2`,
		in:        []byte(`val" " #ue`),
		expErr:    io.EOF,
		expFormat: `%s #ue`,
		expValue:  `val `,
	}, {
		desc:      `Double quote and comment #3`,
		in:        []byte(`val " " #ue`),
		expErr:    io.EOF,
		expFormat: `%s #ue`,
		expValue:  `val  `,
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
			test.Assert(t, "error", c.expErr, err)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "format", c.expFormat, reader._var.format)
		test.Assert(t, "value", c.expValue, reader._var.value)
	}
}
