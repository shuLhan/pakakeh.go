// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

//
// Storage define an interface for storing and retrieving mail object into
// permanent storage (for example, file system or database).
//
type Storage interface {
	Delete(id string) error
	Load(id string) (mail *MailTx, err error)
	LoadAll() (mail []*MailTx, err error)
	Bounce(id string) error
	Store(mail *MailTx) error
}
