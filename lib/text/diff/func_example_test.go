// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package diff

import (
	"fmt"
)

func ExampleBytes() {
	fmt.Println(`Additions 1:`)
	adds, dels := Bytes([]byte(`00000`), []byte(`00000111`), 0, 0)
	fmt.Printf("  adds: %+v\n", adds)
	fmt.Printf("  dels: %+v\n", dels)

	fmt.Println(`Additions 2:`)
	adds, dels = Bytes([]byte(`000000`), []byte(`0001000`), 0, 0)
	fmt.Printf("  adds: %+v\n", adds)
	fmt.Printf("  dels: %+v\n", dels)

	fmt.Println(`Deletions 1:`)
	adds, dels = Bytes([]byte(`00000111`), []byte(`00000`), 0, 0)
	fmt.Printf("  adds: %+v\n", adds)
	fmt.Printf("  dels: %+v\n", dels)

	fmt.Println(`Deletions 2:`)
	adds, dels = Bytes([]byte(`0001000`), []byte(`000000`), 0, 0)
	fmt.Printf("  adds: %+v\n", adds)
	fmt.Printf("  dels: %+v\n", dels)

	fmt.Println(`Both changes 1:`)
	adds, dels = Bytes([]byte(`11000`), []byte(`22000`), 0, 0)
	fmt.Printf("  adds: %+v\n", adds)
	fmt.Printf("  dels: %+v\n", dels)

	fmt.Println(`Both changes 2:`)
	adds, dels = Bytes([]byte(`0001000`), []byte(`0002000`), 0, 0)
	fmt.Printf("  adds: %+v\n", adds)
	fmt.Printf("  dels: %+v\n", dels)

	// Output:
	// Additions 1:
	//   adds: [{StartAt:5,V:111}]
	//   dels: []
	// Additions 2:
	//   adds: [{StartAt:3,V:1}]
	//   dels: []
	// Deletions 1:
	//   adds: []
	//   dels: [{StartAt:5,V:111}]
	// Deletions 2:
	//   adds: []
	//   dels: [{StartAt:3,V:1}]
	// Both changes 1:
	//   adds: [{StartAt:0,V:22}]
	//   dels: [{StartAt:0,V:11}]
	// Both changes 2:
	//   adds: [{StartAt:3,V:2}]
	//   dels: [{StartAt:3,V:1}]
}

func ExampleBytesRatio() {
	ratio, m, maxlen := BytesRatio([]byte(`00000`), []byte(`00000111`), 1)
	fmt.Printf("Additions 1: %f %d %d\n", ratio, m, maxlen)

	ratio, m, maxlen = BytesRatio([]byte(`000000`), []byte(`0001000`), 1)
	fmt.Printf("Additions 2: %f %d %d\n", ratio, m, maxlen)

	ratio, m, maxlen = BytesRatio([]byte(`00000111`), []byte(`00000`), 1)
	fmt.Printf("Deletions 1: %f %d %d\n", ratio, m, maxlen)

	ratio, m, maxlen = BytesRatio([]byte(`0001000`), []byte(`000000`), 1)
	fmt.Printf("Deletions 2: %f %d %d\n", ratio, m, maxlen)

	ratio, m, maxlen = BytesRatio([]byte(`11000`), []byte(`22000`), 1)
	fmt.Printf("Both changes 1: %f %d %d\n", ratio, m, maxlen)

	ratio, m, maxlen = BytesRatio([]byte(`0001000`), []byte(`0002000`), 1)
	fmt.Printf("Both changes 2: %f %d %d\n", ratio, m, maxlen)

	// Output:
	// Additions 1: 0.625000 5 8
	// Additions 2: 0.857143 6 7
	// Deletions 1: 0.625000 5 8
	// Deletions 2: 0.857143 6 7
	// Both changes 1: 0.600000 3 5
	// Both changes 2: 0.857143 6 7
}
