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
	lineModeKeyOnly    lineMode = 8
	lineModeKeyValue   lineMode = 16
)

// isLineModeVar true if mode is variable, which is either lineModeKeyOnly or
// lineModeKeyValue;
// otherwise it will return false.
func isLineModeVar(mode lineMode) bool {
	return mode >= lineModeKeyOnly
}
