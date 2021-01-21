// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewAnswers(t *testing.T) {
	cases := []struct {
		desc   string
		an     *Answer
		expLen int
		expV   []*Answer
	}{{
		desc: "With nil parameter",
		expV: make([]*Answer, 0, 1),
	}, {
		desc:   "With nil message",
		an:     &Answer{},
		expLen: 0,
		expV:   []*Answer{},
	}, {
		desc: "With valid answer",
		an: &Answer{
			msg: &Message{},
		},
		expLen: 1,
		expV: []*Answer{{
			msg: &Message{},
		}},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := newAnswers(c.an)

		test.Assert(t, "len(answers.v)", len(got.v), c.expLen, true)
		test.Assert(t, "answers.v", got.v, c.expV, true)
	}
}

func TestAnswersGet(t *testing.T) {
	msg := &Message{
		Question: SectionQuestion{
			Name:  "test",
			Type:  1,
			Class: 1,
		},
		Answer: []ResourceRecord{{
			Name:  "test",
			Type:  QueryTypeA,
			Class: QueryClassIN,
		}},
	}
	an := newAnswer(msg, true)
	ans := newAnswers(an)

	cases := []struct {
		desc     string
		QType    uint16
		QClass   uint16
		exp      *Answer
		expIndex int
	}{{
		desc:     "With query type and class not found",
		expIndex: 1,
	}, {
		desc:     "With query type not found",
		QClass:   1,
		expIndex: 1,
	}, {
		desc:     "With query class not found",
		QType:    1,
		expIndex: 1,
	}, {
		desc:     "With valid query type and class",
		QType:    1,
		QClass:   1,
		exp:      an,
		expIndex: 0,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, x := ans.get(c.QType, c.QClass)

		test.Assert(t, "answers.get", c.exp, got, true)
		test.Assert(t, "answers.get index", c.expIndex, x, true)
	}
}

func TestAnswersRemove(t *testing.T) {
	msg := &Message{
		Question: SectionQuestion{
			Name:  "test",
			Type:  1,
			Class: 1,
		},
		Answer: []ResourceRecord{{
			Name:  "test",
			Type:  QueryTypeA,
			Class: QueryClassIN,
		}},
	}

	an := newAnswer(msg, true)
	ans := newAnswers(an)

	cases := []struct {
		desc          string
		QType, QClass uint16
		exp           *answers
		expLen        int
	}{{
		desc:   "With query type and class not found",
		exp:    ans,
		expLen: 1,
	}, {
		desc:   "With query type not found",
		QClass: 1,
		exp:    ans,
		expLen: 1,
	}, {
		desc:   "With query class not found",
		QType:  1,
		exp:    ans,
		expLen: 1,
	}, {
		desc:   "With valid query type and class",
		QType:  1,
		QClass: 1,
		exp: &answers{
			v: make([]*Answer, 0, 1),
		},
		expLen: 0,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ans.remove(c.QType, c.QClass)

		test.Assert(t, "len(answers.v)", c.expLen, len(ans.v), true)
		test.Assert(t, "cap(answers.v)", 1, cap(ans.v), true)
		test.Assert(t, "answers", c.exp, ans, true)
	}
}

func TestAnswersUpdate(t *testing.T) {
	an1 := &Answer{
		QType:  1,
		QClass: 1,
		msg: &Message{
			Header: SectionHeader{
				ID: 1,
			},
		},
	}
	an2 := &Answer{
		QType:  2,
		QClass: 1,
		msg:    &Message{},
	}
	an3 := &Answer{
		QType:  1,
		QClass: 2,
		msg:    &Message{},
	}
	an4 := &Answer{
		QType:  1,
		QClass: 1,
		msg: &Message{
			Header: SectionHeader{
				ID: 2,
			},
		},
	}

	ans := newAnswers(an1)

	cases := []struct {
		desc string
		nu   *Answer
		exp  *answers
	}{{
		desc: "With nil parameter",
		exp:  ans,
	}, {
		desc: "With query type not found",
		nu:   an2,
		exp: &answers{
			v: []*Answer{
				an1,
				an2,
			},
		},
	}, {
		desc: "With query class not found",
		nu:   an3,
		exp: &answers{
			v: []*Answer{
				an1,
				an2,
				an3,
			},
		},
	}, {
		desc: "With query found",
		nu:   an4,
		exp: &answers{
			v: []*Answer{
				{
					QType:  1,
					QClass: 1,
					msg:    an4.msg,
				},
				an2,
				an3,
			},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		ans.upsert(c.nu)

		test.Assert(t, "answers.upsert", c.exp, ans, true)
	}
}
