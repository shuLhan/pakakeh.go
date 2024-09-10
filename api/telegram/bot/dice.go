// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package bot

// Dice represents a dice with random value from 1 to 6. (Yes, we're aware of
// the “proper” singular of die. But it's awkward, and we decided to help it
// change. One dice at a time!).
type Dice struct {
	Value int `json:"value"` // Value of the dice, 1-6
}
