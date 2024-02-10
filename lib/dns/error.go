// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import "errors"

// errUnpack define an error if packet failed to be parsed.
var errUnpack = errors.New(`unpack: invalid message`)
