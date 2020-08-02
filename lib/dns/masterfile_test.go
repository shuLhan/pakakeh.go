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

		libtest.Assert(t, "origin", c.exp, m.origin, true)
	}
}

func TestMasterParseDirectiveInclude(t *testing.T) {
	cases := []struct {
		desc   string
		in     string
		expErr string
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
		desc:   "RFC1035 section 5.3",
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
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("isi.edu"),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("isi.edu"),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataSOA{
					MName:   []byte("venera.isi.edu"),
					RName:   []byte("action\\.domains.isi.edu"),
					Serial:  20,
					Refresh: 7200,
					Retry:   600,
					Expire:  3600000,
					Minimum: 60,
				},
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 3,
			},
			Question: SectionQuestion{
				Name:  []byte("isi.edu"),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("isi.edu"),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("a.isi.edu"),
			}, {
				Name:  []byte("isi.edu"),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("venera.isi.edu"),
			}, {
				Name:  []byte("isi.edu"),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("vaxa.isi.edu"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 2,
			},
			Question: SectionQuestion{
				Name:  []byte("isi.edu"),
				Type:  QueryTypeMX,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("isi.edu"),
				Type:  QueryTypeMX,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataMX{
					Preference: 10,
					Exchange:   []byte("venera.isi.edu"),
				},
			}, {
				Name:  []byte("isi.edu"),
				Type:  QueryTypeMX,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataMX{
					Preference: 20,
					Exchange:   []byte("vaxa.isi.edu"),
				},
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("a.isi.edu"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("a.isi.edu"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("26.3.0.103"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 2,
			},
			Question: SectionQuestion{
				Name:  []byte("venera.isi.edu"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("venera.isi.edu"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("10.1.0.52"),
			}, {
				Name:  []byte("venera.isi.edu"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("128.9.0.32"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 2,
			},
			Question: SectionQuestion{
				Name:  []byte("vaxa.isi.edu"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("vaxa.isi.edu"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("10.2.0.27"),
			}, {
				Name:  []byte("vaxa.isi.edu"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("128.9.0.33"),
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
				libtest.Assert(t, "Answer.Name", c.exp[x].Answer[y].Name, answer.Name, true)
				libtest.Assert(t, "Answer.Type", c.exp[x].Answer[y].Type, answer.Type, true)
				libtest.Assert(t, "Answer.Class", c.exp[x].Answer[y].Class, answer.Class, true)
				libtest.Assert(t, "Answer.TTL", c.exp[x].Answer[y].TTL, answer.TTL, true)
				libtest.Assert(t, "Answer.Value", c.exp[x].Answer[y].Value, answer.Value, true)
			}
			for y, auth := range msg.Authority {
				libtest.Assert(t, "Authority.Name", c.exp[x].Authority[y].Name, auth.Name, true)
				libtest.Assert(t, "Authority.Type", c.exp[x].Authority[y].Type, auth.Type, true)
				libtest.Assert(t, "Authority.Class", c.exp[x].Authority[y].Class, auth.Class, true)
				libtest.Assert(t, "Authority.TTL", c.exp[x].Authority[y].TTL, auth.TTL, true)
				libtest.Assert(t, "Authority.Value", c.exp[x].Authority[y].Value, auth.Value, true)
			}
			for y, add := range msg.Additional {
				libtest.Assert(t, "Additional.Name", c.exp[x].Additional[y].Name, add.Name, true)
				libtest.Assert(t, "Additional.Type", c.exp[x].Additional[y].Type, add.Type, true)
				libtest.Assert(t, "Additional.Class", c.exp[x].Additional[y].Class, add.Class, true)
				libtest.Assert(t, "Additional.TTL", c.exp[x].Additional[y].TTL, add.TTL, true)
				libtest.Assert(t, "Additional.Value", c.exp[x].Additional[y].Value, add.Value, true)
			}
		}
	}
}

func TestMasterInit2(t *testing.T) {
	cases := []struct {
		desc   string
		origin string
		ttl    uint32
		in     string
		expErr error
		exp    []*Message
	}{{
		desc: "From http://www.tcpipguide.com/free/t_DNSMasterFileFormat-4.htm",
		in: `
$ORIGIN pcguide.com.
@ IN SOA ns23.pair.com. root.pair.com. (
2001072300 ; Serial
3600 ; Refresh
300 ; Retry
604800 ; Expire
3600 ) ; Minimum

@ IN NS ns23.pair.com.
@ IN NS ns0.ns0.com.

localhost IN A 127.0.0.1
@ IN A 209.68.14.80
  IN MX 50 qs939.pair.com.

www IN CNAME @
ftp IN CNAME @
mail IN CNAME @
relay IN CNAME relay.pair.com.
`,
		exp: []*Message{{
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeSOA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataSOA{
					MName:   []byte("ns23.pair.com"),
					RName:   []byte("root.pair.com"),
					Serial:  2001072300,
					Refresh: 3600,
					Retry:   300,
					Expire:  604800,
					Minimum: 3600,
				},
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 2,
			},
			Question: SectionQuestion{
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("ns23.pair.com"),
			}, {
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeNS,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("ns0.ns0.com"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("localhost.pcguide.com"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("localhost.pcguide.com"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("127.0.0.1"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("209.68.14.80"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeMX,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("pcguide.com"),
				Type:  QueryTypeMX,
				Class: QueryClassIN,
				TTL:   3600,
				Value: &RDataMX{
					Preference: 50,
					Exchange:   []byte("qs939.pair.com"),
				},
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("www.pcguide.com"),
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("www.pcguide.com"),
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("pcguide.com"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("ftp.pcguide.com"),
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("ftp.pcguide.com"),
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("pcguide.com"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("mail.pcguide.com"),
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("mail.pcguide.com"),
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("pcguide.com"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("relay.pcguide.com"),
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("relay.pcguide.com"),
				Type:  QueryTypeCNAME,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("relay.pair.com"),
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
				libtest.Assert(t, "Answer.Name", c.exp[x].Answer[y].Name, answer.Name, true)
				libtest.Assert(t, "Answer.Type", c.exp[x].Answer[y].Type, answer.Type, true)
				libtest.Assert(t, "Answer.Class", c.exp[x].Answer[y].Class, answer.Class, true)
				libtest.Assert(t, "Answer.TTL", c.exp[x].Answer[y].TTL, answer.TTL, true)
				libtest.Assert(t, "Answer.Value", c.exp[x].Answer[y].Value, answer.Value, true)
			}
			for y, auth := range msg.Authority {
				libtest.Assert(t, "Authority.Name", c.exp[x].Authority[y].Name, auth.Name, true)
				libtest.Assert(t, "Authority.Type", c.exp[x].Authority[y].Type, auth.Type, true)
				libtest.Assert(t, "Authority.Class", c.exp[x].Authority[y].Class, auth.Class, true)
				libtest.Assert(t, "Authority.TTL", c.exp[x].Authority[y].TTL, auth.TTL, true)
				libtest.Assert(t, "Authority.Value", c.exp[x].Authority[y].Value, auth.Value, true)
			}
			for y, add := range msg.Additional {
				libtest.Assert(t, "Additional.Name", c.exp[x].Additional[y].Name, add.Name, true)
				libtest.Assert(t, "Additional.Type", c.exp[x].Additional[y].Type, add.Type, true)
				libtest.Assert(t, "Additional.Class", c.exp[x].Additional[y].Class, add.Class, true)
				libtest.Assert(t, "Additional.TTL", c.exp[x].Additional[y].TTL, add.TTL, true)
				libtest.Assert(t, "Additional.Value", c.exp[x].Additional[y].Value, add.Value, true)
			}
		}
	}
}

func TestMasterInit3(t *testing.T) {
	cases := []struct {
		desc   string
		origin string
		ttl    uint32
		in     string
		expErr error
		exp    []*Message
	}{{
		desc:   "From http://www.tcpipguide.com/free/t_DNSMasterFileFormat-4.htm",
		origin: "localdomain",
		in: `
; Applications.
dev.kilabit.info.  A  127.0.0.1
dev.kilabit.com.   A  127.0.0.1

; Documentations.
angularjs.doc       A  127.0.0.1
`,
		exp: []*Message{{
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("dev.kilabit.info"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("dev.kilabit.info"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("127.0.0.1"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("dev.kilabit.com"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("dev.kilabit.com"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("127.0.0.1"),
			}},
		}, {
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("angularjs.doc.localdomain"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("angularjs.doc.localdomain"),
				Type:  QueryTypeA,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte("127.0.0.1"),
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
				libtest.Assert(t, "Answer.Name", c.exp[x].Answer[y].Name, answer.Name, true)
				libtest.Assert(t, "Answer.Type", c.exp[x].Answer[y].Type, answer.Type, true)
				libtest.Assert(t, "Answer.Class", c.exp[x].Answer[y].Class, answer.Class, true)
				libtest.Assert(t, "Answer.TTL", c.exp[x].Answer[y].TTL, answer.TTL, true)
				libtest.Assert(t, "Answer.Value", c.exp[x].Answer[y].Value, answer.Value, true)
			}
			for y, auth := range msg.Authority {
				libtest.Assert(t, "Authority.Name", c.exp[x].Authority[y].Name, auth.Name, true)
				libtest.Assert(t, "Authority.Type", c.exp[x].Authority[y].Type, auth.Type, true)
				libtest.Assert(t, "Authority.Class", c.exp[x].Authority[y].Class, auth.Class, true)
				libtest.Assert(t, "Authority.TTL", c.exp[x].Authority[y].TTL, auth.TTL, true)
				libtest.Assert(t, "Authority.Value", c.exp[x].Authority[y].Value, auth.Value, true)
			}
			for y, add := range msg.Additional {
				libtest.Assert(t, "Additional.Name", c.exp[x].Additional[y].Name, add.Name, true)
				libtest.Assert(t, "Additional.Type", c.exp[x].Additional[y].Type, add.Type, true)
				libtest.Assert(t, "Additional.Class", c.exp[x].Additional[y].Class, add.Class, true)
				libtest.Assert(t, "Additional.TTL", c.exp[x].Additional[y].TTL, add.TTL, true)
				libtest.Assert(t, "Additional.Value", c.exp[x].Additional[y].Value, add.Value, true)
			}
		}
	}
}

func TestMasterParseTXT(t *testing.T) {
	cases := []struct {
		in       string
		exp      []*Message
		expError string
	}{{
		in: `@ IN TXT "This is a test"`,
		exp: []*Message{{
			Header: SectionHeader{
				IsAA:    true,
				QDCount: 1,
				ANCount: 1,
			},
			Question: SectionQuestion{
				Name:  []byte("kilabit.local"),
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  []byte("kilabit.local"),
				Type:  QueryTypeTXT,
				Class: QueryClassIN,
				TTL:   3600,
				Value: []byte(`This is a test`),
			}},
		}},
	}}

	m := newMaster()

	for _, c := range cases {
		m.Init(c.in, "kilabit.local", 3600)

		err := m.parse()
		if err != nil {
			libtest.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}

		libtest.Assert(t, "messages length:", len(c.exp), len(m.msgs), true)

		for x, msg := range m.msgs {
			libtest.Assert(t, "Message.Header", c.exp[x].Header, msg.Header, true)
			libtest.Assert(t, "Message.Question", c.exp[x].Question, msg.Question, true)

			for y, answer := range msg.Answer {
				libtest.Assert(t, "Answer.Name", c.exp[x].Answer[y].Name, answer.Name, true)
				libtest.Assert(t, "Answer.Type", c.exp[x].Answer[y].Type, answer.Type, true)
				libtest.Assert(t, "Answer.Class", c.exp[x].Answer[y].Class, answer.Class, true)
				libtest.Assert(t, "Answer.TTL", c.exp[x].Answer[y].TTL, answer.TTL, true)
				libtest.Assert(t, "Answer.Value", c.exp[x].Answer[y].Value, answer.Value, true)
			}
			for y, auth := range msg.Authority {
				libtest.Assert(t, "Authority.Name", c.exp[x].Authority[y].Name, auth.Name, true)
				libtest.Assert(t, "Authority.Type", c.exp[x].Authority[y].Type, auth.Type, true)
				libtest.Assert(t, "Authority.Class", c.exp[x].Authority[y].Class, auth.Class, true)
				libtest.Assert(t, "Authority.TTL", c.exp[x].Authority[y].TTL, auth.TTL, true)
				libtest.Assert(t, "Authority.Value", c.exp[x].Authority[y].Value, auth.Value, true)
			}
			for y, add := range msg.Additional {
				libtest.Assert(t, "Additional.Name", c.exp[x].Additional[y].Name, add.Name, true)
				libtest.Assert(t, "Additional.Type", c.exp[x].Additional[y].Type, add.Type, true)
				libtest.Assert(t, "Additional.Class", c.exp[x].Additional[y].Class, add.Class, true)
				libtest.Assert(t, "Additional.TTL", c.exp[x].Additional[y].TTL, add.TTL, true)
				libtest.Assert(t, "Additional.Value", c.exp[x].Additional[y].Value, add.Value, true)
			}
		}
	}
}
