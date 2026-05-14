// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package diff_test

import (
	"fmt"
	"strings"

	"kilabit.info/pakakeh.go/lib/text/diff"
)

func ExampleUnified() {
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
	dif := diff.Unified([]byte(lao), []byte(tzu))
	var sb strings.Builder
	dif.WriteUnified(&sb, 3)
	fmt.Println(sb.String())

	// Output:
	// --- old
	// +++ new
	// @@ -1,7 +1,6 @@
	// -The Way that can be told of is not the eternal Way;
	// -The name that can be named is not the eternal name.
	//  The Nameless is the origin of Heaven and Earth;
	// -The Named is the mother of all things.
	// +The named is the mother of all things.
	// +
	//  Therefore let there always be non-being,
	//    so we may see their subtlety,
	//  And let there always be being,
	// @@ -9,3 +8,6 @@
	//  The two are the same,
	//  But after they are produced,
	//    they have different names.
	// +They both may be called deep and profound.
	// +Deeper and more profound,
	// +The door of all subtleties!
}
