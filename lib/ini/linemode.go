// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

type lineMode uint

const (
	lineModeEmpty      lineMode = 0
	lineModeComment    lineMode = 1
	lineModeSection    lineMode = 2
	lineModeSubsection lineMode = 4
	lineModeValue      lineMode = 8
	lineModeMulti      lineMode = 16
)

// isLineModeVar will return true if mode is variable, which is either
// lineModeValue or lineModeMulti; otherwise it will return
// false.
func isLineModeVar(mode lineMode) bool {
	if mode&lineModeValue > 0 {
		return true
	}
	if mode&lineModeMulti > 0 {
		return true
	}
	return false
}
