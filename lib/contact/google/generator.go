// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

// Generator define Google contact generator.
type Generator struct {
	Version string `json:"version,omitempty"`
	URI     string `json:"uri,omitempty"`
	Value   string `json:"$t,omitempty"`
}
