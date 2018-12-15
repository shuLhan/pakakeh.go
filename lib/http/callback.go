// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"
)

//
// Callback define a function type for handling registered handler.
//
// The function will have the query URL, request multipart form data,
// and request body ready to read.
//
type Callback func(req *http.Request, reqBody []byte) ([]byte, error)
