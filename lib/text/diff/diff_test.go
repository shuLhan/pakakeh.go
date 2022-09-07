// Copyright 2018 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"os"
	"reflect"
	"testing"

	libstrings "github.com/shuLhan/share/lib/strings"
	"github.com/shuLhan/share/lib/text"
)

func compareChunks(t *testing.T, adds, dels text.Chunks, expAdds, expDels []string) {
	if len(adds) != len(expAdds) {
		t.Fatalf(`Expecting adds '%v' got '%v'`, expAdds, adds)
	}

	var (
		x     int
		chunk text.Chunk
		vstr  string
	)

	for x, chunk = range adds {
		vstr = string(chunk.V)
		if vstr != expAdds[x] {
			t.Fatalf(`[%d] Expecting add '%v' got '%v'`, x, expAdds[x], vstr)
		}
	}

	if len(dels) != len(expDels) {
		t.Fatalf(`Expecting deletes '%v' got '%v'`, expDels, dels)
	}

	for x, chunk = range dels {
		vstr = string(chunk.V)
		if vstr != expDels[x] {
			t.Fatalf(`[%d] Expecting delete '%v' got '%v'`, x, expDels[x], vstr)
		}
	}
}

func testDiffBytes(t *testing.T, old, new text.Line,
	expAdds, expDels []string,
) {
	var (
		adds, dels text.Chunks = Bytes(old.V, new.V, 0, 0)
	)

	compareChunks(t, adds, dels, expAdds, expDels)
}

func TestBytes(t *testing.T) {
	var (
		old = text.Line{N: 0, V: []byte(`lorem ipsum dolmet`)}
		new = text.Line{N: 0, V: []byte(`lorem all ipsum`)}

		expAdds = libstrings.Row{
			[]string{`all `},
		}
		expDels = libstrings.Row{
			[]string{` dolmet`},
		}
	)

	testDiffBytes(t, old, new, expAdds[0], expDels[0])

	old = text.Line{N: 0, V: []byte(`lorem ipsum dolmet`)}
	new = text.Line{N: 0, V: []byte(`lorem ipsum`)}

	testDiffBytes(t, old, new, []string{}, expDels[0])

	old = text.Line{N: 0, V: []byte(`lorem ipsum`)}
	new = text.Line{N: 0, V: []byte(`lorem ipsum dolmet`)}

	testDiffBytes(t, old, new, expDels[0], []string{})

	old = text.Line{N: 0, V: []byte(`{{Pharaoh Infobox |`)}
	new = text.Line{N: 0, V: []byte(`{{Infobox pharaoh`)}

	testDiffBytes(t, old, new, []string{`pharaoh`}, []string{`Pharaoh `, `|`})
}

func TestBytesRatio(t *testing.T) {
	var (
		old      = `# [[...Baby One More Time (song)|...Baby One More Time]]`
		new      = `# "[[...Baby One More Time (song)|...Baby One More Time]]"`
		newlen   = len(new)
		expMatch = newlen - 2
		expRatio = float32(expMatch) / float32(newlen)

		ratio float32
	)

	ratio, _, _ = BytesRatio([]byte(old), []byte(new), DefMatchLen)

	if expRatio != ratio {
		t.Fatalf(`Expecting ratio %f got %f`, expRatio, ratio)
	}
}

