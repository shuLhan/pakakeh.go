// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

// Email format.
type Email struct {
	Rel     string `json:"rel,omitempty"`
	Address string `json:"address,omitempty"`
	Primary string `json:"primary,omitempty"`
}
