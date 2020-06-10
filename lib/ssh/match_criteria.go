// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import "strings"

const (
	criteriaAll          = "all"
	criteriaCanonical    = "canonical"
	criteriaExec         = "exec"
	criteriaFinal        = "final"
	criteriaHost         = "host"
	criteriaLocalUser    = "localuser"
	criteriaOriginalHost = "originalhost"
	criteriaUser         = "user"
)

type matchCriteria struct {
	name     string
	arg      string
	patterns []*configPattern
	isNegate bool
}

func newMatchCriteria(name, arg string) (criteria *matchCriteria, err error) {
	criteria = &matchCriteria{
		name: name,
		arg:  arg,
	}
	if len(arg) == 0 {
		return criteria, nil
	}
	if name == criteriaExec {
		return criteria, nil
	}

	listPattern := strings.Split(arg, ",")
	criteria.patterns = make([]*configPattern, 0, len(listPattern))

	for _, raw := range listPattern {
		pattern := newConfigPattern(raw)
		criteria.patterns = append(criteria.patterns, pattern)
	}

	return criteria, nil
}

func (mcriteria *matchCriteria) isMatch(s string) bool {
	switch mcriteria.name {
	case criteriaAll:
		if mcriteria.isNegate {
			return false
		}
		return true
	case criteriaCanonical:
		//TODO
	case criteriaExec:
		//TODO
	case criteriaFinal:
		//TODO
	case criteriaHost, criteriaLocalUser, criteriaOriginalHost, criteriaUser:
		for _, pattern := range mcriteria.patterns {
			if pattern.isMatch(s) {
				return !mcriteria.isNegate
			}
		}
	}
	return false
}
