// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"sync/atomic"
)

var idPool uint32

//
// getID return a new ID for message header.
//
func getID() uint16 {
	atomic.AddUint32(&idPool, 1)
	id := atomic.LoadUint32(&idPool)

	return uint16(id)
}
