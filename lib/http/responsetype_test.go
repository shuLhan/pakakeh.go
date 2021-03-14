// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestResponseType_String(t *testing.T) {
	cases := []struct {
		restype ResponseType
		exp     string
	}{
		{0, ""},
		{1, ContentTypeBinary},
		{2, ContentTypeHTML},
		{3, ContentTypeJSON},
		{4, ContentTypePlain},
		{5, ContentTypeXML},
		{ResponseTypeNone, ""},
		{ResponseTypeBinary, ContentTypeBinary},
		{ResponseTypeHTML, ContentTypeHTML},
		{ResponseTypeJSON, ContentTypeJSON},
		{ResponseTypePlain, ContentTypePlain},
		{ResponseTypeXML, ContentTypeXML},
	}

	for _, c := range cases {
		test.Assert(t, "ResponseType.String", c.exp, c.restype.String())
	}
}
