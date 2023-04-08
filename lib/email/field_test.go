// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

import (
	"testing"
	"time"

	"github.com/shuLhan/share/lib/ascii"
	"github.com/shuLhan/share/lib/test"
)

func TestParseField(t *testing.T) {
	longValue := string(ascii.Random([]byte(ascii.Letters), 994))

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
		expErr: "email: field line greater than 998 characters",
	}, {
		desc:   "With only whitespaces",
		raw:    []byte("  "),
		expErr: "email: invalid field at '  '",
	}, {
		desc:   "With only CRLF",
		raw:    []byte("\r\n"),
		expErr: "email: invalid field at ''",
	}, {
		desc:   "Without separator and CRLF",
		raw:    []byte("name"),
		expErr: "email: invalid field at 'name'",
	}, {
		desc:   "Without separator",
		raw:    []byte("name\r\n"),
		expErr: "email: invalid field at 'name'",
	}, {
		desc:   "With space on name",
		raw:    []byte("na me\r\n"),
		expErr: "email: missing field separator at 'na '",
	}, {
		desc:   "Without value and CRLF",
		raw:    []byte("name:"),
		expErr: "email: empty field value at 'name:'",
	}, {
		desc:   "Without value and CRLF",
		raw:    []byte("name: "),
		expErr: "email: empty field value at 'name: '",
	}, {
		desc:   "Without value",
		raw:    []byte("name:\r\n"),
		expErr: "email: empty field value at 'name:\r\n'",
	}, {
		desc:   "Without value",
		raw:    []byte("name: \r\n"),
		expErr: "email: empty field value at 'name: \r\n'",
	}, {
		desc:   "Without CRLF",
		raw:    []byte("name:value"),
		expErr: "email: field value without CRLF at 'name:value'",
	}, {
		desc:   "Without CR",
		raw:    []byte("name:value\n"),
		expErr: "email: invalid field value at 'name:value'",
	}, {
		desc:   "Without LF",
		raw:    []byte("name:value\r"),
		expErr: "email: field value without CRLF at 'name:value\r'",
	}, {
		desc:   "With CR inside value",
		raw:    []byte("name:valu\re"),
		expErr: "email: field value without CRLF at 'name:valu\r'",
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
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}
		if got == nil {
			test.Assert(t, "Field", c.exp, got)
			continue
		}

		test.Assert(t, "Field.oriName", c.exp.oriName, got.oriName)
		test.Assert(t, "Field.oriValue", c.exp.oriValue, got.oriValue)
		test.Assert(t, "Field.Name", c.exp.Name, got.Name)
		test.Assert(t, "Field.Value", c.exp.Value, got.Value)

		test.Assert(t, "rest", c.expRest, rest)
	}
}

