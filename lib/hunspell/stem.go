// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"fmt"
	"strconv"

	"github.com/shuLhan/share/lib/parser"
)

//
// Stem contains the word and its attributes.
//
type Stem struct {
	Word      string
	Morphemes Morphemes

	rawFlags     string
	rawMorphemes []string

	IsForbidden  bool
	IsDerivative bool // It will true if stem is derivative word.
}

//
// newStem create and initialize new stem using the root Stem, word, and
// optional list of morpheme.
//
func newStem(root *Stem, word string, morphs Morphemes) (stem *Stem) {
	stem = &Stem{
		Word:      word,
		Morphemes: make(Morphemes, len(root.Morphemes)+len(morphs)),
	}

	if root != nil {
		for k, v := range root.Morphemes {
			stem.Morphemes.set(k, v)
		}
		stem.IsDerivative = true
		stem.Morphemes.set("st", root.Word)
	} else {
		stem.Morphemes.set("st", word)
	}

	if len(morphs) > 0 {
		for k, v := range morphs {
			stem.Morphemes.set(k, v)
		}
	}

	return stem
}

func parseStem(line string) (stem *Stem, err error) {
	if len(line) == 0 {
		return nil, nil
	}

	stem = &Stem{
		Morphemes: make(Morphemes),
	}

	err = stem.parse(line)
	if err != nil {
		return nil, err
	}

	return stem, nil
}

//
// parse the single line of word with optional flags and zero or more
// morphemes attributes.
//
//	STEM := WORD [ " " WORD ] [ "/" FLAGS ] [ *MORPHEME ]
//
func (stem *Stem) parse(line string) (err error) {
	var (
		token  string
		sep    rune
		nwords int
		p      = parser.New(line, " \t")
	)

	// Parse one or two words with optional flags, and possibly one
	// morpheme.
	for {
		token, sep = p.Token()
		if len(token) == 0 {
			return nil
		}
		ok, err := isValidMorpheme(token)
		if err != nil {
			return err
		}
		if ok {
			stem.rawMorphemes = append(stem.rawMorphemes, token)
			p.SkipHorizontalSpaces()
			break
		}

		token, stem.rawFlags, err = parseWordFlags(token)
		if err != nil {
			return err
		}

		if len(stem.Word) > 0 {
			stem.Word += " "
		}
		stem.Word += token
		nwords++
		if nwords > 2 {
			return fmt.Errorf("only one or two words allowed: %q", line)
		}

		p.SkipHorizontalSpaces()

		if len(stem.rawFlags) > 0 {
			break
		}
		if sep == 0 {
			// Its words without a flags.
			return nil
		}
	}
	// Parse the rest of morphemes.
	for {
		token, sep = p.Token()
		if len(token) == 0 {
			break
		}
		ok, err := isValidMorpheme(token)
		if err != nil {
			return err
		}
		if !ok {
			return errInvalidMorpheme(token)
		}
		stem.rawMorphemes = append(stem.rawMorphemes, token)
		p.SkipHorizontalSpaces()
	}

	return nil
}

//
// unpack parse the stem and flags.
//
func (stem *Stem) unpack(opts *affixOptions) (derivatives []*Stem, err error) {
	if stem.Word[0] == '*' {
		stem.IsForbidden = true
		stem.Word = stem.Word[1:]
	}

	stem.Morphemes = newMorphemes(opts, stem.rawMorphemes)
	stem.Morphemes.set("st", stem.Word)

	derivatives, err = stem.unpackFlags(opts)
	if err != nil {
		return derivatives, err
	}

	return derivatives, nil
}

func (stem *Stem) unpackFlags(opts *affixOptions) (
	derivatives []*Stem, err error,
) {
	if len(opts.afAliases) > 1 {
		afIdx, err := strconv.Atoi(stem.rawFlags)
		if err == nil {
			stem.rawFlags = opts.afAliases[afIdx]
		}
	}

	flags, err := unpackFlags(opts.flag, stem.rawFlags)
	if err != nil {
		return nil, err
	}
	if len(flags) == 0 {
		return nil, nil
	}

	for _, flag := range flags {
		pfx, ok := opts.prefixes[flag]
		if ok {
			stems := pfx.apply(stem)
			derivatives = append(derivatives, stems...)
			if pfx.isCrossProduct {
				stems = stem.applySuffixes(opts, flags, stems)
				derivatives = append(derivatives, stems...)
			}
			continue
		}
		sfx, ok := opts.suffixes[flag]
		if ok {
			stems := sfx.apply(stem)
			derivatives = append(derivatives, stems...)
			continue
		}
		return nil, fmt.Errorf("unknown affix flag %q", flag)
	}

	return derivatives, nil
}

//
// applySuffixes apply any cross-product "suffixes" in "flags" for each word
// in "stems".
//
func (stem *Stem) applySuffixes(
	opts *affixOptions, flags []string, stems []*Stem,
) (
	derivatives []*Stem,
) {
	for _, substem := range stems {
		for _, flag := range flags {
			sfx, ok := opts.suffixes[flag]
			if !ok {
				continue
			}
			if !sfx.isCrossProduct {
				continue
			}
			ss := sfx.apply(substem)
			derivatives = append(derivatives, ss...)
		}
	}
	return derivatives
}

func parseWordFlags(in string) (word, flags string, err error) {
	var (
		end int = -1
		esc bool
		v   = make([]rune, 0, len(in))
	)
	for x, c := range in {
		if esc {
			if c != '/' {
				return "", "", fmt.Errorf("invalid escape %q", in)
			}
			esc = false
			v = append(v, c)
			continue
		}
		if c == '\\' {
			esc = true
			continue
		}
		if c == '/' {
			end = x
			break
		}
		v = append(v, c)
	}
	if esc {
		return "", "", fmt.Errorf("invalid escape %q", in)
	}
	if end == 0 {
		return "", "", fmt.Errorf("invalid word format %q", in)
	}
	if end == -1 {
		// No flags found.
		return string(v), "", nil
	}
	flags = in[end+1:]
	return string(v), flags, nil
}
