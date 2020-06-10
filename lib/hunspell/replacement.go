// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