func TestUnpackDate(t *testing.T) {
	cases := []struct {
		exp    time.Time
		expErr string
		desc   string
		value  string
	}{{
		desc:   `With empty value`,
		expErr: `unpackDate: empty date`,
	}, {
		desc:   `With only spaces`,
		value:  `  `,
		expErr: `unpackDate: empty date`,
	}, {
		desc:   `With missing day`,
		value:  `Sat`,
		expErr: `unpackDate: invalid or missing day Sat`,
	}, {
		desc:   `With missing day`,
		value:  `Sat,`,
		expErr: `unpackDate: invalid or missing day `,
	}, {
		desc:   `With invalid day of week`,
		value:  `Satu, 2`,
		expErr: `unpackDate: invalid day of week Satu`,
	}, {
		desc:   `With missing month`,
		value:  `Sat, 2`,
		expErr: `unpackDate: invalid or missing month `,
	}, {
		desc:   `With invalid month`,
		value:  `Sat, 2 X 2019`,
		expErr: `unpackDate: invalid or missing month X`,
	}, {
		desc:   `With missing year`,
		value:  `Sat, 2 Feb`,
		expErr: `unpackDate: invalid or missing year `,
	}, {
		desc:   `With missing hour`,
		value:  `Sat, 2 Feb 2019`,
		expErr: `unpackDate: invalid or missing hour `,
	}, {
		desc:   `With missing minute`,
		value:  `Sat, 2 Feb 2019 00`,
		expErr: `unpackDate: invalid or missing time separator`,
	}, {
		desc:   `With invalid hour`,
		value:  `Sat, 2 Feb 2019 24:55:16 +0000`,
		expErr: `unpackDate: invalid hour 24`,
	}, {
		desc:   `With invalid minute`,
		value:  `Sat, 2 Feb 2019 00:a`,
		expErr: `unpackDate: invalid or missing minute a`,
	}, {
		desc:   `With invalid minute #2`,
		value:  `Sat, 2 Feb 2019 00:60:16 +0000`,
		expErr: `unpackDate: invalid minute 60`,
	}, {
		desc:   `Without second and missing zone`,
		value:  `Sat, 2 Feb 2019 00:55`,
		expErr: `unpackDate: invalid or missing zone `,
	}, {
		desc:   `With invalid second`,
		value:  `Sat, 2 Feb 2019 00:55:xx +0000`,
		expErr: `unpackDate: invalid second xx`,
	}, {
		desc:   `With invalid second #2`,
		value:  `Sat, 2 Feb 2019 00:55:60 +0000`,
		expErr: `unpackDate: invalid second 60`,
	}, {
		desc:   `With second and missing zone`,
		value:  `Sat, 2 Feb 2019 00:55:16`,
		expErr: `unpackDate: invalid or missing zone `,
	}, {
		desc:   `With invalid zone offset`,
		value:  `Sat, 2 Feb 2019 00:55:16 +00T00`,
		expErr: `unpackDate: invalid or missing zone offset +00T00`,
	}, {
		desc:  `With zone`,
		value: `Sat, 2 Feb 2019 00:55:16 UTC`,
		exp:   time.Date(2019, time.February, 2, 0, 55, 16, 0, time.UTC),
	}, {
		desc:  `With +0800`,
		value: `Sat, 2 Feb 2019 00:55:16 +0800`,
		exp:   time.Date(2019, time.February, 2, 0, 55, 16, 0, time.FixedZone(`UTC`, 8*60*60)),
	}, {
		desc:  `Without week day`,
		value: `2 Feb 2019 00:55:16 UTC`,
		exp:   time.Date(2019, time.February, 2, 0, 55, 16, 0, time.UTC),
	}, {
		desc:  `Without second`,
		value: `Sat, 2 Feb 2019 00:55 UTC`,
		exp:   time.Date(2019, time.February, 2, 0, 55, 0, 0, time.UTC),
	}, {
		desc:  `Without week-day and second`,
		value: `2 Feb 2019 00:55 UTC`,
		exp:   time.Date(2019, time.February, 2, 0, 55, 0, 0, time.UTC),
	}, {
		desc:  `With obsolete year 2 digits`,
		value: `2 Feb 19 00:55 UTC`,
		exp:   time.Date(2019, time.February, 2, 0, 55, 0, 0, time.UTC),
	}, {
		desc:  `With obsolete year 3 digits`,
		value: `2 Feb 89 00:55 UTC`,
		exp:   time.Date(1989, time.February, 2, 0, 55, 0, 0, time.UTC),
	}}

	field := &Field{
		Type: FieldTypeDate,
	}

	for _, c := range cases {
		t.Log(c.desc)

		field.setValue([]byte(c.value))

		err := field.unpack()
		if err != nil {
			test.Assert(t, `error`, c.expErr, err.Error())
			continue
		}

		test.Assert(t, `date`, c.exp.String(), field.date.String())
	}
}

func TestUnpackMailbox(t *testing.T) {
	cases := []struct {
		expErr string
		exp    string
		in     []byte
	}{{
		in:     []byte("Sender: local\r\n"),
		expErr: `ParseMailboxes: empty or invalid address`,
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
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		err = field.unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "Sender:", []byte(c.exp), field.Relaxed())
	}
}

func TestUnpackMailboxList(t *testing.T) {
	cases := []struct {
		expErr string
		exp    string
		in     []byte
	}{{
		in:     []byte("From: \r\n"),
		expErr: "email: empty field value at 'From: \r\n'",
	}, {
		in:  []byte("From: test@one, test@two\r\n"),
		exp: "from:test@one, test@two\r\n",
	}}

	for _, c := range cases {
		t.Logf("%s", c.in)

		field, _, err := ParseField(c.in)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		err = field.unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "From:", []byte(c.exp), field.Relaxed())
	}
}

func TestUnpackContentType(t *testing.T) {
	cases := []struct {
		expErr string
		exp    string
		in     []byte
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
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		err = field.unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "Content-Type", c.exp, field.ContentType.String())
		test.Assert(t, "field.unpacked", true, field.unpacked)

		err = field.unpack()
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		test.Assert(t, "Content-Type", c.exp, field.ContentType.String())
		test.Assert(t, "field.unpacked", true, field.unpacked)
	}
}
