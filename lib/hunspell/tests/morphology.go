// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tests

import (
	"fmt"

	"github.com/shuLhan/share/lib/hunspell"
	libstrings "github.com/shuLhan/share/lib/strings"
)

const (
	stateWord    = ">"
	stateAnalyze = "analyze"
	stateStem    = "stem"
)

type morphology struct {
	word    string
	analyze hunspell.Morphemes
	stem    string
}

func parseMorphologiesFile(morphFile string) (morphs map[string]morphology, err error) {
	lines, err := libstrings.LinesOfFile(morphFile)
	if err != nil {
		return nil, err
	}

	morphs = make(map[string]morphology)
	state := stateWord

	morph := &morphology{}

	for x := 0; x < len(lines); {
		line := lines[x]
		switch state {
		case stateWord:
			if line[0] != '>' {
				return nil, fmt.Errorf("%s line %d: expecting '>'",
					morphFile, x)
			}
			morph.word = line[2:]
			state = stateAnalyze
			x++
		case stateAnalyze:
			err = morph.parseAnalyze(line)
			if err != nil {
				return nil, fmt.Errorf("%s line %d: %w",
					morphFile, x, err)
			}
			state = stateStem
			x++
		case stateStem:
			err = morph.parseStem(line)
			if err != nil {
				state = stateWord
				continue
			}
			state = stateWord

			morphs[morph.word] = *morph
			morph = &morphology{}
			x++
		}
	}

	return morphs, nil
}

func (morph *morphology) parseAnalyze(line string) (err error) {
	p, err := morph.initParser(line, stateAnalyze)
	if err != nil {
		return fmt.Errorf("parseAnalyze: %w", err)
	}

	morph.analyze = make(hunspell.Morphemes)
	var (
		token string
		sep   rune
	)
	for {
		token, sep = p.Read()
		if sep == 0 {
			break
		}
		morph.analyze[token], _ = p.Read()
		p.SkipHorizontalSpaces()
	}
	return nil
}

func (morph *morphology) parseStem(line string) (err error) {
	p, err := morph.initParser(line, stateStem)
	if err != nil {
		return fmt.Errorf("parseStem: %w", err)
	}

	morph.stem, _ = p.Read()

	return nil
}

func (morph *morphology) initParser(line, exp string) (p *libstrings.Parser, err error) {
	p = libstrings.NewParser(line, "()=: \t")

	p.SkipHorizontalSpaces()

	token, sep := p.Read()
	if token != exp {
		return nil, fmt.Errorf("expecting %q, got %q", exp, token)
	}
	if sep != '(' {
		return nil, fmt.Errorf("expecting '(', got %q", sep)
	}

	p.SkipHorizontalSpaces()

	token, sep = p.Read()
	if sep != ')' {
		return nil, fmt.Errorf("expecting ')', got %q", sep)
	}
	if token != morph.word {
		return nil, fmt.Errorf("expecting %q, got %q", morph.word, token)
	}

	p.SkipHorizontalSpaces()

	token, sep = p.Read()
	if sep != '=' {
		return nil, fmt.Errorf("expecting '=', got %q", sep)
	}
	if len(token) != 0 {
		return nil, fmt.Errorf("unexpected token %q", token)
	}

	p.SkipHorizontalSpaces()

	return p, nil
}
