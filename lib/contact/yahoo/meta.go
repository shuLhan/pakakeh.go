// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package yahoo

import (
	"time"
)

// Meta define a common metadata inside a struct.
type Meta struct {
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	URI     string    `json:"uri"`
}
