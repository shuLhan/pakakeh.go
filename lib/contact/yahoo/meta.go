// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
