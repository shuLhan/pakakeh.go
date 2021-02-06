package http

import (
	"fmt"
	"net/http"
)

func ExampleIPAddressOfRequest() {
	reqWithXRealIP := &http.Request{
		Header: http.Header{
			"X-Real-Ip": []string{"127.0.0.1"},
		},
		RemoteAddr: "192.168.100.1",
	}
	fmt.Println("Request with X-Real-IP:", IPAddressOfRequest(reqWithXRealIP))

	reqWithXForwardedFor := &http.Request{
		Header: http.Header{
			"X-Forwarded-For": []string{"127.0.0.2, 192.168.100.1"},
		},
		RemoteAddr: "192.168.100.1",
	}
	fmt.Println("Request with X-Forwarded-For:", IPAddressOfRequest(reqWithXForwardedFor))

	reqWithRemoteAddr := &http.Request{
		Header:     http.Header{},
		RemoteAddr: "127.0.0.3",
	}
	fmt.Println("Request without X-* headers:", IPAddressOfRequest(reqWithRemoteAddr))

	// Output:
	// Request with X-Real-IP: 127.0.0.1
	// Request with X-Forwarded-For: 127.0.0.2
	// Request without X-* headers: 127.0.0.3
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
