// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"net/http"

	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
)

// FSHandler define the function to inspect each GET request to Server
// [memfs.MemFS] instance.
// The node parameter contains the requested file inside the memfs.
//
// If the handler return true, server will continue processing the node
// (writing the [memfs.Node] content type, body, and so on).
//
// If the handler return false, server stop processing the node and return
// immediately, which means the function should have already handle writing
// the header, status code, and/or body.
type FSHandler func(node *memfs.Node, res http.ResponseWriter, req *http.Request) bool
