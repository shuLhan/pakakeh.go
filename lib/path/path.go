// Copyright 2024, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package path implements utility routines for manipulating slash-separated
// paths.
package path

import "errors"

// ErrPathKeyEmpty define an error when path contains an empty
// key, for example "/:/y".
var ErrPathKeyEmpty = errors.New(`empty path key`)

// ErrPathKeyDuplicate define an error when registering path with
// the same keys, for example "/:x/:x".
var ErrPathKeyDuplicate = errors.New(`duplicate key in path`)
