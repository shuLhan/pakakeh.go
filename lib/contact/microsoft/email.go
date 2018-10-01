// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package microsoft

//
// Email format on response.
//
type Email struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}
