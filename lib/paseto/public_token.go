// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

// PublicToken contains the unpacked public token.
type PublicToken struct {
	Token  JSONToken
	Data   []byte
	Footer JSONFooter
}
