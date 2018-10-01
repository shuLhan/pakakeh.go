// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

// Category format.
type Category struct {
	Scheme string `json:"scheme,omitempty"`
	Term   string `json:"term,omitempty"`
}
