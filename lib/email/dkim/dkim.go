// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

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
	sepColon = []byte{':'}
	sepVBar  = []byte{'|'}
)
