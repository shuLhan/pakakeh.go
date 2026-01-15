// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package yahoo

// Contacts define the holder for root of contacts response.
type Contacts struct {
	URI     string    `json:"uri"`
	Contact []Contact `json:"contact"`
	Start   int       `json:"start"`
	Count   int       `json:"count"`
	Total   int       `json:"total"`
}
