// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package path

// routeNode represent sub-path as key or as raw path.
// When a path is splitted by "/" by [Route], each splitted string will be
// stored as routeNode.
// A sub-path that start with colon ":" is a key; otherwise its normal
// sub-path.
type routeNode struct {
	key   string
	name  string
	isKey bool
}
