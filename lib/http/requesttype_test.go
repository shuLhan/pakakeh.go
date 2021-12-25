// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestRequestType_String(t *testing.T) {
	cases := []struct {
		exp string
		rt  RequestType
	}{
		{rt: 0, exp: ""},
		{rt: 1, exp: ""},
		{rt: 2, exp: ContentTypeForm},
		{rt: 3, exp: ContentTypeMultipartForm},
		{rt: 4, exp: ContentTypeJSON},
	}

	for _, c := range cases {
		test.Assert(t, "RequestMethod.String", c.exp, c.rt.String())
	}
}
