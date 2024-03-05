// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sshconfig

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errCriteriaAll = errors.New(`the "all" criteria must appear alone or immediately after "canonical" or "final`)
)

// newSectionMatch create new Match section using one or more criteria or the
// single token "all" which always matches.
//
// The available criteria keywords are: canonical, final, exec, host,
// originalhost, user, and localuser.
// Other criteria may be combined arbitrarily.
// All criteria but "all", "canonical", and "final" require an argument.
// Criteria may be negated by prepending an exclamation mark (`!').
func newSectionMatch(cfg *Config, rawPattern string) (match *Section, err error) {
	var (
		prevCriteria *matchCriteria
		criteria     *matchCriteria
	)

	match = NewSection(cfg, rawPattern)
	match.useCriteria = true

	args := parseArgs(rawPattern, ' ')

	var (
		arg      string
		isNegate bool
	)

	for x := 0; x < len(args); x++ {
		token := strings.ToLower(args[x])
		if x+1 < len(args) {
			arg = args[x+1]
		} else {
			arg = ""
		}

		if token[0] == '!' {
			isNegate = true
			token = token[1:]
		} else {
			isNegate = false
		}

		switch token {
		case criteriaAll:
			criteria, err = parseCriteriaAll(prevCriteria, arg)

		case criteriaCanonical, criteriaFinal:
			criteria, err = newMatchCriteria(token, "")

		case criteriaExec, criteriaHost, criteriaLocalUser, criteriaOriginalHost,
			criteriaUser:
			criteria, err = newMatchCriteria(token, arg)
			x++
		default:
			err = fmt.Errorf(`unknown criteria %q`, token)
		}
		if err != nil {
			return nil, err
		}

		criteria.isNegate = isNegate

		match.criteria = append(match.criteria, criteria)
		prevCriteria = criteria
		criteria = nil
	}

	return match, nil
}

func parseCriteriaAll(prevCriteria *matchCriteria, arg string) (
	criteria *matchCriteria, err error,
) {
	// The "all" criteria must appear alone or immediately
	// after "canonical" or "final".
	if prevCriteria != nil {
		if !(prevCriteria.name == criteriaCanonical ||
			prevCriteria.name == criteriaFinal) {
			return nil, errCriteriaAll
		}
	}
	if len(arg) > 0 {
		return nil, errCriteriaAll
	}

	return newMatchCriteria(criteriaAll, "")
}
