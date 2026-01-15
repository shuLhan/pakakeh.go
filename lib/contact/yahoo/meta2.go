// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package yahoo

// Meta2 define a common metadata inside a struct.
type Meta2 struct {
	Created string `json:"created"`
	Updated string `json:"updated"`
	URI     string `json:"uri"`
}
