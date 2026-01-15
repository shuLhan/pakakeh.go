// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package dns

import (
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestRecordType(t *testing.T) {
	test.Assert(t, "RecordTypeA", RecordTypeA, RecordType(1))
	test.Assert(t, "RecordTypeTXT", RecordTypeTXT, RecordType(16))
	test.Assert(t, "RecordTypeAXFR", RecordTypeAXFR, RecordType(252))
	test.Assert(t, `RecordTypeANY`, RecordTypeANY, RecordType(255))
}
