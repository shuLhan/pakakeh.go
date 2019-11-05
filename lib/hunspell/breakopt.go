// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

//
// breakopt contains the break role and token from option BREAK.
//
type breakopt struct {
	delEnd   bool
	delStart bool
	token    string
}
