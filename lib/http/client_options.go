// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 M. Shulhan <ms@kilabit.info>

package http

import (
	"net/http"
	"strings"
	"time"
)

const (
	defClientTimeout = 10 * time.Second
)

// ClientOptions options for HTTP client.
type ClientOptions struct {
	// Default headers that will be send in any request.
	// This field is optional.
	Headers http.Header

	// The server address without path, for example "https://example.com"
	// or "http://10.148.0.12:8080".
	// This value should not changed during call of client's method.
	// This field is required.
	ServerURL string

	// Set the http Transport Timeout and TLSHandshakeTimeout.
	// This field is optional, default to 10 seconds.
	Timeout time.Duration

	// Allow to connect to server with unknown or invalid certificate
	// authority.
	// This field is optional.
	AllowInsecure bool
}

func (opts *ClientOptions) init() {
	if opts.Headers == nil {
		opts.Headers = make(http.Header)
	}
	if opts.Timeout <= 0 {
		opts.Timeout = defClientTimeout
	}
	opts.ServerURL = strings.TrimSuffix(opts.ServerURL, `/`)
}
