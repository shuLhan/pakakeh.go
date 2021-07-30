// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

//
// node represent sub-path as key or as raw path.
// The original path is splitted by "/" and each splitted string will be
// stored as node.  A sub-path that start with colon ":" is a key; otherwise
// its normal sub-path.
//
type node struct {
	key   string
	name  string
	isKey bool
}
