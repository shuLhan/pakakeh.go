// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package http

import (
	"net/http"
)

// Evaluator evaluate the request.
// If request is invalid, the error will tell the response code and the
// error message to be written back to client.
type Evaluator func(req *http.Request, reqBody []byte) error
