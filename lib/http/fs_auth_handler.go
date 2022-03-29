// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import "net/http"

//
// FSAuthHandler define the function to authorized each GET request to
// Server MemFS using value from the HTTP Request instance.
//
// If request is not authorized it must return false and the HTTP response
// will be set to 401 Unauthorized with empty body.
//
type FSAuthHandler func(req *http.Request) (ok bool)
