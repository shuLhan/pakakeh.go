// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Account represent an SMTP account in the server that can send and receive
// email.
type Account struct {
	Mailbox
	// HashPass user password that has been hashed using bcrypt.
	HashPass string
}

// NewAccount create new account.
// Password will be hashed using bcrypt.
// An account with empty password is system account, which mean it will not
// allowed in SMTP AUTH.
func NewAccount(name, local, domain, pass string) (acc *Account, err error) {
	var hpass []byte
	local = strings.ToLower(local)

	if len(pass) > 0 {
		for x := 0; x < 3; x++ {
			hpass, err = bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
			if err == nil {
				break
			}
		}
		if err != nil {
			err = fmt.Errorf("smtp: NewAccount: %s", err.Error())
			return nil, err
		}
	}

	acc = &Account{
		Mailbox: Mailbox{
			Name:   name,
			Local:  local,
			Domain: domain,
		},
		HashPass: string(hpass),
	}

	return acc, nil
}

// Authenticate a user using plain text password.  It will return an error if
// password does not match.
func (acc *Account) Authenticate(pass string) (err error) {
	return bcrypt.CompareHashAndPassword([]byte(acc.HashPass), []byte(pass))
}

// String representation of account in the format of "Name <local@domain>" if
// Name is not empty, or "local@domain" is Name is empty.
func (acc *Account) String() (out string) {
	if len(acc.Name) > 0 {
		out = acc.Name + " <"
	}

	out += acc.Local + "@" + acc.Domain

	if len(acc.Name) > 0 {
		out += ">"
	}
	return
}

// Short return the account email address without Name, "local@domain".
func (acc *Account) Short() (out string) {
	return acc.Local + "@" + acc.Domain
}
