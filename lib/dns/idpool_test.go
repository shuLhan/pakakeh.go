// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIDPool(t *testing.T) {
	test.Assert(t, "getNextID()=5", getNextID(), uint16(5), true)
	test.Assert(t, "getNextID()=6", getNextID(), uint16(6), true)
	test.Assert(t, "getNextID()=7", getNextID(), uint16(7), true)
	test.Assert(t, "getNextID()=8", getNextID(), uint16(8), true)
}
