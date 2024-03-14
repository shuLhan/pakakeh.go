// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestResponseType_String(t *testing.T) {
	cases := []struct {
		exp     string
		restype ResponseType
	}{
		{restype: ``, exp: ``},
		{restype: `binary`, exp: ContentTypeBinary},
		{restype: `html`, exp: ContentTypeHTML},
		{restype: `json`, exp: ContentTypeJSON},
		{restype: `plain`, exp: ContentTypePlain},
		{restype: `xml`, exp: ContentTypeXML},
		{restype: ResponseTypeNone, exp: ""},
		{restype: ResponseTypeBinary, exp: ContentTypeBinary},
		{restype: ResponseTypeHTML, exp: ContentTypeHTML},
		{restype: ResponseTypeJSON, exp: ContentTypeJSON},
		{restype: ResponseTypePlain, exp: ContentTypePlain},
		{restype: ResponseTypeXML, exp: ContentTypeXML},
	}

	for _, c := range cases {
		test.Assert(t, "ResponseType.String", c.exp, c.restype.String())
	}
}
