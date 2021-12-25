// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestRequestMethod_String(t *testing.T) {
	cases := []struct {
		exp string
		m   RequestMethod
	}{
		{m: 0, exp: "GET"},
		{m: 1, exp: "CONNECT"},
		{m: 2, exp: "DELETE"},
		{m: 3, exp: "HEAD"},
		{m: 4, exp: "OPTIONS"},
		{m: 5, exp: "PATCH"},
		{m: 6, exp: "POST"},
		{m: 7, exp: "PUT"},
		{m: 8, exp: "TRACE"},
		{m: 9, exp: ""},
		{m: RequestMethodGet, exp: "GET"},
		{m: RequestMethodConnect, exp: "CONNECT"},
		{m: RequestMethodDelete, exp: "DELETE"},
		{m: RequestMethodHead, exp: "HEAD"},
		{m: RequestMethodOptions, exp: "OPTIONS"},
		{m: RequestMethodPatch, exp: "PATCH"},
		{m: RequestMethodPost, exp: "POST"},
		{m: RequestMethodPut, exp: "PUT"},
		{m: RequestMethodTrace, exp: "TRACE"},
	}

	for _, c := range cases {
		test.Assert(t, "RequestMethod.String", c.exp, c.m.String())
	}
}
