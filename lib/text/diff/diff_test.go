// Copyright 2018 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"fmt"
	"testing"

	libstrings "github.com/shuLhan/share/lib/strings"
	"github.com/shuLhan/share/lib/text"
)

type DiffExpect struct {
	Adds    []int
	Dels    []int
	Changes []int
}

type DiffExpects []DiffExpect

func testDiffFiles(t *testing.T, old, new string, level int) Data {
	diffs, e := Files(old, new, level)

	if e != nil {
		t.Fatal(e)
	}

	return diffs
}

func compareLineNumber(t *testing.T, diffs Data, exp DiffExpect) {
	if len(exp.Adds) != len(diffs.Adds) {
		t.Fatalf("Expecting adds at %v, got %v", exp.Adds, diffs.Adds)
	} else {
		for x, v := range exp.Adds {
			if diffs.Adds[x].N != v {
				t.Fatalf("Expecting add at %v, got %v", v,
					diffs.Adds[x])
			}
		}
	}

	if len(exp.Dels) != len(diffs.Dels) {
		t.Fatalf("Expecting deletions at %v, got %v", exp.Dels,
			diffs.Dels)
	} else {
		for x, v := range exp.Dels {
			if diffs.Dels[x].N != v {
				t.Fatalf("Expecting deletion at %v, got %v", v,
					diffs.Dels[x])
			}
		}
	}

	if len(exp.Changes) != len(diffs.Changes) {
		t.Fatalf("Expecting changes at %v, got %v", exp.Changes,
			diffs.Changes)
	} else {
		for x, v := range exp.Changes {
			if diffs.Changes[x].Old.N != v {
				t.Fatalf("Expecting change at %v, got %v", v,
					diffs.Changes[x])
			}
		}
	}
}

func TestDiffFilesLevelLine(t *testing.T) {
	diffsExpects := DiffExpects{
		{[]int{}, []int{}, []int{48}},
		{[]int{}, []int{}, []int{48}},
		{[]int{268, 269, 270, 271}, []int{6, 7, 8, 9, 248, 249, 250},
			[]int{}},
		{[]int{6, 7, 8, 9, 248, 249, 250}, []int{268, 269, 270, 271},
			[]int{}},
		{[]int{54}, []int{},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
				15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
				30, 32, 37, 39, 41, 44, 51},
		},
		{[]int{}, []int{54},
			[]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
				15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
				30, 32, 37, 39, 41, 44, 51},
		},
		{[]int{}, []int{5, 6}, []int{}},
		{[]int{5, 6}, []int{}, []int{}},
	}

	oldrev := "testdata/Top_Gear_Series_14.old"
	newrev := "testdata/Top_Gear_Series_14.new"

	diffs := testDiffFiles(t, oldrev, newrev, LevelLines)
	compareLineNumber(t, diffs, diffsExpects[0])

	// reverse test
	diffs = testDiffFiles(t, newrev, oldrev, LevelLines)
	compareLineNumber(t, diffs, diffsExpects[1])

	oldrev = "testdata/List_of_United_Nations.old"
	newrev = "testdata/List_of_United_Nations.new"

	diffs = testDiffFiles(t, oldrev, newrev, LevelLines)
	compareLineNumber(t, diffs, diffsExpects[2])

	// reverse test
	diffs = testDiffFiles(t, newrev, oldrev, LevelLines)
	compareLineNumber(t, diffs, diffsExpects[3])

	oldrev = "testdata/Psusennes_II.old"
	newrev = "testdata/Psusennes_II.new"

	diffs = testDiffFiles(t, oldrev, newrev, LevelLines)
	compareLineNumber(t, diffs, diffsExpects[4])

	diffs = testDiffFiles(t, newrev, oldrev, LevelLines)
	compareLineNumber(t, diffs, diffsExpects[5])

	oldrev = "testdata/empty5lines.txt"
	newrev = "testdata/empty3lines.txt"

	diffs = testDiffFiles(t, oldrev, newrev, LevelLines)
	compareLineNumber(t, diffs, diffsExpects[6])

	diffs = testDiffFiles(t, newrev, oldrev, LevelLines)
	compareLineNumber(t, diffs, diffsExpects[7])
}

func TestDiffFilesLevelWords(t *testing.T) {
	exp_adds := libstrings.Row{
		[]string{"pharaoh"},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"|"},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"|"},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{"| "},
		[]string{" name=\"Kitchen, p.423\""},
		[]string{" name=\"Payraudeau, BIFAO 108, p.294\"", "—",
			"—", " name=\"", "\"/",
		},
		[]string{" name=\"Kitchen, p.290\"", " name=\"", "\"/",
			"–", "—", "—",
		},
		[]string{"—"},
		[]string{
			"—",
			" name=\"Krauss, DE 62, pp.43-48\"",
			" name=\"",
			"\"/",
		},
		[]string{"—", "—", "—", " name=\"", "\"/", "—"},
		[]string{"&nbsp;"},
	}

	exp_dels := libstrings.Row{
		[]string{"Pharaoh ", "| "},
		[]string{"   ", " ", " |"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "  |"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", " |"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", "|"},
		[]string{"   ", " ", " |"},
		[]string{},
		[]string{"--", "--", ">", "</ref"},
		[]string{">", "</ref", "-", "--", "--"},
		[]string{"--"},
		[]string{"--", ">", "</ref"},
		[]string{"--", "--", "--", ">", "</ref", "--"},
		[]string{},
	}

	oldrev := "testdata/text01.old"
	newrev := "testdata/text01.new"

	diffs := testDiffFiles(t, oldrev, newrev, LevelWords)

	compareChunks(t, diffs.Changes[0].Adds, diffs.Changes[0].Dels,
		exp_adds[26], exp_dels[26])

	oldrev = "testdata/text02.old"
	newrev = "testdata/text02.new"

	diffs = testDiffFiles(t, oldrev, newrev, LevelWords)
	compareChunks(t, diffs.Changes[0].Adds, diffs.Changes[0].Dels,
		exp_adds[27], exp_dels[27])

	oldrev = "testdata/Top_Gear_Series_14.old"
	newrev = "testdata/Top_Gear_Series_14.new"

	diffs = testDiffFiles(t, oldrev, newrev, LevelWords)
	compareChunks(t, diffs.Changes[0].Adds, diffs.Changes[0].Dels,
		[]string{","},
		[]string{"alse "},
	)

	oldrev = "testdata/Psusennes_II.old"
	newrev = "testdata/Psusennes_II.new"

	diffs = testDiffFiles(t, oldrev, newrev, LevelWords)
	for x, change := range diffs.Changes {
		if x >= len(exp_adds) {
			break
		}
		compareChunks(t, change.Adds, change.Dels, exp_adds[x],
			exp_dels[x])
	}

	allDels := diffs.Changes.GetAllDels()
	got := allDels.Join("")
	exp := exp_dels.Join("", "")

	if exp != got {
		t.Fatalf("Expecting %s got %s\n", exp, got)
	}

	allAdds := diffs.Changes.GetAllAdds()
	got = allAdds.Join("")
	exp = exp_adds.Join("", "")

	if exp != got {
		t.Fatalf("Expecting %s got %s\n", exp, got)
	}
}

