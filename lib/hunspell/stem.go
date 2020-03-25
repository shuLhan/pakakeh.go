// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/shuLhan/share/lib/parser"
)

//
// Stem contains the word and its attributes.
//
type Stem struct {
	Word      string
	Morphemes map[string][]string

	rawFlags     string
	rawMorphemes []string

	IsForbidden bool
}

func newStem(line string) (stem *Stem, err error) {
	if len(line) == 0 {
		return nil, nil
	}

	stem = &Stem{}

	err = stem.parse(line)
	if err != nil {
		return nil, err
	}

	return stem, nil
}

func (stem *Stem) addMorpheme(id, token string) {
	if stem.Morphemes == nil {
		stem.Morphemes = make(map[string][]string)
	}

	list := stem.Morphemes[id]
	list = append(list, token)
	stem.Morphemes[id] = list
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
func (stem *Stem) unpack(opts *affixOptions) (derivatives []string, err error) {
	if stem.Word[0] == '*' {
		stem.IsForbidden = true
		stem.Word = stem.Word[1:]
	}

	derivatives, err = stem.unpackFlags(opts)
	if err != nil {
		return derivatives, err
	}

	stem.unpackMorphemes(opts)

	return derivatives, nil
}

func (stem *Stem) unpackFlags(opts *affixOptions) (derivatives []string, err error) {
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
			words := pfx.apply(stem.Word)
			derivatives = append(derivatives, words...)
			if pfx.isCrossProduct {
				words = stem.applySuffixes(opts, flags, words)
				derivatives = append(derivatives, words...)
			}
			continue
		}
		sfx, ok := opts.suffixes[flag]
		if ok {
			words := sfx.apply(stem.Word)
			derivatives = append(derivatives, words...)
			continue
		}
		return nil, fmt.Errorf("unknown affix flag %q", flag)
	}

	return derivatives, nil
}

//
// unpackMorphemes convert any raw morphemes or an alias into map of
// key-values.
// At this point, each of the morphemes should be valid, unless its unknown
// and it will logged to stderr.
//
func (stem *Stem) unpackMorphemes(opts *affixOptions) {
	for _, m := range stem.rawMorphemes {
		idx := strings.Index(m, ":")

		if idx == -1 {
			if len(opts.amAliases) > 0 {
				// Convert the AM alias number to actual
				// morpheme.
				amIdx, err := strconv.Atoi(m)
				if err != nil {
					log.Printf("unknown morpheme %q", m)
					continue
				}
				m = opts.amAliases[amIdx]
				idx = strings.Index(m, ":")
			}
		}
		stem.addMorpheme(m[:idx], m[idx+1:])
	}
}

//
// applySuffixes apply any cross-product "suffixes" in "flags" for each word
// in "words".
//
func (stem *Stem) applySuffixes(opts *affixOptions, flags, words []string) (
	derivatives []string,
) {
	for _, word := range words {
		for _, flag := range flags {
			sfx, ok := opts.suffixes[flag]
			if !ok {
				continue
			}
			if !sfx.isCrossProduct {
				continue
			}
			ss := sfx.apply(word)
			derivatives = append(derivatives, ss...)
		}
	}
	return derivatives
}

//
// isValidMorpheme will return true if `in` contains ":" or a number (as an
// alias); otherwise it will return false.
//
func isValidMorpheme(in string) (bool, error) {
	idx := strings.Index(in, ":")
	switch idx {
	case -1:
		_, err := strconv.Atoi(in)
		if err == nil {
			return true, nil
		}
		return false, nil
	case 0:
		return false, errInvalidMorpheme(in)
	}

	return true, nil
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
