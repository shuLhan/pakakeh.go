// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 Shulhan <ms@kilabit.info>

package diff

import (
	"os"
	"strings"
	"testing"
)

func TestData_Unified(t *testing.T) {
	listCase := []struct {
		desc    string
		oldFile string
		newFile string
		expFile string
	}{{
		desc:    `SingleDeletion`,
		oldFile: `testdata/unified/del.old`,
		newFile: `testdata/unified/del.new`,
		expFile: `testdata/unified/del.uni`,
	}, {
		desc:    `MultiDeletion`,
		oldFile: `testdata/unified/multi_del.old`,
		newFile: `testdata/unified/multi_del.new`,
		expFile: `testdata/unified/multi_del.uni`,
	}, {
		desc:    `SingleAddition`,
		oldFile: `testdata/unified/del.new`,
		newFile: `testdata/unified/del.old`,
		expFile: `testdata/unified/add.uni`,
	}, {
		desc:    `MultiAddition`,
		oldFile: `testdata/unified/multi_del.new`,
		newFile: `testdata/unified/multi_del.old`,
		expFile: `testdata/unified/multi_add.uni`,
	}, {
		desc:    `Mixed`,
		oldFile: `testdata/unified/mixed.old`,
		newFile: `testdata/unified/mixed.new`,
		expFile: `testdata/unified/mixed.uni`,
	}, {
		desc:    `MixedReverse`,
		oldFile: `testdata/unified/mixed.new`,
		newFile: `testdata/unified/mixed.old`,
		expFile: `testdata/unified/mixed_reverse.uni`,
	}, {
		desc:    `EmptyLines`,
		oldFile: `testdata/empty3lines.txt`,
		newFile: `testdata/empty5lines.txt`,
		expFile: `testdata/empty_lines.uni`,
	}, {
		desc:    `EmptyLines.reverse`,
		oldFile: `testdata/empty5lines.txt`,
		newFile: `testdata/empty3lines.txt`,
		expFile: `testdata/empty_lines.reverse.uni`,
	}, {
		desc:    `List_of_United_Nations`,
		oldFile: `testdata/List_of_United_Nations.old`,
		newFile: `testdata/List_of_United_Nations.new`,
		expFile: `testdata/List_of_United_Nations.uni`,
	}, {
		desc:    `List_of_United_Nations.reverse`,
		oldFile: `testdata/List_of_United_Nations.new`,
		newFile: `testdata/List_of_United_Nations.old`,
		expFile: `testdata/List_of_United_Nations.reverse.uni`,
	}, {
		desc:    `peeps`,
		oldFile: `testdata/peeps.old`,
		newFile: `testdata/peeps.new`,
		expFile: `testdata/peeps.uni`,
	}, {
		desc:    `peeps.reverse`,
		oldFile: `testdata/peeps.new`,
		newFile: `testdata/peeps.old`,
		expFile: `testdata/peeps.reverse.uni`,
	}, {
		desc:    `Psusennes_II`,
		oldFile: `testdata/Psusennes_II.old`,
		newFile: `testdata/Psusennes_II.new`,
		expFile: `testdata/Psusennes_II.uni`,
	}, {
		desc:    `Psusennes_II.reverse`,
		oldFile: `testdata/Psusennes_II.new`,
		newFile: `testdata/Psusennes_II.old`,
		expFile: `testdata/Psusennes_II.reverse.uni`,
	}, {
		desc:    `text01`,
		oldFile: `testdata/text01.old`,
		newFile: `testdata/text01.new`,
		expFile: `testdata/text01.uni`,
	}, {
		desc:    `text01.reverse`,
		oldFile: `testdata/text01.new`,
		newFile: `testdata/text01.old`,
		expFile: `testdata/text01.reverse.uni`,
	}, {
		desc:    `text02`,
		oldFile: `testdata/text02.old`,
		newFile: `testdata/text02.new`,
		expFile: `testdata/text02.uni`,
	}, {
		desc:    `text02.reverse`,
		oldFile: `testdata/text02.new`,
		newFile: `testdata/text02.old`,
		expFile: `testdata/text02.reverse.uni`,
	}, {
		desc:    `the_singles_collection`,
		oldFile: `testdata/the_singles_collection.old`,
		newFile: `testdata/the_singles_collection.new`,
		expFile: `testdata/the_singles_collection.uni`,
	}, {
		desc:    `the_singles_collection.reverse`,
		oldFile: `testdata/the_singles_collection.new`,
		newFile: `testdata/the_singles_collection.old`,
		expFile: `testdata/the_singles_collection.reverse.uni`,
	}, {
		desc:    `Top_Gear_Series_14`,
		oldFile: `testdata/Top_Gear_Series_14.old`,
		newFile: `testdata/Top_Gear_Series_14.new`,
		expFile: `testdata/Top_Gear_Series_14.uni`,
	}, {
		desc:    `Top_Gear_Series_14reverse.`,
		oldFile: `testdata/Top_Gear_Series_14.new`,
		newFile: `testdata/Top_Gear_Series_14.old`,
		expFile: `testdata/Top_Gear_Series_14.reverse.uni`,
	}}
	var sbuf strings.Builder
	for _, tc := range listCase {
		t.Run(tc.desc, func(t *testing.T) {
			oldbytes, err := os.ReadFile(tc.oldFile)
			if err != nil {
				t.Fatal(err)
			}
			newbytes, err := os.ReadFile(tc.newFile)
			if err != nil {
				t.Fatal(err)
			}
			expbytes, err := os.ReadFile(tc.expFile)
			if err != nil {
				t.Fatal(err)
			}

			diff := Unified(oldbytes, newbytes)
			sbuf.Reset()
			diff.WriteUnified(&sbuf, 3)
			exp := string(expbytes)
			got := sbuf.String()
			if exp != got {
				t.Errorf("%s: not matched", tc.desc)
				t.Errorf("%s: changes:\n%s", tc.desc, diff.String())
				t.Errorf("%s: want:\n%s", tc.desc, exp)
				t.Errorf("%s: got:\n%s", tc.desc, got)
				t.Fatal()
			}
		})
	}
}
