// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 Shulhan <ms@kilabit.info>

package http

import "net/http"

// EndpointRequest wrap the called [Endpoint] and common two parameters in
// HTTP handler: the [http.ResponseWriter] and [http.Request].
//
// The RequestBody field contains the full [http.Request.Body] that has been
// read.
//
// The Error field is used by [CallbackErrorHandler].
type EndpointRequest struct {
	HTTPWriter  http.ResponseWriter
	Error       error
	Endpoint    *Endpoint
	HTTPRequest *http.Request
	RequestBody []byte
}
