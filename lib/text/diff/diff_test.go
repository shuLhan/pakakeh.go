// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Mhd Sulhan <ms@kilabit.info>

package diff

import (
	"os"
	"strings"
	"testing"

	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
	"git.sr.ht/~shulhan/pakakeh.go/lib/text"
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
		oldline = text.Line{N: 0, V: []byte(`lorem ipsum dolmet`)}
		newline = text.Line{N: 0, V: []byte(`lorem all ipsum`)}

		expAdds = libstrings.Row{
			[]string{`all `},
		}
		expDels = libstrings.Row{
			[]string{` dolmet`},
		}
	)

	testDiffBytes(t, oldline, newline, expAdds[0], expDels[0])

	oldline = text.Line{N: 0, V: []byte(`lorem ipsum dolmet`)}
	newline = text.Line{N: 0, V: []byte(`lorem ipsum`)}

	testDiffBytes(t, oldline, newline, []string{}, expDels[0])

	oldline = text.Line{N: 0, V: []byte(`lorem ipsum`)}
	newline = text.Line{N: 0, V: []byte(`lorem ipsum dolmet`)}

	testDiffBytes(t, oldline, newline, expDels[0], []string{})

	oldline = text.Line{N: 0, V: []byte(`{{Pharaoh Infobox |`)}
	newline = text.Line{N: 0, V: []byte(`{{Infobox pharaoh`)}

	testDiffBytes(t, oldline, newline, []string{`pharaoh`}, []string{`Pharaoh `, `|`})
}

func TestBytesRatio(t *testing.T) {
	var (
		oldline  = `# [[...Baby One More Time (song)|...Baby One More Time]]`
		newline  = `# "[[...Baby One More Time (song)|...Baby One More Time]]"`
		newlen   = len(newline)
		expMatch = newlen - 2
		expRatio = float32(expMatch) / float32(newlen)

		ratio float32
	)

	ratio, _, _ = BytesRatio([]byte(oldline), []byte(newline), DefMatchLen)

	if expRatio != ratio {
		t.Fatalf(`Expecting ratio %f got %f`, expRatio, ratio)
	}
}

func TestFiles_LevelWords(t *testing.T) {
	listCase := []struct {
		desc    string
		oldFile string
		newFile string
		exp     string
	}{{
		desc:    `List_of_United_Nations`,
		oldFile: `testdata/List_of_United_Nations.old`,
		newFile: `testdata/List_of_United_Nations.new`,
		// No changes on chunk level.
	}, {
		desc:    `List_of_United_Nations.reverse`,
		oldFile: `testdata/List_of_United_Nations.new`,
		newFile: `testdata/List_of_United_Nations.old`,
		// No changes on chunk level.
	}, {
		desc:    `Psusennes_II`,
		oldFile: `testdata/Psusennes_II.old`,
		newFile: `testdata/Psusennes_II.new`,
		exp:     `testdata/Psusennes_II.chunks`,
	}, {
		desc:    `Psusennes_II.reverse`,
		oldFile: `testdata/Psusennes_II.new`,
		newFile: `testdata/Psusennes_II.old`,
		exp:     `testdata/Psusennes_II.reverse.chunks`,
	}, {
		desc:    `Top_Gear_Series_14`,
		oldFile: `testdata/Top_Gear_Series_14.old`,
		newFile: `testdata/Top_Gear_Series_14.new`,
		exp:     `testdata/Top_Gear_Series_14.chunks`,
	}, {
		desc:    `Top_Gear_Series_14.reverse`,
		oldFile: `testdata/Top_Gear_Series_14.new`,
		newFile: `testdata/Top_Gear_Series_14.old`,
		exp:     `testdata/Top_Gear_Series_14.reverse.chunks`,
	}, {
		desc:    `empty_lines`,
		oldFile: `testdata/empty3lines.txt`,
		newFile: `testdata/empty5lines.txt`,
		exp:     `testdata/empty_lines.chunks`,
	}, {
		desc:    `empty_lines.reverse`,
		oldFile: `testdata/empty5lines.txt`,
		newFile: `testdata/empty3lines.txt`,
		exp:     `testdata/empty_lines.reverse.chunks`,
	}, {
		desc:    `peeps`,
		oldFile: `testdata/peeps.old`,
		newFile: `testdata/peeps.new`,
		exp:     `testdata/peeps.chunks`,
	}, {
		desc:    `peeps.reverse`,
		oldFile: `testdata/peeps.new`,
		newFile: `testdata/peeps.old`,
		exp:     `testdata/peeps.reverse.chunks`,
	}, {
		desc:    `text01`,
		oldFile: `testdata/text01.old`,
		newFile: `testdata/text01.new`,
		exp:     `testdata/text01.chunks`,
	}, {
		desc:    `text01.reverse`,
		oldFile: `testdata/text01.new`,
		newFile: `testdata/text01.old`,
		exp:     `testdata/text01.reverse.chunks`,
	}, {
		desc:    `text02`,
		oldFile: `testdata/text02.old`,
		newFile: `testdata/text02.new`,
		exp:     `testdata/text02.chunks`,
	}, {
		desc:    `text02.reverse`,
		oldFile: `testdata/text02.new`,
		newFile: `testdata/text02.old`,
		exp:     `testdata/text02.reverse.chunks`,
	}, {
		desc:    `the_singles_collection`,
		oldFile: `testdata/the_singles_collection.old`,
		newFile: `testdata/the_singles_collection.new`,
		exp:     `testdata/the_singles_collection.chunks`,
	}, {
		desc:    `the_singles_collection.reverse`,
		oldFile: `testdata/the_singles_collection.new`,
		newFile: `testdata/the_singles_collection.old`,
		exp:     `testdata/the_singles_collection.reverse.chunks`,
	}}

	var sb strings.Builder
	for _, tc := range listCase {
		diff, err := Files(tc.oldFile, tc.newFile, LevelWords)
		if err != nil {
			t.Fatalf(`%s: %s`, tc.desc, err)
		}
		sb.Reset()
		err = diff.WriteLineChunks(&sb)
		if err != nil {
			t.Fatalf(`%s: %s`, tc.desc, err)
		}
		got := sb.String()

		var exp string
		if tc.exp != `` {
			expb, err := os.ReadFile(tc.exp)
			if err != nil {
				t.Fatalf(`%s: %s`, tc.desc, err)
			}
			exp = string(expb)
		}

		if exp != got {
			os.WriteFile(tc.exp, []byte(got), 0600)
			t.Errorf("<<< want:\n%s\n", exp)
			t.Errorf("<<< got:\n%s\n", got)
			t.Fatalf(`%s: not matched`, tc.desc)
		}
	}
}
