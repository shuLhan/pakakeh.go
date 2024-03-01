// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestRecordType(t *testing.T) {
	test.Assert(t, "RecordTypeA", RecordTypeA, RecordType(1))
	test.Assert(t, "RecordTypeTXT", RecordTypeTXT, RecordType(16))
	test.Assert(t, "RecordTypeAXFR", RecordTypeAXFR, RecordType(252))
	test.Assert(t, "RecordTypeALL", RecordTypeALL, RecordType(255))
}
