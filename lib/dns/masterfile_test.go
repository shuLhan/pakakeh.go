// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	libtest "github.com/shuLhan/share/lib/test"
)

func TestMasterParseDirectiveOrigin(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
		exp    string
	}{{
		desc:   "Without value",
		in:     `$origin`,
		expErr: "! (data):1 Empty $origin directive",
	}, {
		desc:   "Without value and with comment",
		in:     `$origin ; comment`,
		expErr: "! (data):1 Empty $origin directive",
	}, {
		desc: "With value",
		in:   `$origin x`,
		exp:  "x",
	}, {
		desc: "With value and comment",
		in:   `$origin x ;comment`,
		exp:  "x",
	}}

	m := newMaster()

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, "", 0)

		err := m.parse()
		if err != nil {
			libtest.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		libtest.Assert(t, "origin", c.exp, string(m.origin), true)
	}
}

func TestMasterParseDirectiveInclude(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
		exp    string
	}{{
		desc:   "Without value",
		in:     `$include`,
		expErr: "! (data):1 Empty $include directive",
	}, {
		desc:   "Without value and with comment",
		in:     `$include ; comment`,
		expErr: "! (data):1 Empty $include directive",
	}, {
		desc: "With value",
		in:   `$include testdata/sub.domain`,
	}, {
		desc: "With value and comment",
		in:   `$origin testdata/sub.domain ;comment`,
	}}

	m := newMaster()

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, "", 0)

		err := m.parse()
		if err != nil {
			libtest.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}
	}
}

func TestMasterParseDirectiveTTL(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
		exp    uint32
	}{{
		desc:   "Without value",
		in:     `$ttl`,
		expErr: "! (data):1 Empty $ttl directive",
	}, {
		desc:   "Without value and with comment",
		in:     `$ttl ; comment`,
		expErr: "! (data):1 Empty $ttl directive",
	}, {
		desc: "With value",
		in:   `$ttl 1`,
		exp:  1,
	}, {
		desc: "With value and comment",
		in:   `$ttl 1 ;comment`,
		exp:  1,
	}}

	m := newMaster()

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, "", 0)

		err := m.parse()
		if err != nil {
			libtest.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		libtest.Assert(t, "ttl", c.exp, m.ttl, true)
	}
}

