// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import "github.com/shuLhan/share/lib/reflect"

//
// zoneRecords contains mapping between domain name and its resource
// records.
//
type zoneRecords map[string][]*ResourceRecord

func (mr zoneRecords) add(rr *ResourceRecord) {
	listRR := mr[rr.Name]

	for x, rr2 := range listRR {
		if rr.Type != rr2.Type {
			continue
		}
		if rr.Class != rr2.Class {
			continue
		}

		// Replace the RR if its type is SOA because only one SOA
		// should exist per domain name.
		if rr.Type == QueryTypeSOA {
			listRR[x] = rr
			return
		}
		break
	}
	listRR = append(listRR, rr)
	mr[rr.Name] = listRR
}

//
// remove a ResourceRecord from list of RR by its Name.
// It will return true if the RR exist and removed, otherwise it will return
// false.
//
func (mr zoneRecords) remove(rr *ResourceRecord) bool {
	listRR := mr[rr.Name]
	for x, rr2 := range listRR {
		if rr.Type != rr2.Type {
			continue
		}
		if rr.Class != rr2.Class {
			continue
		}
		if !reflect.IsEqual(rr.Value, rr2.Value) {
			continue
		}
		copy(listRR[x:], listRR[x+1:])
		listRR[len(listRR)-1] = nil
		mr[rr.Name] = listRR[:len(listRR)-1]
		return true
	}
	return false
}
