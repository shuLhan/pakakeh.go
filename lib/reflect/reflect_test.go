// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

import (
	"net/http"
	"testing"
)

type F func()

type T struct{}

func (t *T) J() bool {
	return true
}

func TestIsNil(t *testing.T) {
	var (
		aChannel  chan int
		aFunction F
		aMap      map[int]int
		aPtr      *T
		aSlice    []int
		fs        http.FileSystem
	)

	cases := []struct {
		v   interface{}
		exp bool
	}{{
		v: true,
	}, {
		v:   aChannel,
		exp: true,
	}, {
		v:   aFunction,
		exp: true,
	}, {
		v:   aMap,
		exp: true,
	}, {
		v:   aPtr,
		exp: true,
	}, {
		v:   aSlice,
		exp: true,
	}, {
		v:   fs,
		exp: true,
	}}

	for _, c := range cases {
		got := IsNil(c.v)
		if c.exp != got {
			t.Errorf("Expecting %v, got %v on %v(%T)", c.exp, got,
				c.v, c.v)
		}
	}
}
