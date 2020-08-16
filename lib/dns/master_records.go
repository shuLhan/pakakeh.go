// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

//
// masterRecords contains mapping between domain name and its resource
// records.
//
type masterRecords map[string][]*ResourceRecord

func (mr masterRecords) add(rr *ResourceRecord) {
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