func TestMasterInitRFC1035(t *testing.T) {
	cases := []struct {
		desc   string
		origin string
		ttl    uint32
		in     string
		expErr error
		exp    []*Message
	}{{
		desc:   "",
		origin: "ISI.EDU.",
		ttl:    3600,
		in: `
@   IN  SOA     VENERA      Action\.domains (
                                 20     ; SERIAL
                                 7200   ; REFRESH
                                 600    ; RETRY
                                 3600000; EXPIRE
                                 60)    ; MINIMUM

        NS      A.ISI.EDU.
        NS      VENERA
        NS      VAXA
        MX      10      VENERA
        MX      20      VAXA

A       A       26.3.0.103

VENERA  A       10.1.0.52
        A       128.9.0.32

VAXA    A       10.2.0.27
        A       128.9.0.33

`,
		exp: []*Message{{
			Header: &SectionHeader{},
			Question: &SectionQuestion{
				Name:  []byte("isi.edu."),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("isi.edu."),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
				TTL:   3600,
				SOA: &RDataSOA{
					MName:   []byte("venera.isi.edu."),
					RName:   []byte("action\\.domains.isi.edu."),
					Serial:  20,
					Refresh: 7200,
					Retry:   600,
					Expire:  3600000,
					Minimum: 60,
				},
			}},
		}, {
			Header: &SectionHeader{},
			Question: &SectionQuestion{
				Name:  []byte("isi.edu."),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("isi.edu."),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("a.isi.edu."),
				},
			}, {
				Name:  []byte("isi.edu."),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("venera.isi.edu."),
				},
			}, {
				Name:  []byte("isi.edu."),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("vaxa.isi.edu."),
				},
			}},
		}, {
			Header: &SectionHeader{},
			Question: &SectionQuestion{
				Name:  []byte("isi.edu."),
				Type:  QueryTypeMX,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("isi.edu."),
				Type:  QueryTypeMX,
				Class: QueryClassIN,
				TTL:   3600,
				MX: &RDataMX{
					Preference: 10,
					Exchange:   []byte("venera.isi.edu."),
				},
			}, {
				Name:  []byte("isi.edu."),
				Type:  QueryTypeMX,
				Class: QueryClassIN,
				TTL:   3600,
				MX: &RDataMX{
					Preference: 20,
					Exchange:   []byte("vaxa.isi.edu."),
				},
			}},
		}, {
			Header: &SectionHeader{},
			Question: &SectionQuestion{
				Name:  []byte("a.isi.edu."),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("a.isi.edu."),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("26.3.0.103"),
				},
			}},
		}, {
			Header: &SectionHeader{},
			Question: &SectionQuestion{
				Name:  []byte("venera.isi.edu."),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("venera.isi.edu."),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("10.1.0.52"),
				},
			}, {
				Name:  []byte("venera.isi.edu."),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("128.9.0.32"),
				},
			}},
		}, {
			Header: &SectionHeader{},
			Question: &SectionQuestion{
				Name:  []byte("vaxa.isi.edu."),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []*ResourceRecord{{
				Name:  []byte("vaxa.isi.edu."),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("10.2.0.27"),
				},
			}, {
				Name:  []byte("vaxa.isi.edu."),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Text: &RDataText{
					v: []byte("128.9.0.33"),
				},
			}},
		}},
	}}

	m := newMaster()

	for _, c := range cases {
		t.Log(c.desc)

		m.Init(c.in, c.origin, c.ttl)

		err := m.parse()
		if err != nil {
			libtest.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}

		libtest.Assert(t, "messages length:", len(c.exp), len(m.msgs), true)

		for x, msg := range m.msgs {
			libtest.Assert(t, "Message.Header", c.exp[x].Header, msg.Header, true)
			libtest.Assert(t, "Message.Question", c.exp[x].Question, msg.Question, true)

			for y, answer := range msg.Answer {
				t.Logf("Expecting answer rdata: %s\n", c.exp[x].Answer[y].RData())
				t.Logf("Got answer rdata: %s\n", answer.RData())

				libtest.Assert(t, "Answer.Name", c.exp[x].Answer[y].Name, answer.Name, true)
				libtest.Assert(t, "Answer.Type", c.exp[x].Answer[y].Type, answer.Type, true)
				libtest.Assert(t, "Answer.Class", c.exp[x].Answer[y].Class, answer.Class, true)
				libtest.Assert(t, "Answer.TTL", c.exp[x].Answer[y].TTL, answer.TTL, true)
				libtest.Assert(t, "Answer.RData()", c.exp[x].Answer[y].RData(), answer.RData(), true)
			}
			for y, auth := range msg.Authority {
				libtest.Assert(t, "Authority.Name", c.exp[x].Authority[y].Name, auth.Name, true)
				libtest.Assert(t, "Authority.Type", c.exp[x].Authority[y].Type, auth.Type, true)
				libtest.Assert(t, "Authority.Class", c.exp[x].Authority[y].Class, auth.Class, true)
				libtest.Assert(t, "Authority.TTL", c.exp[x].Authority[y].TTL, auth.TTL, true)
				libtest.Assert(t, "Authority.RData()", c.exp[x].Authority[y].RData(), auth.RData(), true)
			}
			for y, add := range msg.Additional {
				libtest.Assert(t, "Additional.Name", c.exp[x].Additional[y].Name, add.Name, true)
				libtest.Assert(t, "Additional.Type", c.exp[x].Additional[y].Type, add.Type, true)
				libtest.Assert(t, "Additional.Class", c.exp[x].Additional[y].Class, add.Class, true)
				libtest.Assert(t, "Additional.TTL", c.exp[x].Additional[y].TTL, add.TTL, true)
				libtest.Assert(t, "Additional.RData()", c.exp[x].Additional[y].RData(), add.RData(), true)
			}
		}
	}
}
