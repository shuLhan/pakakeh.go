// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yahoo

// Contacts define the holder for root of contacts response.
type Contacts struct {
	Contact []Contact `json:"contact"`
	Start   int       `json:"start"`
	Count   int       `json:"count"`
	Total   int       `json:"total"`
	URI     string    `json:"uri"`
}
