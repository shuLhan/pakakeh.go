// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bytes"
	"testing"
)

func TestGenerateHandshakeKey(t *testing.T) {
	var key = generateHandshakeKey()
	if len(key) != 24 {
		t.Fatalf(`expecting random 24 characters key, got %d`, len(key))
	}

	var key2 = generateHandshakeKey()
	if bytes.Equal(key2, key) {
		t.Fatalf(`expecting new key %s is not equal with previous %s`, key2, key)
	}
}
