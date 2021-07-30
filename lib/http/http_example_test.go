// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"net/http"
)

func ExampleIPAddressOfRequest() {
	defAddress := "192.168.100.1"

	headers := http.Header{
		"X-Real-Ip": []string{"127.0.0.1"},
	}
	fmt.Println("Request with X-Real-IP:", IPAddressOfRequest(headers, defAddress))

	headers = http.Header{
		"X-Forwarded-For": []string{"127.0.0.2, 192.168.100.1"},
	}
	fmt.Println("Request with X-Forwarded-For:", IPAddressOfRequest(headers, defAddress))

	headers = http.Header{}
	fmt.Println("Request without X-* headers:", IPAddressOfRequest(headers, defAddress))

	// Output:
	// Request with X-Real-IP: 127.0.0.1
	// Request with X-Forwarded-For: 127.0.0.2
	// Request without X-* headers: 192.168.100.1
}

func ExampleParseXForwardedFor() {
	values := []string{
		"",
		"203.0.113.195",
		"203.0.113.195, 70.41.3.18, 150.172.238.178",
		"2001:db8:85a3:8d3:1319:8a2e:370:7348",
	}
	for _, val := range values {
		clientAddr, proxyAddrs := ParseXForwardedFor(val)
		fmt.Println(clientAddr, proxyAddrs)
	}
	// Output:
	// []
	// 203.0.113.195 []
	// 203.0.113.195 [70.41.3.18 150.172.238.178]
	// 2001:db8:85a3:8d3:1319:8a2e:370:7348 []
}
