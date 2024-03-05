// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package html

import "fmt"

func ExampleNormalizeForID() {
	fmt.Println(NormalizeForID(""))
	fmt.Println(NormalizeForID(" id "))
	fmt.Println(NormalizeForID(" ID "))
	fmt.Println(NormalizeForID("_id.1"))
	fmt.Println(NormalizeForID("1-d"))
	fmt.Println(NormalizeForID(".123 ABC def"))
	fmt.Println(NormalizeForID("test 123"))
	fmt.Println(NormalizeForID("⌘"))
	// Output:
	// _
	// _id_
	// _id_
	// _id_1
	// _1-d
	// _123_abc_def
	// test_123
	// ___
}

func ExampleSanitize() {
	input := `
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
</html>
`

	out := Sanitize([]byte(input))
	fmt.Printf("%s", out)

	// Output:
	// This is a link. An another link.
}
