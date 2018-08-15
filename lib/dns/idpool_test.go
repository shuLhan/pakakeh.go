// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestIDPool(t *testing.T) {
	test.Assert(t, "getID()=4", getID(), uint16(4), true)
	test.Assert(t, "getID()=5", getID(), uint16(5), true)
	test.Assert(t, "getID()=6", getID(), uint16(6), true)
	test.Assert(t, "getID()=7", getID(), uint16(7), true)
}
