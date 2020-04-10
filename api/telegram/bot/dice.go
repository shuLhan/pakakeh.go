// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// Dice represents a dice with random value from 1 to 6. (Yes, we're aware of
// the “proper” singular of die. But it's awkward, and we decided to help it
// change. One dice at a time!)
type Dice struct {
	Value int `json:"value"` // Value of the dice, 1-6
}
