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
		m   RequestMethod
		exp string
	}{
		{0, "GET"},
		{1, "CONNECT"},
		{2, "DELETE"},
		{3, "HEAD"},
		{4, "OPTIONS"},
		{5, "PATCH"},
		{6, "POST"},
		{7, "PUT"},
		{8, "TRACE"},
		{9, ""},
		{RequestMethodGet, "GET"},
		{RequestMethodConnect, "CONNECT"},
		{RequestMethodDelete, "DELETE"},
		{RequestMethodHead, "HEAD"},
		{RequestMethodOptions, "OPTIONS"},
		{RequestMethodPatch, "PATCH"},
		{RequestMethodPost, "POST"},
		{RequestMethodPut, "PUT"},
		{RequestMethodTrace, "TRACE"},
	}

	for _, c := range cases {
		test.Assert(t, "RequestMethod.String", c.exp, c.m.String(), true)
	}
}
