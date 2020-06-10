// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package reflect

//
// Equaler is an interface that when implemented by a type, it will be used to
// compare the value in Assert.
//
type Equaler interface {
	IsEqual(v interface{}) bool
}
