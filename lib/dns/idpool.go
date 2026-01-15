// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package dns

import (
	"sync/atomic"
)

var idPool uint32

// getNextID increment and return ID.
func getNextID() uint16 {
	atomic.AddUint32(&idPool, 1)
	var id = atomic.LoadUint32(&idPool)

	return uint16(id)
}

// getID return the current ID value in pool.
func getID() uint16 {
	var id = atomic.LoadUint32(&idPool)
	return uint16(id)
}

func resetIDPool() {
	atomic.StoreUint32(&idPool, 0)
}
