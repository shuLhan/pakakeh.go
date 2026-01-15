// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package microsoft

// Email format on response.
type Email struct {
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}
