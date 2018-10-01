// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

//
// Root define the root of Google's contact in JSON.
//
type Root struct {
	Version  string `json:"version,omitempty"`
	Encoding string `json:"encoding,omitempty"`
	Feed     Feed   `json:"feed,omitempty"`
}
