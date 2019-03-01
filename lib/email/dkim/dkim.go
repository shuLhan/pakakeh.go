// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"errors"
)

//
// DefaultNameServers contains list of nameserver's IP addresses.
//
// If its not empty, the public key lookup using DNS/TXT will use this values;
// otherwise it will try to use the system name servers.
//
var DefaultNameServers []string // nolint:gochecknoglobals

//
// DefaultKeyPool contains cached DKIM key.
//
// Implementor of this library can use the KeyPool.Get method to retrieve key
// instead of LookupKey to minimize network traffic and process to decode and
// parse public key.
//
var DefaultKeyPool = &KeyPool{ // nolint:gochecknoglobals
	pool: make(map[string]*Key),
}

var ( // nolint:gochecknoglobals
	sepColon = []byte{':'} //nolint:gochecknoglobals
	sepSlash = []byte{'/'} //nolint:gochecknoglobals
	sepVBar  = []byte{'|'} //nolint:gochecknoglobals
)

var (
	errEmptySignAlg   = errors.New("dkim: tag algorithm 'a=' is empty")
	errEmptySDID      = errors.New("dkim: tag SDID 'd=' is empty")
	errEmptySelector  = errors.New("dkim: tag selector 's=' is empty")
	errEmptyHeader    = errors.New("dkim: tag header 'h=' is empty")
	errEmptyBodyHash  = errors.New("dkim: tag body hash 'bh=' is empty")
	errEmptySignature = errors.New("dkim: tag signature 'h=' is empty")
	errFromHeader     = errors.New("dkim: 'From' field is not in header tag")
	errCreatedTime    = errors.New("dkim: invalid expiration/creation time")
)
