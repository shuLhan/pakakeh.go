// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package html

import (
	"strings"
	"testing"

	"kilabit.info/pakakeh.go/lib/test"
)

func TestNormalizeForID(t *testing.T) {
	listCase := []struct {
		name string
		exp  string
	}{{
		name: `Test success`,
		exp:  `test_success`,
	}}
	for _, tcase := range listCase {
		name := strings.Clone(tcase.name)
		got := NormalizeForID(tcase.name)
		test.Assert(t, tcase.name, tcase.exp, got)
		// Should not modify the original value.
		test.Assert(t, tcase.name, name, tcase.name)
	}
}
