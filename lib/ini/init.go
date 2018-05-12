// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"os"
	"strconv"
)

type debugMode uint

const (
	debugL0 debugMode = 0
	debugL1 debugMode = 1
	debugL2 debugMode = 2
)

const (
	envDEBUG = "INI_DEBUG"
)

var debug = debugL0

func init() {
	d := os.Getenv(envDEBUG)

	if len(d) == 0 {
		return
	}

	v, err := strconv.Atoi(d)
	if err != nil {
		return
	}

	debug = debugMode(v)
}
