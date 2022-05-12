// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"container/list"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

func TestNewAnswer(t *testing.T) {
	type testCase struct {
		desc      string
		msg       *Message
		exp       *Answer
		expMsg    *Message
		expQName  string
		expRType  RecordType
		expRClass RecordClass
		isLocal   bool
	}

	var (
		at   = time.Now().Unix()
		msg1 = &Message{
			Header: MessageHeader{
				ID: 1,
			},
			Question: MessageQuestion{
				Name:  "test",
				Type:  1,
				Class: 1,
			},
			Answer: []ResourceRecord{{
				Name:  "test",
				Type:  RecordTypeA,
				Class: RecordClassIN,
				TTL:   3600,
				rdlen: 4,
				Value: "127.0.0.1",
			}},
		}

		cases []testCase
		c     testCase
		got   *Answer
	)

	cases = []testCase{{
		desc:    "With local message",
		msg:     msg1,
		isLocal: true,
		exp: &Answer{
			QName:  "test",
			RType:  1,
			RClass: 1,
			msg:    msg1,
		},
		expQName:  "test",
		expRType:  1,
		expRClass: 1,
		expMsg:    msg1,
	}, {
		desc: "With non local message",
		msg:  msg1,
		exp: &Answer{
			QName:  "test",
			RType:  1,
			RClass: 1,
			msg:    msg1,
		},
		expQName:  "test",
		expRType:  1,
		expRClass: 1,
		expMsg:    msg1,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		got = newAnswer(c.msg, c.isLocal)

		if got == nil {
			test.Assert(t, "newAnswer", got, c.exp)
			continue
		}

		if c.isLocal {
			test.Assert(t, "newAnswer.ReceivedAt", int64(0), got.ReceivedAt)
			test.Assert(t, "newAnswer.AccessedAt", int64(0), got.AccessedAt)
		} else {
			test.Assert(t, "newAnswer.ReceivedAt", true, got.ReceivedAt >= at)
			test.Assert(t, "newAnswer.AccessedAt", true, got.AccessedAt >= at)
		}

		test.Assert(t, "newAnswer.QName", c.expQName, got.QName)
		test.Assert(t, "newAnswer.RType", c.expRType, got.RType)
		test.Assert(t, "newAnswer.RClass", c.expRClass, got.RClass)
		test.Assert(t, "newAnswer.msg", c.expMsg, got.msg)
	}
}

func TestAnswerClear(t *testing.T) {
	var (
		msg = NewMessage()
		el  = &list.Element{
			Value: 1,
		}
		an = &Answer{
			msg: msg,
			el:  el,
		}

		expMsg *Message
		expEl  *list.Element
	)

	an.clear()

	test.Assert(t, "answer.msg", expMsg, an.msg)
	test.Assert(t, "answer.el", expEl, an.el)
}

func TestAnswerGet(t *testing.T) {
	type testCase struct {
		msg     *Message
		desc    string
		isLocal bool
	}

	var (
		// kilabit.info A
		res = &Message{
			Header: MessageHeader{
				ID:      1,
				QDCount: 1,
				ANCount: 1,
			},
			Question: MessageQuestion{
				Name:  "kilabit.info",
				Type:  RecordTypeA,
				Class: RecordClassIN,
			},
			Answer: []ResourceRecord{{
				Name:  "kilabit.info",
				Type:  RecordTypeA,
				Class: RecordClassIN,
				TTL:   3600,
				rdlen: 4,
				Value: "127.0.0.1",
			}},
			Authority:  []ResourceRecord{},
			Additional: []ResourceRecord{},
		}
		at = time.Now().Unix()

		cases     []testCase
		c         testCase
		an        *Answer
		got       *Message
		gotPacket []byte
		err       error
	)

	_, err = res.Pack()
	if err != nil {
		t.Fatal("Pack: ", err)
	}

	cases = []testCase{{
		desc:    "With local answer",
		msg:     res,
		isLocal: true,
	}, {
		desc: "With non local answer",
		msg:  res,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		an = newAnswer(c.msg, c.isLocal)

		if !c.isLocal {
			an.ReceivedAt -= 5
		}

		gotPacket = an.get()

		if c.isLocal {
			test.Assert(t, "ReceivedAt", int64(0), an.ReceivedAt)
			test.Assert(t, "AccessedAt", int64(0), an.AccessedAt)
			test.Assert(t, "packet", c.msg.packet, gotPacket)
			continue
		}

		test.Assert(t, "ReceivedAt", an.ReceivedAt >= at-5, true)
		test.Assert(t, "AccessedAt", an.AccessedAt >= at, true)
		got = &Message{
			Header:   MessageHeader{},
			Question: MessageQuestion{},
			packet:   gotPacket,
		}
		err = got.Unpack()
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, "Message.Header", c.msg.Header, got.Header)
		test.Assert(t, "Message.Question", c.msg.Question, got.Question)
		test.Assert(t, "Answer.TTL", c.msg.Answer[0].TTL, got.Answer[0].TTL)
	}
}

func TestAnswerUpdate(t *testing.T) {
	type testCase struct {
		an            *Answer
		nu            *Answer
		expMsg        *Message
		desc          string
		expReceivedAt int64
		expAccessedAt int64
	}

	var (
		at   = time.Now().Unix() - 5
		msg1 = &Message{
			Header: MessageHeader{
				ID: 1,
			},
		}
		msg2 = &Message{
			Header: MessageHeader{
				ID: 1,
			},
		}

		cases []testCase
		c     testCase
	)

	cases = []testCase{{
		desc: "With nil parameter",
		an: &Answer{
			ReceivedAt: 1,
			AccessedAt: 1,
			msg:        msg1,
		},
		expReceivedAt: 1,
		expAccessedAt: 1,
		expMsg:        msg1,
	}, {
		desc: "With local answer",
		an: &Answer{
			ReceivedAt: 0,
			AccessedAt: 0,
			msg:        msg1,
		},
		nu: &Answer{
			ReceivedAt: at,
			AccessedAt: at,
			msg:        msg2,
		},
		expReceivedAt: 0,
		expAccessedAt: 0,
		expMsg:        nil,
	}, {
		desc: "With non local answer",
		an: &Answer{
			ReceivedAt: 1,
			AccessedAt: 1,
			msg:        msg1,
		},
		nu: &Answer{
			ReceivedAt: at,
			AccessedAt: at,
			msg:        msg2,
		},
		expReceivedAt: at,
		expAccessedAt: at,
		expMsg:        nil,
	}}

	for _, c = range cases {
		t.Log(c.desc)

		c.an.update(c.nu)

		test.Assert(t, "ReceivedAt", c.expReceivedAt, c.an.ReceivedAt)
		test.Assert(t, "AccessedAt", c.expAccessedAt, c.an.AccessedAt)
		if c.nu != nil {
			test.Assert(t, "c.nu.msg", c.expMsg, c.nu.msg)
		}
	}
}
