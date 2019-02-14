// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"testing"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/test"
)

func TestParseField(t *testing.T) {
	longValue := string(libbytes.Random([]byte(libbytes.ASCIILetters), 994))

	cases := []struct {
		desc    string
		raw     []byte
		expErr  string
		exp     *Field
		expRest []byte
	}{{
		desc: "With empty input",
	}, {
		desc:   "With long line",
		raw:    []byte("name:" + longValue + "\r\n"),
		expErr: "ParseField: line greater than 998 characters",
	}, {
		desc:   "With only whitespaces",
		raw:    []byte("  "),
		expErr: "ParseField: invalid input",
	}, {
		desc:   "With only CRLF",
		raw:    []byte("\r\n"),
		expErr: "ParseField: invalid character at index 0",
	}, {
		desc:   "Without separator and CRLF",
		raw:    []byte("name"),
		expErr: "ParseField: invalid input",
	}, {
		desc:   "Without separator",
		raw:    []byte("name\r\n"),
		expErr: "ParseField: invalid character at index 4",
	}, {
		desc:   "With space on name",
		raw:    []byte("na me\r\n"),
		expErr: "ParseField: invalid character at index 3",
	}, {
		desc:   "Without value and CRLF",
		raw:    []byte("name:"),
		expErr: "ParseField: invalid input",
	}, {
		desc:   "Without value and CRLF",
		raw:    []byte("name: "),
		expErr: "ParseField: invalid input",
	}, {
		desc:   "Without value",
		raw:    []byte("name:\r\n"),
		expErr: "ParseField: invalid input",
	}, {
		desc:   "Without value",
		raw:    []byte("name: \r\n"),
		expErr: "ParseField: invalid input",
	}, {
		desc:   "Without CRLF",
		raw:    []byte("name:value"),
		expErr: "ParseField: invalid input",
	}, {
		desc:   "Without CR",
		raw:    []byte("name:value\n"),
		expErr: "ParseField: invalid character at index 10",
	}, {
		desc:   "Without LF",
		raw:    []byte("name:value\r"),
		expErr: "ParseField: invalid input",
	}, {
		desc:   "With CR inside value",
		raw:    []byte("name:valu\re"),
		expErr: "ParseField: invalid character at index 10",
	}, {
		desc: "With valid input",
		raw:  []byte("NAME : VALUE\r\n"),
		exp: &Field{
			Name:     []byte("name"),
			Value:    []byte("VALUE\r\n"),
			oriName:  []byte("NAME "),
			oriValue: []byte(" VALUE\r\n"),
		},
	}, {
		desc: "With single folding",
		raw:  []byte("Name : \r\n \t Value\r\n"),
		exp: &Field{
			Name:     []byte("name"),
			Value:    []byte("Value\r\n"),
			oriName:  []byte("Name "),
			oriValue: []byte(" \r\n \t Value\r\n"),
		},
	}, {
		desc: "With multiple folding between value",
		raw:  []byte("namE : This\r\n is\r\n\ta\r\n \tvalue\r\n"),
		exp: &Field{
			Name:     []byte("name"),
			Value:    []byte("This is a value\r\n"),
			oriName:  []byte("namE "),
			oriValue: []byte(" This\r\n is\r\n\ta\r\n \tvalue\r\n"),
		},
	}, {
		desc: "With multiple fields",
		raw:  []byte("a : 1\r\nb : 2\r\n"),
		exp: &Field{
			Name:     []byte("a"),
			Value:    []byte("1\r\n"),
			oriName:  []byte("a "),
			oriValue: []byte(" 1\r\n"),
		},
		expRest: []byte("b : 2\r\n"),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, rest, err := ParseField(c.raw)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}
		if got == nil {
			test.Assert(t, "Field", c.exp, got, true)
			continue
		}

		test.Assert(t, "Field.oriName", c.exp.oriName, got.oriName, true)
		test.Assert(t, "Field.oriValue", c.exp.oriValue, got.oriValue, true)
		test.Assert(t, "Field.Name", c.exp.Name, got.Name, true)
		test.Assert(t, "Field.Value", c.exp.Value, got.Value, true)

		test.Assert(t, "rest", c.expRest, rest, true)
	}
}

