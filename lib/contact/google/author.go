// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

//
// Author define Google contacts author.
//
type Author struct {
	Name  GD `json:"name,omitempty"`
	Email GD `json:"email,omitempty"`
}
