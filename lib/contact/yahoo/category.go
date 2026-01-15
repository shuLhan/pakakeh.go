// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package yahoo

// Category define a contact category.
type Category struct {
	Meta2
	Name string `json:"name"`
	ID   int    `json:"id"`
}
