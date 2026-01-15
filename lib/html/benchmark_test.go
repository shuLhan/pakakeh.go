// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

package html

import "testing"

func BenchmarkNormalizeForID(b *testing.B) {
	var (
		cases = []string{
			"",
			".123 ABC def",
		}
		x int
	)
	for ; x < b.N; x++ {
		NormalizeForID(cases[0])
		NormalizeForID(cases[1])
	}
}

func BenchmarkSanitize(b *testing.B) {
	var (
		input = []byte(`
<html>
	<title>Test</title>
	<head>
	</head>
	<body>
		This
		<p> is </p>
		a
		<a href="/">link</a>.
		An another
		<a href="/">link</a>.
	</body>
</html>`)
		x int
	)

	for ; x < b.N; x++ {
		Sanitize(input)
	}
}
