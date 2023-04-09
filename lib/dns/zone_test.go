// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/test"
)

func TestParseZone(t *testing.T) {
	var (
		listTData []*test.Data
		err       error
	)

	listTData, err = test.LoadDataDir(`testdata/zone/`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tdata  *test.Data
		zone   *Zone
		bb     bytes.Buffer
		origin string
		vstr   string
		vbytes []byte
		ttl    int64
	)
	for _, tdata = range listTData {
		t.Log(tdata.Name)

		origin = tdata.Flag[`origin`]

		vstr = tdata.Flag[`ttl`]
		if len(vstr) > 0 {
			ttl, err = strconv.ParseInt(vstr, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
		} else {
			ttl = 0
		}

		vbytes = tdata.Input[`zone_in.txt`]
		zone, err = ParseZone(vbytes, origin, uint32(ttl))
		if err != nil {
			t.Fatal(err)
		}

		// Compare the zone by writing back to text.

		bb.Reset()
		_, err = zone.WriteTo(&bb)
		if err != nil {
			t.Fatal(err)
		}

		vstr = `zone_out.txt`
		vbytes = tdata.Output[vstr]
		test.Assert(t, vstr, string(vbytes), bb.String())

		// Compare the packed zone as message.

		var (
			msg *Message
			x   int
		)
		for x, msg = range zone.messages {
			bb.Reset()
			libbytes.DumpPrettyTable(&bb, msg.Question.String(), msg.packet)

			vstr = fmt.Sprintf(`message_%d.hex`, x)
			vbytes = tdata.Output[vstr]
			test.Assert(t, vstr, string(vbytes), bb.String())
		}
	}
}

func TestZoneParseDirectiveOrigin(t *testing.T) {
	type testCase struct {
		desc   string
		in     string
		expErr string
		exp    string
	}

	var (
		m = newZoneParser(nil)

		cases []testCase
		c     testCase
		err   error
	)

	cases = []testCase{{
		desc:   `Without value`,
		in:     `$origin`,
		expErr: `parse: parseDirectiveOrigin: line 1: empty $origin directive`,
	}, {
		desc:   `Without value and with comment`,
		in:     `$origin ; comment`,
		expErr: `parse: parseDirectiveOrigin: line 1: empty $origin directive`,
	}, {
		desc: `With value`,
		in:   `$origin x`,
		exp:  `x`,
	}, {
		desc: `With value and comment`,
		in:   `$origin x ;comment`,
		exp:  `x`,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		m.Reset([]byte(c.in), ``, 0)

		err = m.parse()
		if err != nil {
			test.Assert(t, `error`, c.expErr, err.Error())
			continue
		}

		test.Assert(t, `origin`, c.exp, m.origin)
	}
}

func TestZoneParseDirectiveInclude(t *testing.T) {
	type testCase struct {
		desc   string
		in     string
		expErr string
	}

	var (
		m = newZoneParser(nil)

		cases []testCase
		c     testCase
		err   error
	)

	cases = []testCase{{
		desc:   `Without value`,
		in:     `$include`,
		expErr: `parse: parseDirectiveInclude: line 1: empty $include directive`,
	}, {
		desc:   `Without value and with comment`,
		in:     `$include ; comment`,
		expErr: `parse: parseDirectiveInclude: line 1: empty $include directive`,
	}, {
		desc: `With value`,
		in:   `$include testdata/sub.domain`,
	}, {
		desc: `With value and comment`,
		in:   `$include testdata/sub.domain ;comment`,
	}, {
		desc: `With value and domain name`,
		in:   `$include testdata/sub.domain sub.include`,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		m.Reset([]byte(c.in), ``, 0)

		err = m.parse()
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error())
			continue
		}
	}
}

func TestZoneParseDirectiveTTL(t *testing.T) {
	type testCase struct {
		desc   string
		in     string
		expErr string
		exp    uint32
	}

	var (
		m = newZoneParser(nil)

		cases []testCase
		c     testCase
		err   error
	)

	cases = []testCase{{
		desc:   `Without value`,
		in:     `$ttl`,
		expErr: `parse: parseDirectiveTTL: line 1: empty $TTL directive`,
	}, {
		desc:   `Without value and with comment`,
		in:     `$ttl ; comment`,
		expErr: `parse: parseDirectiveTTL: line 1: empty $TTL directive`,
	}, {
		desc:   `With invalid value`,
		in:     `$ttl a`,
		expErr: `parse: parseDirectiveTTL: line 1: invalid TTL value 'a'`,
	}, {
		desc: `With seconds value`,
		in:   `$ttl 1`,
		exp:  1,
	}, {
		desc: `With seconds value and comment`,
		in:   `$ttl 1 ;comment`,
		exp:  1,
	}, {
		desc: `With time.Duration value and comment`,
		in:   `$ttl 1m ;comment`,
		exp:  60,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		m.Reset([]byte(c.in), ``, 0)

		err = m.parse()
		if err != nil {
			test.Assert(t, `error`, c.expErr, err.Error())
			continue
		}

		test.Assert(t, `ttl`, c.exp, m.ttl)
	}
}
