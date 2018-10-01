// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package contact

//
// Name define contact's name.
//
type Name struct {
	Given       string `json:"givenName"`
	Middle      string `json:"middleName"`
	Family      string `json:"familyName"`
	Prefix      string `json:"prefix"`
	Suffix      string `json:"suffix"`
	GivenSound  string `json:"givenNameSound"`
	FamilySound string `json:"familyNameSound"`
}
