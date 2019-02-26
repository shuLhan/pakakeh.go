// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

type testStorage struct{}

func (ts *testStorage) MailBounce(id string) (err error) {
	return
}

func (ts *testStorage) MailDelete(id string) (err error) {
	return nil
}

func (ts *testStorage) MailLoad(id string) (mail *MailTx, err error) {
	return nil, nil
}

func (ts *testStorage) MailLoadAll() (mail []*MailTx, err error) {
	return
}

func (ts *testStorage) MailSave(mail *MailTx) (err error) {
	return
}
