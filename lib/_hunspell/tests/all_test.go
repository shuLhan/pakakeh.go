// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

// Packages tests contains test for hunspell package.
package tests

import (
	"errors"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/hunspell"
	libstrings "git.sr.ht/~shulhan/pakakeh.go/lib/strings"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestHunspell(t *testing.T) {
	testFiles := []string{
		"affixes",
		"alias",
		"alias2",
		"alias3",
		"allcaps",
	}

	for _, file := range testFiles {
		t.Logf("test file: %s", file)

		var (
			affFile   = file + `.aff`
			dicFile   = file + `.dic`
			goodFile  = file + `.good`
			morphFile = file + `.morph`
		)

		spell, err := hunspell.Open(affFile, dicFile)
		if err != nil {
			t.Fatalf("%s: %s", affFile, err)
		}

		exps, err := libstrings.LinesOfFile(goodFile)
		if err != nil {
			t.Fatalf("%s: %s", goodFile, err)
		}

		for _, exp := range exps {
			gotStem := spell.Spell(exp)
			if gotStem == nil {
				t.Fatalf("%q not found in dictionary %q", exp, file)
			}
		}

		expMorphs, err := parseMorphologiesFile(morphFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			t.Fatalf("%s: %s", morphFile, err)
		}

		for _, morph := range expMorphs {
			gotAnalyze := spell.Analyze(morph.word)

			test.Assert(t, "Analyze("+morph.word+")", morph.analyze.String(), gotAnalyze.String())

			gotStem := spell.Stem(morph.word)

			test.Assert(t, "Stem("+morph.word+")", morph.stem, gotStem.Word)
		}
	}
}
