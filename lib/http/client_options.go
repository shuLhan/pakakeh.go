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
	// Headers define default headers that will be send in any request to
	// server.
	// This field is optional.
	Headers http.Header

	// ServerUrl define the server address without path, for example
	// "https://example.com" or "http://10.148.0.12:8080".
	// This value should not changed during call of client's method.
	// This field is required.
	ServerUrl string

	// Timeout affect the http Transport Timeout and TLSHandshakeTimeout.
	// This field is optional, if not set it will set to 10 seconds.
	Timeout time.Duration

	// AllowInsecure if its true, it will allow to connect to server with
	// unknown certificate authority.
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
	opts.ServerUrl = strings.TrimSuffix(opts.ServerUrl, `/`)
}
