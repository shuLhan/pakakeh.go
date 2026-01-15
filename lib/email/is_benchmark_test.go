// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package email

import "testing"

func BenchmarkIsValidLocal(b *testing.B) {
	var (
		valid   = []byte(`a-quite-long-local-name`)
		invalid = []byte(`a-quite-long-local-name]`)
	)
	for x := 0; x < b.N; x++ {
		IsValidLocal(valid)
		IsValidLocal(invalid)
	}
}
