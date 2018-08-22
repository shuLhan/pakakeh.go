// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIDPool(t *testing.T) {
	test.Assert(t, "getNextID()=1", getNextID(), uint16(1), true)
	test.Assert(t, "getNextID()=2", getNextID(), uint16(2), true)
	test.Assert(t, "getNextID()=3", getNextID(), uint16(3), true)
	test.Assert(t, "getNextID()=4", getNextID(), uint16(4), true)
}
