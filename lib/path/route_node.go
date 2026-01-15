// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package path

// routeNode represent sub-path as key or as raw path.
// When a path is splitted by "/" by [Route], each splitted string will be
// stored as routeNode.
// A sub-path that start with colon ":" is a key; otherwise its normal
// sub-path.
type routeNode struct {
	name  string
	val   string
	isKey bool
}
