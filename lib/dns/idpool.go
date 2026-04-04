// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package dns

import (
	"sync/atomic"
)

var idPool atomic.Uint32

// getNextID increment and return ID.
func getNextID() uint16 {
	idPool.Add(1)
	var id = idPool.Load()

	return uint16(id)
}

// getID return the current ID value in pool.
func getID() uint16 {
	var id = idPool.Load()
	return uint16(id)
}

func resetIDPool() {
	idPool.Store(0)
}
