// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package http

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestRequestType_String(t *testing.T) {
	cases := []struct {
		exp string
		rt  RequestType
	}{
		{rt: RequestTypeNone, exp: ``},
		{rt: RequestTypeQuery, exp: ``},
		{rt: RequestTypeForm, exp: ContentTypeForm},
		{rt: RequestTypeMultipartForm, exp: ContentTypeMultipartForm},
		{rt: RequestTypeJSON, exp: ContentTypeJSON},
		{rt: RequestTypeXML, exp: ContentTypeXML},
	}

	for _, c := range cases {
		test.Assert(t, "RequestMethod.String", c.exp, c.rt.String())
	}
}
