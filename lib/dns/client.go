// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"time"
)

//
// Client is interface that implement sending and receiving DNS message.
//
type Client interface {
	Close() error
	Lookup(q MessageQuestion, allowRecursion bool) (*Message, error)
	Query(req *Message) (*Message, error)
	RemoteAddr() string
	SetRemoteAddr(addr string) error
	SetTimeout(t time.Duration)
}
