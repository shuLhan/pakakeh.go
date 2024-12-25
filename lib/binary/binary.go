// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package binary complement the standard [binary] package.
package binary

import "time"

var timeNow = func() time.Time {
	return time.Now().UTC()
}
