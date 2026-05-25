// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package http

import (
	"net/http"
)

// Evaluator evaluates the request before [Endpoint.Call].
// If request is invalid, the returned error contains the response code and
// the error message to be written back to client, in that way the error
// should be the instance of [liberrors.E]
type Evaluator func(req *http.Request, reqBody []byte) error
