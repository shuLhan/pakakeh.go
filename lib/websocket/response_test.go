// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

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
