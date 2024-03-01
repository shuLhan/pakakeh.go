// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestResponseReset(t *testing.T) {
	var res = _resPool.Get().(*Response)

	res.reset()

	test.Assert(t, "Response.ID", uint64(0), res.ID)
	test.Assert(t, "Response.Code", int32(0), res.Code)
	test.Assert(t, "Response.Message", "", res.Message)
	test.Assert(t, "Response.Body", "", res.Body)
}
