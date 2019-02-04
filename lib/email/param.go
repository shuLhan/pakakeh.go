// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

//
// Param represent a key-value in slice of bytes.
//
type Param struct {
	Key    []byte
	Value  []byte
	Quoted bool // Quoted is true if value is contains special characters.
}
