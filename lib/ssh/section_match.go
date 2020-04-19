// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"errors"
	"fmt"
	"strings"

	"github.com/shuLhan/share/lib/parser"
)

var (
	errCriteriaAll = errors.New(`the "all" criteria must appear alone` +
		` or immediately after "canonical" or "final`)
)

//
// newSectionMatch create new Match section using one or more criteria or the
// single token "all" which always matches.
//
// The available criteria keywords are: canonical, final, exec, host,
// originalhost, user, and localuser.
// Other criteria may be combined arbitrarily.
// All criteria but "all", "canonical", and "final" require an argument.
// Criteria may be negated by prepending an exclamation mark (`!').
//
func newSectionMatch(rawPattern string) (match *ConfigSection, err error) {
	var (
		prevCriteria *matchCriteria
		criteria     *matchCriteria
	)

	match = newConfigSection()
	match.criterias = make([]*matchCriteria, 0)
	match.useCriterias = true

	p := parser.New(rawPattern, ` "`)

	for {
		var (
			err      error
			isNegate bool
		)

		token, _ := p.Token()
		if len(token) == 0 {
			break
		}

		token = strings.ToLower(token)

		if token[0] == '!' {
			isNegate = true
			token = token[1:]
		}

		switch token {
		case criteriaAll:
			criteria, err = parseCriteriaAll(p, prevCriteria)

		case criteriaCanonical, criteriaFinal:
			criteria, err = newMatchCriteria(token, "")

		case criteriaExec, criteriaHost, criteriaLocalUser, criteriaOriginalHost,
			criteriaUser:
			criteria, err = parseCriteriaWithArg(p, token)
		default:
			return nil, fmt.Errorf("unknown criteria %q", token)
		}
		if err != nil {
			return nil, err
		}

		criteria.isNegate = isNegate

		match.criterias = append(match.criterias, criteria)
		prevCriteria = criteria
		criteria = nil
	}

	return match, nil
}

func parseCriteriaAll(p *parser.Parser, prevCriteria *matchCriteria) (
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
	token, sep := p.Token()
	if len(token) > 0 || sep != 0 {
		return nil, errCriteriaAll
	}

	return newMatchCriteria(criteriaAll, "")
}

func parseCriteriaWithArg(p *parser.Parser, name string) (
	criteria *matchCriteria, err error,
) {
	arg, sep := p.Token()
	if sep == '"' {
		p.RemoveDelimiters(` `)
		arg, sep = p.Token()
		if sep != '"' {
			return nil, fmt.Errorf(`%q: expecting '"' got %q`,
				name, sep)
		}
		p.AddDelimiters(`"`)
	}

	return newMatchCriteria(name, arg)
}
