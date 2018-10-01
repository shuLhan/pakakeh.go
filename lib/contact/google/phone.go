// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

// Phone format.
type Phone struct {
	Rel    string `json:"rel,omitempty"`
	Number string `json:"$t,omitempty"`
}
