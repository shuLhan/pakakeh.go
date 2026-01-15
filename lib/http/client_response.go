// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>

package http

import "net/http"

// ClientResponse contains the response from HTTP client request.
type ClientResponse struct {
	HTTPResponse *http.Response
	Body         []byte
}
