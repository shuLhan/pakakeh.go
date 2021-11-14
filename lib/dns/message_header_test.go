// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

type testMessageHeader struct {
	desc   string
	packet []byte
	hdr    MessageHeader
}

func testMessageHeaderCases() []testMessageHeader {
	return []testMessageHeader{{
		desc: "As query",
		hdr: MessageHeader{
			ID:      0xABCD,
			IsQuery: true,
			Op:      OpCodeQuery,
			QDCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x00, 0x00,
			0x00, 0x01,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As inverse query",
		hdr: MessageHeader{
			ID:      0xABCD,
			IsQuery: true,
			Op:      OpCodeIQuery,
			QDCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x08, 0x00,
			0x00, 0x01,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As query server status",
		hdr: MessageHeader{
			ID:      0xABCD,
			IsQuery: true,
			Op:      OpCodeStatus,
		},
		packet: []byte{
			0xab, 0xcd,
			0x10, 0x00,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As query, RD=1",
		hdr: MessageHeader{
			ID:      0xABCD,
			IsQuery: true,
			Op:      OpCodeQuery,
			IsRD:    true,
			QDCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x01, 0x00,
			0x00, 0x01,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			QDCount: 1,
			ANCount: 0x04,
		},
		packet: []byte{
			0xab, 0xcd,
			0x80, 0x00,
			0x00, 0x01,
			0x00, 0x04,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer, RCode=5",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			RCode:   RCodeRefused,
			QDCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x80, 0x05,
			0x00, 0x01,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer, IsAA=1",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			IsAA:    true,
			QDCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x84, 0x00,
			0x00, 0x01,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer, IsTC=1",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			IsTC:    true,
			QDCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x82, 0x00,
			0x00, 0x01,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer, IsRD=1",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			IsRD:    true,
			QDCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x81, 0x00,
			0x00, 0x01,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer, IsTC=1 IsRD=1",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			IsTC:    true,
			IsRD:    true,
			QDCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x83, 0x00,
			0x00, 0x01,
			0x00, 0x00,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer, IsAA=1 IsRD=1",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			IsAA:    true,
			IsRD:    true,
			QDCount: 1,
			ANCount: 3,
		},
		packet: []byte{
			0xab, 0xcd,
			0x85, 0x00,
			0x00, 0x01,
			0x00, 0x03,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer, IsAA=1 IsTC=1 IsRD=1",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			IsAA:    true,
			IsTC:    true,
			IsRD:    true,
			QDCount: 1,
			ANCount: 3,
		},
		packet: []byte{
			0xab, 0xcd,
			0x87, 0x00,
			0x00, 0x01,
			0x00, 0x03,
			0x00, 0x00,
			0x00, 0x00,
		},
	}, {
		desc: "As answer, IsAA=1 IsRD=1 all counts",
		hdr: MessageHeader{
			ID:      0xABCD,
			Op:      OpCodeQuery,
			IsAA:    true,
			IsRD:    true,
			QDCount: 1,
			ANCount: 4,
			NSCount: 1,
			ARCount: 1,
		},
		packet: []byte{
			0xab, 0xcd,
			0x85, 0x00,
			0x00, 0x01,
			0x00, 0x04,
			0x00, 0x01,
			0x00, 0x01,
		},
	}}
}

func TestMessageHeader_pack(t *testing.T) {
	cases := testMessageHeaderCases()

	for _, c := range cases {
		got := c.hdr.pack()
		test.Assert(t, c.desc, c.packet, got)
	}
}

func TestMessageHeader_unpack(t *testing.T) {
	cases := testMessageHeaderCases()

	for _, c := range cases {
		got := MessageHeader{}
		got.unpack(c.packet)
		test.Assert(t, c.desc, c.hdr, got)
	}
}
