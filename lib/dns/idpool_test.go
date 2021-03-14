// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIDPool(t *testing.T) {
	test.Assert(t, "getNextID()=8", getNextID(), uint16(8))
	test.Assert(t, "getNextID()=9", getNextID(), uint16(9))
	test.Assert(t, "getNextID()=10", getNextID(), uint16(10))
	test.Assert(t, "getNextID()=11", getNextID(), uint16(11))
}
