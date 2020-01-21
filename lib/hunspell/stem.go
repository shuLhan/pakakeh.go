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
// stem contains the word and its attributes.
//
type stem struct {
	root        *stem
	value       string
	flags       string
	morphemes   map[string][]string
	isForbidden bool
}

func newStem(line string) (s *stem, err error) {
	if len(line) == 0 {
		return nil, nil
	}

	s = &stem{}

	err = s.parse(line)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *stem) addMorpheme(id, token string) {
	if s.morphemes == nil {
		s.morphemes = make(map[string][]string)
	}

	list := s.morphemes[id]
	list = append(list, token)
	s.morphemes[id] = list
}

func (s *stem) parse(line string) (err error) {
	var (
		id, token string
		sep       rune
		p         = parser.New(line, " \t/:")
	)

	// It's worth to add not only words, but word pairs to the dictionary
	// to get correct suggestions for common misspellings with missing
	// space.
	nword := 0
	for {
		token, sep = p.TokenEscaped('\\')
		if sep == ':' {
			break
		}
		if len(s.value) > 0 {
			s.value += " "
		}
		s.value += token
		nword++
		if nword > 2 {
			return fmt.Errorf("only one or two words allowed: %q", line)
		}
		if sep == 0 {
			return nil
		}
		if sep == '/' {
			break
		}
		_ = p.SkipHorizontalSpaces()
	}

	switch sep {
	case ' ', '\t':
		sep = p.SkipHorizontalSpaces()
		if sep == 0 {
			return nil
		}
	}

	// Each word may optionally be followed by a slash ("/")  and  one
	// or more flags, which represents the word attributes, for example
	// affixes.
	if sep == '/' {
		s.flags, sep = p.Token()
		if sep == 0 {
			return nil
		}

		sep = p.SkipHorizontalSpaces()
		if sep == 0 {
			return nil
		}
	}

	p.RemoveDelimiters("/")

	// Parse morphemes...
	for {
		if sep == ':' {
			id = token
		} else {
			id, sep = p.Token()
			if sep != ':' {
				return fmt.Errorf("invalid character in morpheme: %q", sep)
			}
			if len(id) == 0 {
				return fmt.Errorf("empty morpheme id at line %q", line)
			}
		}

		token, sep = p.Token()
		if len(token) == 0 {
			return fmt.Errorf("empty morphemes at line %q", line)
		}

		s.addMorpheme(id, token)

		sep = p.SkipHorizontalSpaces()
		if sep == 0 {
			break
		}
	}

	return nil
}

//
// unpack parse the stem and flags.
//
func (s *stem) unpack(opts *affixOptions) (derivatives []string, err error) {
	if s.value[0] == '*' {
		s.isForbidden = true
		s.value = s.value[1:]
	}

	if len(opts.afAliases) > 1 {
		afIdx, err := strconv.Atoi(s.flags)
		if err == nil {
			s.flags = opts.afAliases[afIdx]
		}
	}

	flags, err := unpackFlags(opts.flag, s.flags)
	if err != nil {
		return nil, err
	}
	if len(flags) == 0 {
		return nil, nil
	}

	for x, flag := range flags {
		pfx, ok := opts.prefixes[flag]
		if ok {
			words := pfx.apply(s.value)
			derivatives = append(derivatives, words...)
			if pfx.isCrossProduct {
				words = s.applySuffixes(opts, flags[x+1:], words)
				derivatives = append(derivatives, words...)
			}
			continue
		}
		sfx, ok := opts.suffixes[flag]
		if ok {
			words := sfx.apply(s.value)
			derivatives = append(derivatives, words...)
			continue
		}
		return nil, fmt.Errorf("unknown affix flag %q", flag)
	}

	return derivatives, nil
}

//
// applySuffixes apply any cross-product "suffixes" in "flags" for each word
// in "words".
//
func (s *stem) applySuffixes(opts *affixOptions, flags, words []string) (
	derivatives []string,
) {
	for _, word := range words {
		for _, s := range flags {
			sfx, ok := opts.suffixes[s]
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
