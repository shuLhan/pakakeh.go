// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestNewCaches(t *testing.T) {
	cases := []struct {
		desc           string
		pruneDelay     time.Duration
		pruneThreshold time.Duration
		expDelay       time.Duration
		expThreshold   time.Duration
	}{{
		desc:         "With invalid delay and threshold",
		expDelay:     time.Hour,
		expThreshold: -time.Hour,
	}, {
		desc:           "With 2m delay and threshold",
		pruneDelay:     2 * time.Minute,
		pruneThreshold: -2 * time.Minute,
		expDelay:       2 * time.Minute,
		expThreshold:   -2 * time.Minute,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := newCaches(c.pruneDelay, c.pruneThreshold)

		test.Assert(t, "caches.pruneDelay", c.expDelay,
			got.pruneDelay, true)
		test.Assert(t, "caches.pruneThreshold", c.expThreshold,
			got.pruneThreshold, true)
	}
}

func TestCachesGet(t *testing.T) {
	an1 := &answer{
		receivedAt: 1,
		qname:      "test",
		qtype:      1,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 1,
			},
		},
	}
	an2 := &answer{
		receivedAt: 2,
		qname:      "test",
		qtype:      2,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 2,
			},
		},
	}
	an3 := &answer{
		receivedAt: 3,
		qname:      "test",
		qtype:      3,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 3,
			},
		},
	}

	ca := newCaches(0, 0)

	ca.upsert(an1)
	ca.upsert(an2)
	ca.upsert(an3)

	cases := []struct {
		desc    string
		qname   string
		qtype   uint16
		qclass  uint16
		exp     *answer
		expList []*answer
	}{{
		desc: "With query not found",
		expList: []*answer{
			an1, an2, an3,
		},
	}, {
		desc:   "With query found",
		qname:  "test",
		qtype:  1,
		qclass: 1,
		exp:    an1,
		expList: []*answer{
			an2, an3, an1,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := ca.get(c.qname, c.qtype, c.qclass)
		gotList := ca.list()

		test.Assert(t, "caches.get", c.exp, got, true)
		test.Assert(t, "caches.list", c.expList, gotList, true)
	}
}

func TestCachesPrune(t *testing.T) {
	at := time.Now().Unix()

	an1 := &answer{
		receivedAt: 1,
		accessedAt: 1,
		qname:      "test",
		qtype:      1,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 1,
			},
		},
	}
	an2 := &answer{
		receivedAt: 2,
		accessedAt: 2,
		qname:      "test",
		qtype:      2,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 2,
			},
		},
	}
	an3 := &answer{
		receivedAt: at,
		accessedAt: at,
		qname:      "test",
		qtype:      3,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 3,
			},
		},
	}

	ca := newCaches(0, 0)

	ca.upsert(an1)
	ca.upsert(an2)
	ca.upsert(an3)

	t.Logf("%+v\n", ca.list())

	cases := []struct {
		desc    string
		expList []*answer
	}{{
		desc: "With several caches got pruned",
		expList: []*answer{
			an3,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ca.prune()

		gotList := ca.list()

		test.Assert(t, "caches.list", c.expList, gotList, true)
	}
}

func TestCachesUpsert(t *testing.T) {
	ca := newCaches(0, 0)

	an1 := &answer{
		receivedAt: 1,
		accessedAt: 1,
		qname:      "test",
		qtype:      1,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 1,
			},
		},
	}
	an1Update := &answer{
		receivedAt: 3,
		accessedAt: 3,
		qname:      "test",
		qtype:      1,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 3,
			},
		},
	}
	an2 := &answer{
		receivedAt: 2,
		accessedAt: 2,
		qname:      "test",
		qtype:      2,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 2,
			},
		},
	}
	an2Update := &answer{
		receivedAt: 4,
		accessedAt: 4,
		qname:      "test",
		qtype:      2,
		qclass:     1,
		msg: &Message{
			Header: &SectionHeader{
				ID: 4,
			},
		},
	}

	cases := []struct {
		desc    string
		nu      *answer
		expLen  int
		expList []*answer
	}{{
		desc: "With empty answer",
	}, {
		desc:    "With new answer",
		nu:      an1,
		expLen:  1,
		expList: []*answer{an1},
	}, {
		desc:    "With new answer, different type",
		nu:      an2,
		expLen:  2,
		expList: []*answer{an1, an2},
	}, {
		desc:    "With update on answer",
		nu:      an1Update,
		expLen:  2,
		expList: []*answer{an2, an1},
	}, {
		desc:    "With update on answer (2)",
		nu:      an2Update,
		expLen:  2,
		expList: []*answer{an1, an2},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ca.upsert(c.nu)

		gotList := ca.list()

		test.Assert(t, "len(caches.list)", c.expLen, len(gotList), true)

		for x := 0; x < len(gotList); x++ {
			test.Assert(t, "caches.list", c.expList[x], gotList[x], true)
		}
	}
}
