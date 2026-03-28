// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package pasetov2

// PublicToken contains the unpacked public token.
type PublicToken struct {
	Token  JSONToken
	Footer JSONFooter
	Data   []byte
}
