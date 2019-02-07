// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

//
// DefaultNameServers contains list of nameserver's IP addresses.
// If its not empty, the public key lookup using DNS/TXT will use this values.
//
var DefaultNameServers []string // nolint: gochecknoglobals

var ( // nolint: gochecknoglobals
	sepColon      = []byte{':'}
	sepVBar       = []byte{'|'}
	dkimSubdomain = []byte("_domainkey")
)
