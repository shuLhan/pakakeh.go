// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

// breakopt contains the break role and token from option BREAK.
type breakopt struct {
	token string

	delEnd   bool
	delStart bool
}
