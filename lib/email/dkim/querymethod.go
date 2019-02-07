// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

//
// QueryMethod define a method to retrieve public key.
//
type QueryMethod struct {
	Type   QueryType
	Option QueryOption
}
