// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

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
