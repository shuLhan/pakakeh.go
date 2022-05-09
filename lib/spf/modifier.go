// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
