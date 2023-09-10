// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yahoo

// Category define a contact category.
type Category struct {
	Meta2
	Name string `json:"name"`
	ID   int    `json:"id"`
}
