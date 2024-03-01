// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"io"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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
		desc      string
		in        string
		expErr    error
		expFormat string
		expKey    string
		expValue  string

		expMode lineMode
	}{{
		desc:   `Empty`,
		expErr: errVarNameInvalid,
	}, {
		desc:   `Empty with space`,
		in:     `  	`,
		expErr: errVarNameInvalid,
	}, {
		desc:   `Digit at start`,
		in:     `0name`,
		expErr: errVarNameInvalid,
	}, {
		desc:      `Digit at end`,
		in:        `name0`,
		expErr:    io.EOF,
		expMode:   lineModeKeyOnly,
		expFormat: `%s`,
		expKey:    `name0`,
	}, {
		desc:      `Digit at middle`,
		in:        `na0me`,
		expErr:    io.EOF,
		expMode:   lineModeKeyOnly,
		expFormat: `%s`,
		expKey:    `na0me`,
	}, {
		desc:   `Hyphen at start`,
		in:     `-name`,
		expErr: errVarNameInvalid,
	}, {
		desc:      `Hyphen at end`,
		in:        `name-`,
		expErr:    io.EOF,
		expMode:   lineModeKeyOnly,
		expFormat: `%s`,
		expKey:    `name-`,
	}, {
		desc:      `Gyphen at middle`,
		in:        `na-me`,
		expErr:    io.EOF,
		expMode:   lineModeKeyOnly,
		expFormat: `%s`,
		expKey:    `na-me`,
	}, {
		desc:   `Invalid chart at start`,
		in:     `!name`,
		expErr: errVarNameInvalid,
	}, {
		desc:   `Invalid chart at end`,
		in:     `name!`,
		expErr: errVarNameInvalid,
	}, {
		desc:   `Invalid char at middle`,
		in:     `na!me`,
		expErr: errVarNameInvalid,
	}, {
		desc:   `With escaped char \\`,
		in:     `na\me`,
		expErr: errVarNameInvalid,
	}, {
		desc:      `Without space before comment`,
		in:        `name; comment`,
		expErr:    io.EOF,
		expMode:   lineModeKeyOnly,
		expKey:    `name`,
		expFormat: `%s; comment`,
	}, {
		desc:      `With space before comment`,
		in:        `name ; comment`,
		expErr:    io.EOF,
		expMode:   lineModeKeyOnly,
		expKey:    `name`,
		expFormat: `%s ; comment`,
	}, {
		desc:      `With empty value #1`,
		in:        `name=`,
		expErr:    io.EOF,
		expMode:   lineModeKeyValue,
		expKey:    `name`,
		expFormat: `%s=%s`,
	}, {
		desc:      `With empty value #2`,
		in:        `name =`,
		expErr:    io.EOF,
		expMode:   lineModeKeyValue,
		expKey:    `name`,
		expFormat: `%s =%s`,
	}, {
		desc:      `With empty value and comment`,
		in:        `name = # a comment`,
		expErr:    io.EOF,
		expMode:   lineModeKeyValue,
		expKey:    `name`,
		expFormat: `%s =%s# a comment`,
	}, {
		desc:      `With empty value #3`,
		in:        `name     `,
		expErr:    io.EOF,
		expMode:   lineModeKeyOnly,
		expKey:    `name`,
		expFormat: `%s     `,
	}, {
		desc:      `With newline`,
		in:        "name \n",
		expErr:    io.EOF,
		expMode:   lineModeKeyOnly,
		expKey:    `name`,
		expFormat: "%s \n",
	}, {
		desc:   `With space in the middle`,
		in:     `name 1`,
		expErr: errVarNameInvalid,
	}, {
		desc:      `With dot`,
		in:        `name.subname`,
		expMode:   lineModeKeyOnly,
		expErr:    io.EOF,
		expKey:    `name.subname`,
		expFormat: `%s`,
	}, {
		desc:      `With underscore char`,
		in:        `name_subname`,
		expMode:   lineModeKeyOnly,
		expErr:    io.EOF,
		expKey:    `name_subname`,
		expFormat: `%s`,
	}}

	reader := newReader()

	for _, c := range cases {
		t.Log(c.desc)

		reader.reset([]byte(c.in))

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
	type testCase struct {
		expErr      error
		desc        string
		in          string
		expFormat   string
		expValue    string
		expRawValue string
	}

	var cases = []testCase{{
		desc:      `Empty input`,
		expErr:    io.EOF,
		expFormat: `%s`,
	}, {
		desc:        `Input with spaces`,
		in:          `   `,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expRawValue: `   `,
	}, {
		desc:        `Input with tab`,
		in:          `	`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expRawValue: `	`,
	}, {
		desc:        `Input with newline`,
		in:          "\n",
		expErr:      nil,
		expFormat:   "%s\n",
		expRawValue: "",
	}, {
		desc:        `Double quoted with spaces`,
		in:          `"   "`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    `   `,
		expRawValue: `"   "`,
	}, {
		desc:   `Double quote at start only`,
		in:     `"\\ value`,
		expErr: errValueInvalid,
	}, {
		desc:   `Double quote at end only`,
		in:     `\\ value "`,
		expErr: errValueInvalid,
	}, {
		desc:        `Double quoted at start`,
		in:          `"\\" value`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    `\ value`,
		expRawValue: `"\\" value`,
	}, {
		desc:        `Double quoted at end only`,
		in:          `value "\""`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    `value "`,
		expRawValue: `value "\""`,
	}, {
		desc:        `Double quoted at start and end`,
		in:          `"\\" value "\""`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    `\ value "`,
		expRawValue: `"\\" value "\""`,
	}, {
		desc:        `With comment #`,
		in:          `value # comment`,
		expErr:      io.EOF,
		expFormat:   `%s# comment`,
		expValue:    `value`,
		expRawValue: `value `,
	}, {
		desc:        `With comment ;`,
		in:          `value ; comment`,
		expErr:      io.EOF,
		expFormat:   `%s; comment`,
		expValue:    `value`,
		expRawValue: `value `,
	}, {
		desc:        `With comment # inside double-quote`,
		in:          `"value # comment"`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    `value # comment`,
		expRawValue: `"value # comment"`,
	}, {
		desc:        `With comment ; inside double-quote`,
		in:          `"value ; comment"`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    `value ; comment`,
		expRawValue: `"value ; comment"`,
	}, {
		desc:        `Double quote and comment #1`,
		in:          `val" "#ue`,
		expErr:      io.EOF,
		expFormat:   `%s#ue`,
		expValue:    `val `,
		expRawValue: `val" "`,
	}, {
		desc:        `Double quote and comment #2`,
		in:          `val" " #ue`,
		expErr:      io.EOF,
		expFormat:   `%s#ue`,
		expValue:    `val `,
		expRawValue: `val" " `,
	}, {
		desc:        `Double quote and comment #3`,
		in:          `val " " #ue`,
		expErr:      io.EOF,
		expFormat:   `%s#ue`,
		expValue:    `val  `,
		expRawValue: `val " " `,
	}, {
		desc:        `Escaped chars #1`,
		in:          `value \"escaped\" here`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    `value "escaped" here`,
		expRawValue: `value \"escaped\" here`,
	}, {
		desc:        `Escaped chars #2`,
		in:          `"value\b\n\t\"escaped\" here"`,
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    "value\b\n\t\"escaped\" here",
		expRawValue: `"value\b\n\t\"escaped\" here"`,
	}, {
		desc:   `Invalid escaped chars`,
		in:     `"value\b\n\x\"escaped\" here"`,
		expErr: errValueInvalid,
	}, {
		desc:        `Multiline no space`,
		in:          "multi\\\nvalue",
		expErr:      io.EOF,
		expFormat:   `%s`,
		expValue:    `multivalue`,
		expRawValue: "multi\\\nvalue",
	}}

	var (
		reader = newReader()

		c   testCase
		err error
	)

	for _, c = range cases {
		t.Log(c.desc)

		reader.reset([]byte(c.in))

		err = reader.parseVarValue()
		if err != nil {
			test.Assert(t, "error", c.expErr, err)
			if err != io.EOF {
				continue
			}
		}

		test.Assert(t, "raw value", c.expRawValue, string(reader._var.rawValue))
		test.Assert(t, "value", c.expValue, reader._var.value)
		test.Assert(t, "format", c.expFormat, reader._var.format)
	}
}

func TestParseRawValue(t *testing.T) {
	type testCase struct {
		in  string
		exp string
	}

	var cases = []testCase{{
		in:  "\\\n \ta",
		exp: `a`,
	}, {
		in:  " a\\\n\t b\\\n \tc",
		exp: `a b c`,
	}, {
		in:  " a\\\n \"\\\" b \"\\\n c",
		exp: `a " b  c`,
	}}

	var (
		c   testCase
		got string
	)

	for _, c = range cases {
		got = parseRawValue([]byte(c.in))
		test.Assert(t, "parseRawValue", c.exp, got)
	}
}