func TestUnpackDate(t *testing.T) {
	cases := []struct {
		desc   string
		value  []byte
		exp    time.Time
		expErr string
	}{{
		desc:   "With empty value",
		expErr: "unpackDate: empty date",
	}, {
		desc:   "With only spaces",
		value:  []byte("  "),
		expErr: "unpackDate: empty date",
	}, {
		desc:   "With invalid date format",
		value:  []byte("Sat"),
		expErr: "unpackDate: invalid date format",
	}, {
		desc:   "With invalid date format",
		value:  []byte("Sat,"),
		expErr: "unpackDate: invalid date format",
	}, {
		desc:   "With missing month",
		value:  []byte("Sat, 2"),
		expErr: "unpackDate: missing month",
	}, {
		desc:   "With missing month",
		value:  []byte("Sat, 2 "),
		expErr: "unpackDate: missing month",
	}, {
		desc:   "With invalid month",
		value:  []byte("Sat, 2 X 2019"),
		expErr: "unpackDate: invalid month: 'X'",
	}, {
		desc:   "With missing year",
		value:  []byte("Sat, 2 Feb"),
		expErr: "unpackDate: invalid year",
	}, {
		desc:   "With invalid year",
		value:  []byte("Sat, 2 Feb 2019"),
		expErr: "unpackDate: invalid year",
	}, {
		desc:   "With invalid hour",
		value:  []byte("Sat, 2 Feb 2019 00"),
		expErr: "unpackDate: invalid hour",
	}, {
		desc:   "With invalid hour",
		value:  []byte("Sat, 2 Feb 2019 24:55:16 +0000"),
		expErr: "unpackDate: invalid hour: 24",
	}, {
		desc:   "With invalid minute",
		value:  []byte("Sat, 2 Feb 2019 00:60:16 +0000"),
		expErr: "unpackDate: invalid minute: 60",
	}, {
		desc:   "Without second and missing zone",
		value:  []byte("Sat, 2 Feb 2019 00:55"),
		expErr: "unpackDate: missing zone",
	}, {
		desc:   "With invalid second",
		value:  []byte("Sat, 2 Feb 2019 00:55:60 +0000"),
		expErr: "unpackDate: invalid second: 60",
	}, {
		desc:   "With missing zone",
		value:  []byte("Sat, 2 Feb 2019 00:55:16"),
		expErr: "unpackDate: missing zone",
	}, {
		desc:  "With zone",
		value: []byte("Sat, 2 Feb 2019 00:55:16 UTC"),
		exp:   time.Date(2019, time.February, 2, 0, 55, 16, 0, time.UTC),
	}, {
		desc:  "With +0800",
		value: []byte("Sat, 2 Feb 2019 00:55:16 +0800"),
		exp:   time.Date(2019, time.February, 2, 0, 55, 16, 0, time.FixedZone("UTC", 8*60*60)),
	}, {
		desc:  "Without week day",
		value: []byte("2 Feb 2019 00:55:16 UTC"),
		exp:   time.Date(2019, time.February, 2, 0, 55, 16, 0, time.UTC),
	}, {
		desc:  "Without second",
		value: []byte("Sat, 2 Feb 2019 00:55 UTC"),
		exp:   time.Date(2019, time.February, 2, 0, 55, 0, 0, time.UTC),
	}, {
		desc:  "Without week-day and second",
		value: []byte("2 Feb 2019 00:55 UTC"),
		exp:   time.Date(2019, time.February, 2, 0, 55, 0, 0, time.UTC),
	}, {
		desc:  "With obsolete year 2 digits",
		value: []byte("2 Feb 19 00:55 UTC"),
		exp:   time.Date(2019, time.February, 2, 0, 55, 0, 0, time.UTC),
	}, {
		desc:  "With obsolete year 3 digits",
		value: []byte("2 Feb 89 00:55 UTC"),
		exp:   time.Date(1989, time.February, 2, 0, 55, 0, 0, time.UTC),
	}}

	field := &Field{
		Type: FieldTypeDate,
	}

	for _, c := range cases {
		t.Log(c.desc)

		field.setValue(c.value)

		err := field.Unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "date", c.exp.String(), field.date.String(), true)
	}
}

func TestUnpackMailbox(t *testing.T) {
	cases := []struct {
		in     []byte
		expErr string
		exp    string
	}{{
		in:     []byte("Sender: local\r\n"),
		expErr: "ParseAddress: empty or invalid address",
	}, {
		in:     []byte("Sender: test@one, test@two\r\n"),
		expErr: "multiple address in sender: 'test@one, test@two\r\n'",
	}, {
		in:  []byte("Sender: <test@one>\r\n"),
		exp: "sender:<test@one>\r\n",
	}}

	for _, c := range cases {
		t.Logf("%s", c.in)

		field, _, err := ParseField(c.in)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		err = field.Unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Sender:", []byte(c.exp), field.Relaxed(), true)
	}
}

func TestUnpackMailboxList(t *testing.T) {
	cases := []struct {
		in     []byte
		expErr string
		exp    string
	}{{
		in:     []byte("From: \r\n"),
		expErr: "ParseField: invalid input",
	}, {
		in:  []byte("From: test@one, test@two\r\n"),
		exp: "from:test@one, test@two\r\n",
	}}

	for _, c := range cases {
		t.Logf("%s", c.in)

		field, _, err := ParseField(c.in)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		err = field.Unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "From:", []byte(c.exp), field.Relaxed(), true)
	}
}

func TestUnpackContentType(t *testing.T) {
	cases := []struct {
		in     []byte
		expErr string
		exp    string
	}{{
		in:     []byte("Content-Type: text;\r\n"),
		expErr: "ParseContentType: missing subtype",
	}, {
		in:  []byte("Content-Type: text/plain;\r\n"),
		exp: "text/plain;",
	}}

	for _, c := range cases {
		t.Logf("%s", c.in)

		field, _, err := ParseField(c.in)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		err = field.Unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Content-Type", c.exp, field.ContentType.String(), true)
		test.Assert(t, "field.unpacked", true, field.unpacked, true)

		err = field.Unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		test.Assert(t, "Content-Type", c.exp, field.ContentType.String(), true)
		test.Assert(t, "field.unpacked", true, field.unpacked, true)
	}
}
