// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package microsoft

// Root of response.
type Root struct {
	Context  string    `json:"@odata.context"`
	Contacts []Contact `json:"value"`
}
