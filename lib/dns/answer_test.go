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
	at := time.Now().Unix()

	msg1 := &Message{
		Header: SectionHeader{
			ID: 1,
		},
		Question: SectionQuestion{
			Name:  []byte("test"),
			Type:  1,
			Class: 1,
		},
		Answer: []ResourceRecord{{
			Name:  []byte("test"),
			Type:  QueryTypeA,
			Class: QueryClassIN,
			TTL:   3600,
			rdlen: 4,
			Text: &RDataText{
				Value: []byte("127.0.0.1"),
			},
		}},
	}

	cases := []struct {
		desc      string
		msg       *Message
		exp       *answer
		expMsg    *Message
		expQName  string
		expQType  uint16
		expQClass uint16
		isLocal   bool
	}{{
		desc:    "With local message",
		msg:     msg1,
		isLocal: true,
		exp: &answer{
			qname:  "test",
			qtype:  1,
			qclass: 1,
			msg:    msg1,
		},
		expQName:  "test",
		expQType:  1,
		expQClass: 1,
		expMsg:    msg1,
	}, {
		desc: "With non local message",
		msg:  msg1,
		exp: &answer{
			qname:  "test",
			qtype:  1,
			qclass: 1,
			msg:    msg1,
		},
		expQName:  "test",
		expQType:  1,
		expQClass: 1,
		expMsg:    msg1,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got := newAnswer(c.msg, c.isLocal)

		if got == nil {
			test.Assert(t, "newAnswer", got, c.exp, true)
			continue
		}

		if c.isLocal {
			test.Assert(t, "newAnswer.receivedAt", int64(0), got.receivedAt, true)
			test.Assert(t, "newAnswer.accessedAt", int64(0), got.accessedAt, true)
		} else {
			test.Assert(t, "newAnswer.receivedAt", true, got.receivedAt >= at, true)
			test.Assert(t, "newAnswer.accessedAt", true, got.accessedAt >= at, true)
		}

		test.Assert(t, "newAnswer.qname", c.expQName, got.qname, true)
		test.Assert(t, "newAnswer.qtype", c.expQType, got.qtype, true)
		test.Assert(t, "newAnswer.qclass", c.expQClass, got.qclass, true)
		test.Assert(t, "newAnswer.msg", c.expMsg, got.msg, true)
	}
}

func TestAnswerClear(t *testing.T) {
	msg := NewMessage()
	el := &list.Element{
		Value: 1,
	}

	an := &answer{
		msg: msg,
		el:  el,
	}

	an.clear()

	var expMsg *Message
	var expEl *list.Element

	test.Assert(t, "answer.msg", expMsg, an.msg, true)
	test.Assert(t, "answer.el", expEl, an.el, true)
}

func TestAnswerGet(t *testing.T) {
	// kilabit.info A
	res := &Message{
		Header: SectionHeader{
			ID:      1,
			QDCount: 1,
			ANCount: 1,
		},
		Question: SectionQuestion{
			Name:  []byte("kilabit.info"),
			Type:  QueryTypeA,
			Class: QueryClassIN,
		},
		Answer: []ResourceRecord{{
			Name:  []byte("kilabit.info"),
			Type:  QueryTypeA,
			Class: QueryClassIN,
			TTL:   3600,
			rdlen: 4,
			Text: &RDataText{
				Value: []byte("127.0.0.1"),
			},
		}},
		Authority:  []ResourceRecord{},
		Additional: []ResourceRecord{},
	}

	_, err := res.Pack()
	if err != nil {
		t.Fatal("Pack: ", err)
	}

	at := time.Now().Unix()

	cases := []struct {
		desc    string
		msg     *Message
		isLocal bool
	}{{
		desc:    "With local answer",
		msg:     res,
		isLocal: true,
	}, {
		desc: "With non local answer",
		msg:  res,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		an := newAnswer(c.msg, c.isLocal)

		if !c.isLocal {
			an.receivedAt -= 5
		}

		gotPacket := an.get()

		if c.isLocal {
			test.Assert(t, "receivedAt", int64(0), an.receivedAt, true)
			test.Assert(t, "accessedAt", int64(0), an.accessedAt, true)
			test.Assert(t, "packet", c.msg.Packet, gotPacket, true)
			continue
		}

		test.Assert(t, "receivedAt", an.receivedAt >= at-5, true, true)
		test.Assert(t, "accessedAt", an.accessedAt >= at, true, true)
		got := &Message{
			Header:   SectionHeader{},
			Question: SectionQuestion{},
			Packet:   gotPacket,
		}
		err := got.Unpack()
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, "Message.Header", c.msg.Header, got.Header, true)
		test.Assert(t, "Message.Question", c.msg.Question, got.Question, true)
		test.Assert(t, "Answer.TTL", c.msg.Answer[0].TTL, got.Answer[0].TTL, true)
	}
}

func TestAnswerUpdate(t *testing.T) {
	at := time.Now().Unix() - 5
	msg1 := &Message{
		Header: SectionHeader{
			ID: 1,
		},
	}
	msg2 := &Message{
		Header: SectionHeader{
			ID: 1,
		},
	}

	cases := []struct {
		desc          string
		an            *answer
		nu            *answer
		expReceivedAt int64
		expAccessedAt int64
		expMsg        *Message
	}{{
		desc: "With nil parameter",
		an: &answer{
			receivedAt: 1,
			accessedAt: 1,
			msg:        msg1,
		},
		expReceivedAt: 1,
		expAccessedAt: 1,
		expMsg:        msg1,
	}, {
		desc: "With local answer",
		an: &answer{
			receivedAt: 0,
			accessedAt: 0,
			msg:        msg1,
		},
		nu: &answer{
			receivedAt: at,
			accessedAt: at,
			msg:        msg2,
		},
		expReceivedAt: 0,
		expAccessedAt: 0,
		expMsg:        nil,
	}, {
		desc: "With non local answer",
		an: &answer{
			receivedAt: 1,
			accessedAt: 1,
			msg:        msg1,
		},
		nu: &answer{
			receivedAt: at,
			accessedAt: at,
			msg:        msg2,
		},
		expReceivedAt: at,
		expAccessedAt: at,
		expMsg:        nil,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		c.an.update(c.nu)

		test.Assert(t, "receivedAt", c.expReceivedAt, c.an.receivedAt, true)
		test.Assert(t, "accessedAt", c.expAccessedAt, c.an.accessedAt, true)
		if c.nu != nil {
			test.Assert(t, "c.nu.msg", c.expMsg, c.nu.msg, true)
		}
	}
}
