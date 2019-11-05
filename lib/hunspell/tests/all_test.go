// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tests

import (
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/hunspell"
	"github.com/shuLhan/share/lib/parser"
)

func TestHunspell(t *testing.T) {
	testFiles := []string{
		"affixes",
		"alias",
	}

	for _, file := range testFiles {
		affFile := filepath.Join(file + ".aff")
		dicFile := filepath.Join(file + ".dic")
		goodFile := filepath.Join(file + ".good")

		spell, err := hunspell.Open(affFile, dicFile)
		if err != nil {
			t.Fatalf("%s: %s", file, err)
		}

		exps, err := parser.Lines(goodFile)
		if err != nil {
			t.Fatalf("%s: %s", file, err)
		}

		for _, exp := range exps {
			_, ok := spell.Spell(exp)
			if !ok {
				t.Fatalf("%q not found in dictionary %q", exp, file)
			}
		}
	}
}
