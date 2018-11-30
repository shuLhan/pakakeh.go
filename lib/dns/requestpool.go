// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"sync"
)

var requestPool = sync.Pool{ // nolint: gochecknoglobals
	New: func() interface{} {
		return NewRequest()
	},
}

//
// AllocRequest allocate new request.
//
func AllocRequest() *Request {
	return requestPool.Get().(*Request)
}

//
// FreeRequest put the request back to the pool.
//
func FreeRequest(req *Request) {
	if req.ChanResponded != nil {
		close(req.ChanResponded)
		req.ChanResponded = nil
	}
	req.Reset()
	requestPool.Put(req)
}
