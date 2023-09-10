// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

// JSONFooter define the optional metadata and data at the footer of the
// token that are not included in signature.
type JSONFooter struct {
	Data map[string]interface{} `json:"data,omitempty"`
	KID  string                 `json:"kid"`
}
