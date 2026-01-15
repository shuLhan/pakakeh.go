// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package yahoo

// Root define the root of JSON response.
type Root struct {
	Contacts Contacts `json:"contacts"`
}
