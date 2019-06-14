// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"sync/atomic"
)

//nolint:gochecknoglobals
var idPool uint32

//
// getNextID increment and return ID.
//
func getNextID() uint16 {
	atomic.AddUint32(&idPool, 1)
	id := atomic.LoadUint32(&idPool)

	return uint16(id)
}

//
// getID return the current ID value in pool.
//
func getID() uint16 {
	id := atomic.LoadUint32(&idPool)
	return uint16(id)
}
