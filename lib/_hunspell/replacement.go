// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package hunspell

import (
	"regexp"
	"strings"
)

type replacement struct {
	from *regexp.Regexp
	to   string
}

func newReplacement(from, to string) (rep replacement, err error) {
	rep.from, err = regexp.Compile(from)
	if err != nil {
		return rep, err
	}

	rep.to = strings.ReplaceAll(to, "_", " ")

	return rep, nil
}
