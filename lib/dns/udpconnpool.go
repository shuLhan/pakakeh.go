// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"log"
	"net"
	"sync"
)

var udpConnPool = sync.Pool{
	New: func() interface{} {
		conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: nil, Port: 0})
		if err != nil {
			log.Fatal("net.ListenPacket:", err)
			return nil
		}

		return conn
	},
}
