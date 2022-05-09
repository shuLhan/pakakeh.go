// Copyright 2017, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package tabula

// Matrix is a combination of columns and rows.
type Matrix struct {
	Columns *Columns
	Rows    *Rows
}
