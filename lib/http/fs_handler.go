// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2022 Shulhan <ms@kilabit.info>

package http

import (
	"net/http"

	"kilabit.info/pakakeh.go/lib/memfs"
)

// FSHandler defines the function to inspect each GET request to Server
// using the [ServerOptions.Memfs] instance.
//
// The node parameter contains the requested file inside the [memfs.MemFS] or
// nil if the file does not exist.
//
// The handler returns two values: the node out that is used to process the
// request and response; and the HTTP statusCode for response.
//
// Non-zero statusCode indicates that the handler already response to the
// request, and the server will stop processing it.
//
// Zero statusCode indicates that the handler did not process the request,
// it is up to server to process the returned node out.
// The returned node out may be the same as node argument, modified, of
// completely new.
//
// If the node (or out) exist, server write the [memfs.Node.Content] in the
// body with content-type set to [memfs.Node.ContentType] and with header ETag
// set to [memfs.Node.ModTime].
type FSHandler func(node *memfs.Node, res http.ResponseWriter, req *http.Request) (
	out *memfs.Node, statusCode int)
