// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func testGenerateZoneRecords() (zoneRR zoneRecords, listRR []*ResourceRecord) {
	var (
		rr *ResourceRecord
	)

	zoneRR = zoneRecords{}

	listRR = []*ResourceRecord{{
		Name:  "test",
		Type:  RecordTypeA,
		Class: RecordClassIN,
		Value: "127.0.0.1",
		TTL:   1,
	}, {
		Name:  "test",
		Type:  RecordTypeSOA,
		Class: RecordClassIN,
		Value: &RDataSOA{},
		TTL:   2,
	}, {
		Name:  "test",
		Type:  RecordTypeMX,
		Class: RecordClassIN,
		TTL:   3,
	}, {
		Name:  "test",
		Type:  RecordTypeSOA,
		Class: RecordClassIN,
		TTL:   4,
	}, {
		Name:  "test",
		Type:  RecordTypeA,
		Class: RecordClassCH,
		TTL:   5,
	}}

	for _, rr = range listRR {
		zoneRR.add(rr)
	}

	return zoneRR, listRR
}

func TestZoneRecords_add(t *testing.T) {
	var (
		expZoneRR zoneRecords
		gotZoneRR zoneRecords
		listRR    []*ResourceRecord
	)

	gotZoneRR, listRR = testGenerateZoneRecords()

	expZoneRR = zoneRecords{
		"test": []*ResourceRecord{
			listRR[0],
			listRR[3],
			listRR[2],
			listRR[4],
		},
	}

	test.Assert(t, "add", expZoneRR, gotZoneRR)
}

func TestZoneRecords_remove(t *testing.T) {
	type testCase struct {
		rr           *ResourceRecord
		expZoneRR    zoneRecords
		expIsRemoved bool
	}

	var (
		gotZoneRR    zoneRecords
		listRR       []*ResourceRecord
		cases        []testCase
		c            testCase
		gotIsRemoved bool
	)

	gotZoneRR, listRR = testGenerateZoneRecords()

	cases = []testCase{{
		// With different value.
		rr: &ResourceRecord{
			Name:  "test",
			Type:  RecordTypeA,
			Class: RecordClassIN,
			Value: "127.0.0.2",
		},
		expZoneRR:    gotZoneRR,
		expIsRemoved: false,
	}, {
		// With different Class.
		rr: &ResourceRecord{
			Name:  "test",
			Type:  RecordTypeA,
			Class: RecordClassCH,
			Value: "127.0.0.1",
		},
		expZoneRR:    gotZoneRR,
		expIsRemoved: false,
	}, {
		// With RR removed at the end.
		rr: listRR[4],
		expZoneRR: zoneRecords{
			"test": []*ResourceRecord{
				listRR[0],
				listRR[3],
				listRR[2],
			},
		},
		expIsRemoved: true,
	}}

	for _, c = range cases {
		gotIsRemoved = gotZoneRR.remove(c.rr)
		test.Assert(t, "is removed", c.expIsRemoved, gotIsRemoved)
		test.Assert(t, "after removed", c.expZoneRR, gotZoneRR)
	}
}
