// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIDPool(t *testing.T) {
	test.Assert(t, "getID()=1", getID(), uint16(1), true)
	test.Assert(t, "getID()=2", getID(), uint16(2), true)
	test.Assert(t, "getID()=3", getID(), uint16(3), true)
	test.Assert(t, "getID()=4", getID(), uint16(4), true)
}
