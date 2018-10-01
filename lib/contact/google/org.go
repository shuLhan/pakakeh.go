// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

// Org as organisation.
type Org struct {
	Type     string `json:"rel,omitempty"`
	Name     GD     `json:"gd$orgName,omitempty"`
	JobTitle GD     `json:"gd$orgTitle,omitempty"`
}
