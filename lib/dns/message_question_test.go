// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestMessageQuestion_String(t *testing.T) {
	cases := []struct {
		desc string
		mq   MessageQuestion
		exp  string
	}{{
		desc: "With unknown type",
		mq: MessageQuestion{
			Name: "test",
			Type: RecordType(17),
		},
		exp: `{Name:test Type:}`,
	}, {
		desc: "With known type",
		mq: MessageQuestion{
			Name: "test",
			Type: RecordTypeA,
		},
		exp: `{Name:test Type:A}`,
	}}

	for _, c := range cases {
		test.Assert(t, c.desc, c.exp, c.mq.String())
	}
}

func TestMessageQuestion_unpack(t *testing.T) {
	cases := []struct {
		desc   string
		mq     MessageQuestion
		packet []byte
		expErr string
	}{{
		desc: "With empty packet",
		mq:   MessageQuestion{},
	}, {
		desc: "With zero label",
		packet: []byte{
			0x00,
			0x00, 0x01,
			0x00, 0x01,
		},
		mq: MessageQuestion{
			Name:  "",
			Type:  RecordTypeA,
			Class: RecordClassIN,
		},
	}, {
		desc: "With invalid label length",
		packet: []byte{
			0x06, 'a',
			0x0,
			0x00, 0x01,
			0x00, 0x01,
		},
		expErr: "MessageQuestion.unpack: label length overflow at index 0",
	}, {
		desc: "With packet too small",
		packet: []byte{
			0x01, 'a',
			0x00,
			0x00, 0x01,
			0x00,
		},
		expErr: "MessageQuestion.unpack: packet too small, missing type and/or class",
	}, {
		desc: "With label",
		packet: []byte{
			0x01, 'a',
			0x01, 'B',
			0x01, 'c',
			0x00,
			0x00, 0x01,
			0x00, 0x01,
		},
		mq: MessageQuestion{
			Name:  "a.b.c",
			Type:  RecordTypeA,
			Class: RecordClassIN,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)
		gotMQ := MessageQuestion{}
		err := gotMQ.unpack(c.packet)
		if err != nil {
			test.Assert(t, c.desc, c.expErr, err.Error())
			continue
		}
		test.Assert(t, c.desc, c.mq, gotMQ)
	}
}
