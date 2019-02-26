// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"strings"
)

//
// Account represent an SMTP account in the server that can send and receive
// email.
//
type Account struct {
	Name  string
	Local string
	Pass  string
}

//
// NewAccount create new account.
// Password must be already hashed with SHA256.
// An account with empty password is system account, which mean it will not
// allowed in SMTP AUTH.
//
func NewAccount(name, local, pass string) (acc *Account) {
	local = strings.ToLower(local)

	return &Account{
		Name:  name,
		Local: local,
		Pass:  pass,
	}
}
