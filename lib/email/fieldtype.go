// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

type FieldType int

const (
	FieldTypeOptional FieldType = 0
	FieldTypeDate     FieldType = 1 << iota
)
