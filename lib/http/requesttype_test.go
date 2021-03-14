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
		rt  RequestType
		exp string
	}{
		{0, ""},
		{1, ""},
		{2, ContentTypeForm},
		{3, ContentTypeMultipartForm},
		{4, ContentTypeJSON},
	}

	for _, c := range cases {
		test.Assert(t, "RequestMethod.String", c.exp, c.rt.String(), true)
	}
}
