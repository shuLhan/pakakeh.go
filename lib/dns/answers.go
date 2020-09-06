// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

//
// answers contains list of answer with the same query name but different
// query types.
//
type answers struct {
	v []*answer
}

//
// newAnswers create and initialize list of answer with one element.
//
func newAnswers(an *answer) (ans *answers) {
	ans = &answers{
		v: make([]*answer, 0, 1),
	}
	if an != nil && an.msg != nil {
		ans.v = append(ans.v, an)
	}
	return
}

//
// get an answer with specific query type and class in slice.
// If found, it will return its element and index in slice; otherwise it will
// return nil on answer.
//
func (ans *answers) get(qtype, qclass uint16) (an *answer, x int) {
	for x = 0; x < len(ans.v); x++ {
		if ans.v[x].qtype != qtype {
			continue
		}
		if ans.v[x].qclass != qclass {
			continue
		}

		an = ans.v[x]
		return
	}
	return
}

//
// remove the answer from list.
//
func (ans *answers) remove(qtype, qclass uint16) {
	an, x := ans.get(qtype, qclass)
	if an != nil {
		ans.v[x] = ans.v[len(ans.v)-1]
		ans.v[len(ans.v)-1] = nil
		ans.v = ans.v[:len(ans.v)-1]
	}
}

//
// upsert update or insert new answer to list.
// If new answer is updated, it will return the old answer.
// If new answer is inserted, it will return nil instead.
//
func (ans *answers) upsert(nu *answer) (an *answer) {
	if nu == nil || nu.msg == nil {
		return
	}
	an, _ = ans.get(nu.qtype, nu.qclass)
	if an != nil {
		an.update(nu)
	} else {
		ans.v = append(ans.v, nu)
	}
	return
}
