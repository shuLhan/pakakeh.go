// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

import (
	"reflect"
	"testing"
)

func TestAppendSlice(t *testing.T) {
	var (
		v123  = 123
		v456  = 456
		pv123 = &v123
		pv456 = &v456

		sliceT       []int
		slicePtrT    []*int
		slicePtrPtrT []**int
		sliceBytes   [][]byte
	)

	type testCase struct {
		obj  interface{}
		exp  interface{}
		desc string
		vals []string
	}

	var (
		cases = []testCase{{
			desc: "setSlice []int",
			obj:  sliceT,
			exp:  []int{v123, v456},
			vals: []string{"123", "456"},
		}, {
			desc: "setSlice []*int",
			obj:  slicePtrT,
			exp:  []*int{&v123, &v456},
			vals: []string{"123", "456"},
		}, {
			desc: "setSlice []**int",
			obj:  slicePtrPtrT,
			exp:  []**int{&pv123, &pv456},
			vals: []string{"123", "456"},
		}, {
			desc: "setSlice [][]byte",
			obj:  sliceBytes,
			exp:  [][]byte{[]byte("123"), []byte("456")},
			vals: []string{"123", "456"},
		}}

		got reflect.Value
		c   testCase
		val string
		err error
	)
	for _, c = range cases {
		t.Log(c.desc)

		got = reflect.ValueOf(c.obj)
		for _, val = range c.vals {
			got, err = setSlice(got, val)
			if err != nil {
				t.Fatal(err)
			}
		}

		if !reflect.DeepEqual(c.exp, got.Interface()) {
			t.Fatalf("expecting %v, got %v", c.exp, c.obj)
		}
	}
}
