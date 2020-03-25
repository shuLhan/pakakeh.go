// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"log"
	"sort"
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
	rawMorphs := make([]string, 0, len(raws))
	morphs := make(Morphemes, len(raws))

	// Normalize the raws strings.
	for _, raw := range raws {
		for _, m := range strings.Fields(raw) {
			idx := strings.Index(m, ":")
			switch idx {
			case -1:
				if len(opts.amAliases) == 0 {
					continue
				}

				// Convert the AM alias number to actual
				// morpheme.
				amIdx, err := strconv.Atoi(m)
				if err != nil {
					log.Printf("unknown morpheme %q", m)
					continue
				}
				m = opts.amAliases[amIdx]
				if len(m) > 0 {
					rawMorphs = append(rawMorphs, strings.Fields(m)...)
				}
			case 0:
				continue
			default:
				rawMorphs = append(rawMorphs, m)
			}
		}
	}

	for _, raw := range rawMorphs {
		morphs.add(raw)
	}

	return morphs
}

//
// String return list of morphological fields ordered by key.
//
func (morphs Morphemes) String() string {
	fields := make([]string, 0, len(morphs))
	for k, v := range morphs {
		fields = append(fields, k+":"+v)
	}
	sort.Strings(fields)
	return strings.Join(fields, " ")
}

func (morphs Morphemes) add(raw string) {
	idx := strings.Index(raw, ":")
	switch idx {
	case -1:
		morphs[raw] = ""
	case 0:
		morphs[""] = raw[1:]
	default:
		morphs.set(raw[:idx], raw[idx+1:])
	}
}

func (morphs Morphemes) set(id, attr string) {
	morphs[id] = attr
}

func (morphs Morphemes) clone() Morphemes {
	clone := make(Morphemes, len(morphs))
	for k, v := range morphs {
		clone[k] = v
	}
	return clone
}