func compareChunks(t *testing.T, adds, dels text.Chunks,
	exp_adds, exp_dels []string,
) {
	if len(adds) != len(exp_adds) {
		t.Fatalf("Expecting adds '%v' got '%v'", exp_adds, adds)
	}
	for x, add := range adds {
		addv := string(add.V)
		if addv != exp_adds[x] {
			t.Fatalf("[%d] Expecting add '%v' got '%v'", x,
				exp_adds[x], addv)
		}
	}

	if len(dels) != len(exp_dels) {
		t.Fatalf("Expecting deletes '%v' got '%v'", exp_dels, dels)
	}
	for x, del := range dels {
		delv := string(del.V)
		if delv != exp_dels[x] {
			t.Fatalf("[%d] Expecting delete '%v' got '%v'", x,
				exp_dels[x], delv)
		}
	}
}

func testDiffLines(t *testing.T, old, new text.Line,
	exp_adds, exp_dels []string) {

	adds, dels := Lines(old.V, new.V, 0, 0)

	compareChunks(t, adds, dels, exp_adds, exp_dels)
}

func TestDiffLines(t *testing.T) {
	old := text.Line{N: 0, V: []byte("lorem ipsum dolmet")}
	new := text.Line{N: 0, V: []byte("lorem all ipsum")}

	exp_adds := libstrings.Row{
		[]string{"all "},
	}
	exp_dels := libstrings.Row{
		[]string{" dolmet"},
	}

	testDiffLines(t, old, new, exp_adds[0], exp_dels[0])

	old = text.Line{N: 0, V: []byte("lorem ipsum dolmet")}
	new = text.Line{N: 0, V: []byte("lorem ipsum")}

	testDiffLines(t, old, new, []string{}, exp_dels[0])

	old = text.Line{N: 0, V: []byte("lorem ipsum")}
	new = text.Line{N: 0, V: []byte("lorem ipsum dolmet")}

	testDiffLines(t, old, new, exp_dels[0], []string{})

	old = text.Line{N: 0, V: []byte("{{Pharaoh Infobox |")}
	new = text.Line{N: 0, V: []byte("{{Infobox pharaoh")}

	testDiffLines(t, old, new, []string{"pharaoh"},
		[]string{"Pharaoh ", "|"})
}

func diffLevelWords(t *testing.T, oldrev, newrev, expdels, expadds string,
	debug bool) {
	diffs := testDiffFiles(t, oldrev, newrev, LevelWords)

	if debug {
		fmt.Printf(">>> diffs:\n%v", diffs)
	}

	allDels := diffs.GetAllDels()
	got := allDels.Join("")

	if !debug && expdels != got {
		t.Fatalf("Expecting '%s' got '%s'\n", expdels, got)
	}

	allAdds := diffs.GetAllAdds()
	got = allAdds.Join("")

	if !debug && expadds != got {
		t.Fatalf("Expecting '%s' got '%s'\n", expadds, got)
	}
}

func TestDiffFilesLevelWords2(t *testing.T) {
	oldrev := "testdata/peeps.old"
	newrev := "testdata/peeps.new"
	expdels := ""
	expadds := "\r\n\r\n== Definitionz!!!?? ==\r\n" +
		"A peep is a person involved in a gang or posse, who which blows.\r\n" +
		"\r\n"

	diffLevelWords(t, oldrev, newrev, expdels, expadds, false)
}

func TestBytesRatio(t *testing.T) {
	old := "# [[...Baby One More Time (song)|...Baby One More Time]]"
	new := "# \"[[...Baby One More Time (song)|...Baby One More Time]]\""

	ratio, _, _ := BytesRatio([]byte(old), []byte(new), DefMatchLen)

	newlen := len(new)
	expMatch := newlen - 2
	expRatio := float32(expMatch) / float32(newlen)

	if expRatio != ratio {
		t.Fatalf("Expecting ratio %f got %f\n", expRatio, ratio)
	}
}

func TestDiffFilesLevelWords3(t *testing.T) {
	oldrev := "testdata/the_singles_collection.old"
	newrev := "testdata/the_singles_collection.new"
	expdels := "\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\""
	expadds := "\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\"\""

	diffLevelWords(t, oldrev, newrev, expdels, expadds, false)
}
