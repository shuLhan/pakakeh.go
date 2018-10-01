// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

//
// Name define Google contact name format.
//
type Name struct {
	Prefix GD `json:"gd$namePrefix,omitempty"`
	First  GD `json:"gd$givenName,omitempty"`
	Middle GD `json:"gd$additionalName,omitempty"`
	Last   GD `json:"gd$familyName,omitempty"`
	Suffix GD `json:"gd$nameSuffix,omitempty"`
	Full   GD `json:"gd$fullName,omitempty"`
}
