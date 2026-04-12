// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package diff_test

import (
	"fmt"
	"log"

	"git.sr.ht/~shulhan/pakakeh.go/lib/text/diff"
)

func ExampleFiles() {
	diff, err := diff.Files(`testdata/lao.txt`, `testdata/tzu.txt`, diff.LevelLines)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(diff.String())

	// Output:
	// --- testdata/lao.txt
	// +++ testdata/tzu.txt
	// ----
	//    1/   1: -"The Way that can be told of is not the eternal Way;"
	//    2/   2: -"The name that can be named is not the eternal name."
	//    2/   4: -"The Named is the mother of all things."
	// ++++
	//    4/   2: +"The named is the mother of all things."
	//    3/   3: +""
	//   11/  11: +"They both may be called deep and profound."
	//   12/  12: +"Deeper and more profound,"
	//   13/  13: +"The door of all subtleties!"
}

func ExampleLines() {
	lao := `The Way that can be told of is not the eternal Way;
The name that can be named is not the eternal name.
The Nameless is the origin of Heaven and Earth;
The Named is the mother of all things.
Therefore let there always be non-being,
  so we may see their subtlety,
And let there always be being,
  so we may see their outcome.
The two are the same,
But after they are produced,
  they have different names.`

	tzu := `The Nameless is the origin of Heaven and Earth;
The named is the mother of all things.

Therefore let there always be non-being,
  so we may see their subtlety,
And let there always be being,
  so we may see their outcome.
The two are the same,
But after they are produced,
  they have different names.
They both may be called deep and profound.
Deeper and more profound,
The door of all subtleties!
`
	oldlines := diff.ParseLines([]byte(lao))
	newlines := diff.ParseLines([]byte(tzu))
	diff := diff.Lines(oldlines, newlines, diff.LevelLines)
	fmt.Println(diff.String())

	// Output:
	// --- oldlines
	// +++ newlines
	// ----
	//    1/   1: -"The Way that can be told of is not the eternal Way;"
	//    2/   2: -"The name that can be named is not the eternal name."
	//    2/   4: -"The Named is the mother of all things."
	// ++++
	//    4/   2: +"The named is the mother of all things."
	//    3/   3: +""
	//   11/  11: +"They both may be called deep and profound."
	//   12/  12: +"Deeper and more profound,"
	//   13/  13: +"The door of all subtleties!"
}

func ExampleText() {
	lao := `The Way that can be told of is not the eternal Way;
The name that can be named is not the eternal name.
The Nameless is the origin of Heaven and Earth;
The Named is the mother of all things.
Therefore let there always be non-being,
  so we may see their subtlety,
And let there always be being,
  so we may see their outcome.
The two are the same,
But after they are produced,
  they have different names.`

	tzu := `The Nameless is the origin of Heaven and Earth;
The named is the mother of all things.

Therefore let there always be non-being,
  so we may see their subtlety,
And let there always be being,
  so we may see their outcome.
The two are the same,
But after they are produced,
  they have different names.
They both may be called deep and profound.
Deeper and more profound,
The door of all subtleties!
`
	diff := diff.Text([]byte(lao), []byte(tzu), diff.LevelLines)
	fmt.Println(diff.String())

	// Output:
	// --- old
	// +++ new
	// ----
	//    1/   1: -"The Way that can be told of is not the eternal Way;"
	//    2/   2: -"The name that can be named is not the eternal name."
	//    2/   4: -"The Named is the mother of all things."
	// ++++
	//    4/   2: +"The named is the mother of all things."
	//    3/   3: +""
	//   11/  11: +"They both may be called deep and profound."
	//   12/  12: +"Deeper and more profound,"
	//   13/  13: +"The door of all subtleties!"
}
