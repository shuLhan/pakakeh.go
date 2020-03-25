// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tests

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/hunspell"
	"github.com/shuLhan/share/lib/parser"
)

func TestHunspell(t *testing.T) {
	testFiles := []string{
		"affixes",
		"alias",
		"alias2",
	}

	for _, file := range testFiles {
		t.Logf("test file: %s", file)
		affFile := filepath.Join(file + ".aff")
		dicFile := filepath.Join(file + ".dic")
		goodFile := filepath.Join(file + ".good")
		morphFile := filepath.Join(file + ".morph")

		spell, err := hunspell.Open(affFile, dicFile)
		if err != nil {
			t.Fatalf("%s: %s", affFile, err)
		}

		exps, err := parser.Lines(goodFile)
		if err != nil {
			t.Fatalf("%s: %s", goodFile, err)
		}

		for _, exp := range exps {
			gotStem := spell.Spell(exp)
			if gotStem == nil {
				t.Fatalf("%q not found in dictionary %q", exp, file)
			}
			t.Logf("got stem(%s): %+v", exp, gotStem)
		}

		expMorphs, err := parseMorphologiesFile(morphFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			t.Fatalf("%s: %s", morphFile, err)
		}

		for _, word := range exps {
			_ = expMorphs[word]
		}
	}
}
