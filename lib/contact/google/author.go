// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Author define Google contacts author.
type Author struct {
	Name  GD `json:"name,omitempty"`
	Email GD `json:"email,omitempty"`
}
