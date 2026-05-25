// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package http

import "net/http"

// EndpointRequest wrap the called [Endpoint] and common two parameters in
// HTTP handler: the [http.ResponseWriter] and [http.Request].
// The instance is passed to [Endpoint.Call] on each request.
type EndpointRequest struct {
	// The HTTP writer from [Server.ServeHTTP].
	HTTPWriter http.ResponseWriter

	// The returned error value from [Endpoint.Call].
	// Used by [DefaultErrorHandler] or any custom
	// [Endpoint.ErrorHandler].
	Error error

	// The Endpoint that handle the request.
	Endpoint *Endpoint

	// The HTTP request from [Server.ServeHTTP].
	HTTPRequest *http.Request

	// Contains the full [http.Request.Body] that has been read.
	RequestBody []byte
}
