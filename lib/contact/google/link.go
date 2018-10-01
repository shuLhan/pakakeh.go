// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

//
// Link define Google contact link type.
//
type Link struct {
	Rel  string `json:"rel,omitempty"`
	Type string `json:"type,omitempty"`
	HRef string `json:"href,omitempty"`
}
