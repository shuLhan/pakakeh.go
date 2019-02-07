// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	libnet "github.com/shuLhan/share/lib/net"
)

//
// GetSystemNameServers return list of system name servers by reading
// resolv.conf formatted file in path.
//
// Default path is "/etc/resolv.conf".
//
func GetSystemNameServers(path string) []string {
	if len(path) == 0 {
		path = "/etc/resolv.conf"
	}
	rc, err := libnet.NewResolvConf(path)
	if err != nil {
		return nil
	}
	return rc.NameServers
}
