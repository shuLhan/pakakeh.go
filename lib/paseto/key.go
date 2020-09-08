// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package paseto

import "crypto/ed25519"

type Key struct {
	id      string
	private ed25519.PrivateKey
	public  ed25519.PublicKey
}

//
// NewKey create new Key from hex encoded strings.
//
func NewKey(id string, private ed25519.PrivateKey, public ed25519.PublicKey) Key {
	return Key{
		id:      id,
		private: private,
		public:  public,
	}
}
