// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package dns

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestIDPool(t *testing.T) {
	resetIDPool()
	test.Assert(t, `getNextID()=1`, getNextID(), uint16(1))
	test.Assert(t, `getNextID()=2`, getNextID(), uint16(2))
	test.Assert(t, `getNextID()=3`, getNextID(), uint16(3))
	test.Assert(t, `getNextID()=4`, getNextID(), uint16(4))
}
