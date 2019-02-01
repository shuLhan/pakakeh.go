// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package email

//
// MIME represent part of message body with id, content type, encoding,
// description, and content.
//
type MIME struct {
	ID          []byte
	Type        []byte
	Description []byte
	Content     []byte
}
