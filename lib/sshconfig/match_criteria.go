// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sshconfig

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

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
	patterns []*pattern
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
	criteria.patterns = make([]*pattern, 0, len(listPattern))

	for _, raw := range listPattern {
		pat := newPattern(raw)
		criteria.patterns = append(criteria.patterns, pat)
	}

	return criteria, nil
}

// MarshalText encode the criteria back to ssh_config format.
func (mcriteria *matchCriteria) MarshalText() (text []byte, err error) {
	var logp = `MarshalText`
	var buf bytes.Buffer

	if mcriteria.isNegate {
		buf.WriteByte('!')
	}
	buf.WriteString(mcriteria.name)

	var (
		pat *pattern
		x   int
	)
	for x, pat = range mcriteria.patterns {
		if x == 0 {
			buf.WriteByte(' ')
		} else {
			buf.WriteByte(',')
		}
		_, err = pat.WriteTo(&buf)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	return buf.Bytes(), nil
}

// WriteTo marshal the matchCriteria into text and write it to w.
func (mcriteria *matchCriteria) WriteTo(w io.Writer) (n int64, err error) {
	var text []byte
	text, _ = mcriteria.MarshalText()

	var c int
	c, err = w.Write(text)
	return int64(c), err
}

func (mcriteria *matchCriteria) isMatch(s string) bool {
	switch mcriteria.name {
	case criteriaAll:
		if mcriteria.isNegate {
			return false
		}
		return true
	case criteriaCanonical:
		// TODO
	case criteriaExec:
		// TODO
	case criteriaFinal:
		// TODO
	case criteriaHost, criteriaLocalUser, criteriaOriginalHost, criteriaUser:
		for _, pat := range mcriteria.patterns {
			if pat.isMatch(s) {
				return !mcriteria.isNegate
			}
		}
	}
	return false
}
