// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

package http

import (
	"net/http"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
)

// FSHandler define the function to inspect each GET request to Server
// [memfs.MemFS] instance.
// The node parameter contains the requested file inside the memfs or nil
// if the file does not exist.
//
// This function return two values: the node `out` that is used to process the
// request and response; and the HTTP status code `statusCode` returned in
// response.

// Non-zero status code indicates that the function already response
// to the request, and the server will return immediately.
//
// Zero status code indicates that the function did not process the request,
// it is up to server to process the returned node `out`.
// The returned node `out` may be the same as `node`, modified, of completely
// new.
type FSHandler func(node *memfs.Node, res http.ResponseWriter, req *http.Request) (
	out *memfs.Node, statusCode int)