func TestText(t *testing.T) {
	type testCase struct {
		textBefore     string
		textAfter      string
		diffLevelLines string
		diffLevelWords string
	}

	var cases = []testCase{{
		textBefore:     `testdata/List_of_United_Nations.old`,
		textAfter:      `testdata/List_of_United_Nations.new`,
		diffLevelLines: `testdata/List_of_United_Nations_diff_LevelLines`,
		diffLevelWords: `testdata/List_of_United_Nations_diff_LevelWords`,
	}, {
		textBefore:     `testdata/List_of_United_Nations.new`,
		textAfter:      `testdata/List_of_United_Nations.old`,
		diffLevelLines: `testdata/List_of_United_Nations_diff_LevelLines_reverse`,
		diffLevelWords: `testdata/List_of_United_Nations_diff_LevelWords_reverse`,
	}, {
		textBefore:     `testdata/Psusennes_II.old`,
		textAfter:      `testdata/Psusennes_II.new`,
		diffLevelLines: `testdata/Psusennes_II_diff_LevelLines`,
		diffLevelWords: `testdata/Psusennes_II_diff_LevelWords`,
	}, {
		textBefore:     `testdata/Psusennes_II.new`,
		textAfter:      `testdata/Psusennes_II.old`,
		diffLevelLines: `testdata/Psusennes_II_diff_LevelLines_reverse`,
		diffLevelWords: `testdata/Psusennes_II_diff_LevelWords_reverse`,
	}, {
		textBefore:     `testdata/Top_Gear_Series_14.old`,
		textAfter:      `testdata/Top_Gear_Series_14.new`,
		diffLevelLines: `testdata/Top_Gear_Series_14_diff_LevelLines`,
		diffLevelWords: `testdata/Top_Gear_Series_14_diff_LevelWords`,
	}, {
		textBefore:     `testdata/Top_Gear_Series_14.new`,
		textAfter:      `testdata/Top_Gear_Series_14.old`,
		diffLevelLines: `testdata/Top_Gear_Series_14_diff_LevelLines_reverse`,
		diffLevelWords: `testdata/Top_Gear_Series_14_diff_LevelWords_reverse`,
	}, {
		textBefore:     `testdata/empty3lines.txt`,
		textAfter:      `testdata/empty5lines.txt`,
		diffLevelLines: `testdata/empty_lines_diff_LevelLines`,
		diffLevelWords: `testdata/empty_lines_diff_LevelWords`,
	}, {
		textBefore:     `testdata/empty5lines.txt`,
		textAfter:      `testdata/empty3lines.txt`,
		diffLevelLines: `testdata/empty_lines_diff_LevelLines_reverse`,
		diffLevelWords: `testdata/empty_lines_diff_LevelWords_reverse`,
	}, {
		textBefore:     `testdata/peeps.old`,
		textAfter:      `testdata/peeps.new`,
		diffLevelLines: `testdata/peeps_diff_LevelLines`,
		diffLevelWords: `testdata/peeps_diff_LevelWords`,
	}, {
		textBefore:     `testdata/peeps.new`,
		textAfter:      `testdata/peeps.old`,
		diffLevelLines: `testdata/peeps_diff_LevelLines_reverse`,
		diffLevelWords: `testdata/peeps_diff_LevelWords_reverse`,
	}, {
		textBefore:     `testdata/text01.old`,
		textAfter:      `testdata/text01.new`,
		diffLevelLines: `testdata/text01_diff_LevelLines`,
		diffLevelWords: `testdata/text01_diff_LevelWords`,
	}, {
		textBefore:     `testdata/text01.new`,
		textAfter:      `testdata/text01.old`,
		diffLevelLines: `testdata/text01_diff_LevelLines_reverse`,
		diffLevelWords: `testdata/text01_diff_LevelWords_reverse`,
	}, {
		textBefore:     `testdata/text02.old`,
		textAfter:      `testdata/text02.new`,
		diffLevelLines: `testdata/text02_diff_LevelLines`,
		diffLevelWords: `testdata/text02_diff_LevelWords`,
	}, {
		textBefore:     `testdata/text02.new`,
		textAfter:      `testdata/text02.old`,
		diffLevelLines: `testdata/text02_diff_LevelLines_reverse`,
		diffLevelWords: `testdata/text02_diff_LevelWords_reverse`,
	}, {
		textBefore:     `testdata/the_singles_collection.old`,
		textAfter:      `testdata/the_singles_collection.new`,
		diffLevelLines: `testdata/the_singles_collection_diff_LevelLines`,
		diffLevelWords: `testdata/the_singles_collection_diff_LevelWords`,
	}, {
		textBefore:     `testdata/the_singles_collection.new`,
		textAfter:      `testdata/the_singles_collection.old`,
		diffLevelLines: `testdata/the_singles_collection_diff_LevelLines_reverse`,
		diffLevelWords: `testdata/the_singles_collection_diff_LevelWords_reverse`,
	}}

	var (
		c      testCase
		diffs  Data
		got    string
		expStr string
		err    error
		before []byte
		after  []byte
		exp    []byte
	)

	for _, c = range cases {
		before, err = os.ReadFile(c.textBefore)
		if err != nil {
			t.Fatal(err)
		}
		after, err = os.ReadFile(c.textAfter)
		if err != nil {
			t.Fatal(err)
		}

		diffs = Text(before, after, LevelLines)
		got = diffs.String()

		exp, err = os.ReadFile(c.diffLevelLines)
		if err != nil {
			t.Fatal(err)
		}

		expStr = string(exp)
		if !reflect.DeepEqual(expStr, got) {
			t.Fatalf("%s - %s: LevelLines not matched:\n<<< want:\n%s\n<<< got:\n%s",
				c.textBefore, c.textAfter, expStr, got)
		}

		diffs = Text(before, after, LevelWords)
		got = diffs.String()

		exp, err = os.ReadFile(c.diffLevelWords)
		if err != nil {
			t.Fatal(err)
		}

		expStr = string(exp)
		if !reflect.DeepEqual(expStr, got) {
			t.Fatalf("%s - %s: LevelWords not matched:\n<<< want:\n%s\n<<< got:\n%s",
				c.textBefore, c.textAfter, expStr, got)
		}
	}
}
