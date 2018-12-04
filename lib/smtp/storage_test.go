// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

type testStorage struct{}

func (ts *testStorage) Delete(id string) (err error) {
	return nil
}

func (ts *testStorage) Load(id string) (mail *MailTx, err error) {
	return nil, nil
}

func (ts *testStorage) LoadAll() (mail []*MailTx, err error) {
	return
}

func (ts *testStorage) Bounce(id string) (err error) {
	return
}

func (ts *testStorage) Store(mail *MailTx) (err error) {
	return
}
