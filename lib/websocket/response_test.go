// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestResponseReset(t *testing.T) {
	res := _resPool.Get().(*Response)

	res.Reset()

	test.Assert(t, "Response.ID", uint64(0), res.ID, true)
	test.Assert(t, "Response.Code", int32(0), res.Code, true)
	test.Assert(t, "Response.Message", "", res.Message, true)
	test.Assert(t, "Response.Body", "", res.Body, true)
}
