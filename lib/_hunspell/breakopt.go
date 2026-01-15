// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package hunspell

// breakopt contains the break role and token from option BREAK.
type breakopt struct {
	token string

	delEnd   bool
	delStart bool
}
