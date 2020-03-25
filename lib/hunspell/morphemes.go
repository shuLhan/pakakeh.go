// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"log"
	"strconv"
	"strings"
)

//
// Morphemes contains list of morphological attributes.
//
type Morphemes map[string]string

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

//
// newMorphemes convert any raw morphemes or an alias into map of
// key-values.
// At this point, each of the morphemes should be valid, unless its unknown
// and it will logged to stderr.
//
func newMorphemes(opts *affixOptions, raws []string) Morphemes {
	morphs := make(Morphemes, len(raws))
	for _, raw := range raws {
		for _, m := range strings.Fields(raw) {
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
					if idx <= 0 {
						continue
					}
				}
			} else if idx == 0 {
				continue
			}
			morphs.set(m[:idx], m[idx+1:])
		}
	}
	return morphs
}

func (morphs Morphemes) set(id, attr string) {
	morphs[id] = attr
}
