// Copyright 2018 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package diff

import (
	"testing"

	libstrings "github.com/shuLhan/share/lib/strings"
	"github.com/shuLhan/share/lib/test"
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
	var (
		dataFiles = []string{
			`testdata/List_of_United_Nations_test.txt`,
			`testdata/Psusennes_II_test.txt`,
			`testdata/Top_Gear_Series_14_test.txt`,
			`testdata/empty_lines_test.txt`,
			`testdata/peeps_test.txt`,
			`testdata/text01_test.txt`,
			`testdata/text02_test.txt`,
			`testdata/the_singles_collection_test.txt`,
		}

		tdata  *test.Data
		diffs  Data
		dfile  string
		exp    string
		got    string
		err    error
		before []byte
		after  []byte
	)

	for _, dfile = range dataFiles {
		t.Run(dfile, func(t *testing.T) {
			tdata, err = test.LoadData(dfile)
			if err != nil {
				t.Fatal(err)
			}

			before = tdata.Input[`before`]
			after = tdata.Input[`after`]

			// Diff with LevelLines.

			exp = string(tdata.Output[`diffs_LevelLines`])
			diffs = Text(before, after, LevelLines)
			got = diffs.String()
			test.Assert(t, `Text, LevelLines`, exp, string(got))

			// Reverse the inputs.
			exp = string(tdata.Output[`diffs_LevelLines_reverse`])
			diffs = Text(after, before, LevelLines)
			got = diffs.String()
			test.Assert(t, `Text, LevelLines, reverse`, exp, string(got))

			// Diff with LevelWords.

			exp = string(tdata.Output[`diffs_LevelWords`])
			diffs = Text(before, after, LevelWords)
			got = diffs.String()
			test.Assert(t, `Text, LevelWords`, exp, string(got))

			// Reverse the inputs.
			exp = string(tdata.Output[`diffs_LevelWords_reverse`])
			diffs = Text(after, before, LevelWords)
			got = diffs.String()
			test.Assert(t, `Text, LevelWords, reverse`, exp, string(got))
		})
	}
}
