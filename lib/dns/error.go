// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import "errors"

// errInvalidMessage define an error when raw DNS message cannot be parsed.
var errInvalidMessage = errors.New(`invalid message`)
