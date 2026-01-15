// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package hunspell

// compoundPattern define the option for COMPUNDPATTERN.
//
// Forbid compounding, if the first word in the compound ends with endchars,
// and next word begins with beginchars and (optionally) they have the
// requested flags.
// The optional replacement parameter allows simplified compound form.
// The special "endchars" pattern 0 (zero) limits the rule to the  unmodified
// stems (stems and stems with zero affixes):
//
//	CHECKCOMPOUNDPATTERN 0/x /y
type compoundPattern struct {
	end       string
	endFlag   string
	begin     string
	beginFlag string
	rep       string
}
