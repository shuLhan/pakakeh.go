// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
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

func TestParseZone_SVCB(t *testing.T) {
	var (
		logp = `TestParseZone_SVCB`

		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/ParseZone_SVCB_test.txt`)
	if err != nil {
		t.Fatal(logp, err)
	}

	var listCase = []string{
		`AliasMode`,
		`ServiceMode`,
		`ServiceMode:port`,
		`ServiceMode:keyGeneric667`,
		`ServiceMode:keyGenericQuoted`,
		`ServiceMode:TwoQuotedIpv6Hint`,
		`ServiceMode:Ipv6hintEmbedIpv4`,
		`ServiceMode:WithMandatoryKey`,
		`ServiceMode:AlpnWithEscapedComma`,
		`ServiceMode:AlpnWithEscapedBackslash`,
		`FailureMode:DuplicateKey`,
		`FailureMode:KeyMandatoryNoValue`,
		`FailureMode:KeyAlpnNoValue`,
		`FailureMode:KeyPortNoValue`,
		`FailureMode:KeyIpv4hintNoValue`,
		`FailureMode:KeyIpv6hintNoValue`,
		`FailureMode:KeyNodefaultalpnWithValue`,
		`FailureMode:MissingMandatoryKey`,
		`FailureMode:RecursiveMandatoryKey`,
		`FailureMode:DuplicateMandatoryKey`,
	}

	var (
		origin        = `example.com`
		ttl    uint32 = 60

		name   string
		stream []byte
		zone   *Zone
		out    bytes.Buffer

		tag string
		msg *Message
		x   int
	)

	for _, name = range listCase {
		stream = tdata.Input[name]
		if len(stream) == 0 {
			t.Fatalf(`%s: %s: empty input`, logp, name)
		}

		zone, err = ParseZone(stream, origin, ttl)
		if err != nil {
			tag = name + `:error`
			test.Assert(t, tag, string(tdata.Output[tag]), err.Error())
			continue
		}

		out.Reset()

		_, _ = zone.WriteTo(&out)
		stream = tdata.Output[name]
		test.Assert(t, name, string(stream), out.String())

		for x, msg = range zone.messages {
			out.Reset()
			libbytes.DumpPrettyTable(&out, msg.Question.String(), msg.packet)

			tag = fmt.Sprintf(`%s:message_%d.hex`, name, x)
			stream = tdata.Output[tag]
			test.Assert(t, tag, string(stream), out.String())
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
		m = newZoneParser(nil, nil)

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
		exp:  `x.`,
	}, {
		desc: `With value and comment`,
		in:   `$origin x ;comment`,
		exp:  `x.`,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		m.Reset([]byte(c.in), nil)

		err = m.parse()
		if err != nil {
			test.Assert(t, `error`, c.expErr, err.Error())
			continue
		}

		test.Assert(t, `origin`, c.exp, m.zone.Origin)
	}
}

func TestZoneParseDirectiveInclude(t *testing.T) {
	type testCase struct {
		desc   string
		in     string
		expErr string
	}

	var (
		m = newZoneParser(nil, nil)

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

		m.Reset([]byte(c.in), nil)

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
		m = newZoneParser(nil, nil)

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

		m.Reset([]byte(c.in), nil)

		err = m.parse()
		if err != nil {
			test.Assert(t, `error`, c.expErr, err.Error())
			continue
		}

		test.Assert(t, `ttl`, c.exp, m.zone.SOA.Minimum)
	}
}

// TestZone_SOA test related to SOA, when SOA record updated, removed, or
// other records added or removed.
func TestZone_SOA(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/zone_soa_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		zone = NewZone(``, `test.soa`)
		buf  bytes.Buffer
		exp  []byte
	)

	_, _ = zone.WriteTo(&buf)
	exp = tdata.Output[`NewZone`]
	test.Assert(t, `NewZone`, string(exp), buf.String())

	// Add SOA.
	var (
		rdataSoa = NewRDataSOA(`new.soa`, `admin`)
		rrSoa    = &ResourceRecord{
			Value: rdataSoa,
			Name:  zone.Origin,
			Type:  RecordTypeSOA,
			Class: RecordClassIN,
		}
	)

	err = zone.Add(rrSoa)
	if err != nil {
		t.Fatal(err)
	}

	buf.Reset()
	_, _ = zone.WriteTo(&buf)
	exp = tdata.Output[`Add_SOA`]
	test.Assert(t, `Add_SOA`, string(exp), buf.String())

	// Remove SOA.
	_ = zone.Remove(rrSoa)

	buf.Reset()
	_, _ = zone.WriteTo(&buf)
	exp = tdata.Output[`Remove_SOA`]
	test.Assert(t, `Remove_SOA`, string(exp), buf.String())
}

func testGenerateZoneRecords() (zone *Zone, listRR []*ResourceRecord) {
	zone = NewZone(``, `test`)

	listRR = []*ResourceRecord{{
		Name:  `test`,
		Type:  RecordTypeA,
		Class: RecordClassIN,
		Value: `127.0.0.1`,
		TTL:   1,
	}, {
		Name:  `test`,
		Type:  RecordTypeSOA,
		Class: RecordClassIN,
		Value: &RDataSOA{},
		TTL:   2,
	}, {
		Name:  `test`,
		Type:  RecordTypeMX,
		Class: RecordClassIN,
		TTL:   3,
	}, {
		Name:  `test`,
		Type:  RecordTypeSOA,
		Class: RecordClassIN,
		TTL:   4,
	}, {
		Name:  `test`,
		Type:  RecordTypeA,
		Class: RecordClassCH,
		TTL:   5,
	}}

	var rr *ResourceRecord
	for _, rr = range listRR {
		zone.recordAdd(rr)
	}

	return zone, listRR
}

func TestZoneRecordAdd(t *testing.T) {
	var (
		gotZone *Zone
		listRR  []*ResourceRecord
	)

	gotZone, listRR = testGenerateZoneRecords()

	var expZoneRecords = map[string][]*ResourceRecord{
		`test`: []*ResourceRecord{
			listRR[0],
			listRR[3],
			listRR[2],
			listRR[4],
		},
	}

	test.Assert(t, `add`, expZoneRecords, gotZone.Records)
}

func TestZoneRecordRemove(t *testing.T) {
	type testCase struct {
		rr           *ResourceRecord
		expZoneRR    map[string][]*ResourceRecord
		expIsRemoved bool
	}

	var (
		gotZone      *Zone
		listRR       []*ResourceRecord
		cases        []testCase
		c            testCase
		gotIsRemoved bool
	)

	gotZone, listRR = testGenerateZoneRecords()

	cases = []testCase{{
		// With different value.
		rr: &ResourceRecord{
			Name:  `test`,
			Type:  RecordTypeA,
			Class: RecordClassIN,
			Value: `127.0.0.2`,
		},
		expZoneRR:    gotZone.Records,
		expIsRemoved: false,
	}, {
		// With different Class.
		rr: &ResourceRecord{
			Name:  `test`,
			Type:  RecordTypeA,
			Class: RecordClassCH,
			Value: `127.0.0.1`,
		},
		expZoneRR:    gotZone.Records,
		expIsRemoved: false,
	}, {
		// With RR removed at the end.
		rr: listRR[4],
		expZoneRR: map[string][]*ResourceRecord{
			`test`: []*ResourceRecord{
				listRR[0],
				listRR[3],
				listRR[2],
			},
		},
		expIsRemoved: true,
	}}

	for _, c = range cases {
		gotIsRemoved = gotZone.recordRemove(c.rr)
		test.Assert(t, `is removed`, c.expIsRemoved, gotIsRemoved)
		test.Assert(t, `after removed`, c.expZoneRR, gotZone.Records)
	}
}
