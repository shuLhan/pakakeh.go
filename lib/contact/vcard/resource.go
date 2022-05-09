// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vcard

// Resource define common resource located in URI or embedded in Data.
type Resource struct {
	Type string
	URI  string
	Data []byte
}
