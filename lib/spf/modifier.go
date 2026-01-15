// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package spf

// List of modifiers.
const (
	modifierExp      = "exp"
	modifierRedirect = "redirect"
)

type modifier struct {
	name  string
	value string
}
