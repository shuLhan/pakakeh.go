// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"testing"

	"kilabit.info/pakakeh.go/lib/test"
)

type testVectorLocal struct {
	Name       string
	Key        string
	Nonce      string
	Token      string
	Payload    string
	Footer     string
	Implicit   string `json:"implicit-assertion"`
	key        [32]byte
	salt       [32]byte
	ExpectFail bool `json:"expect-fail"`
}

func (tvl *testVectorLocal) init() {
	v, err := hex.DecodeString(tvl.Key)
	if err != nil {
		log.Fatalf(`key: %s`, err)
	}
	copy(tvl.key[:], v)

	v, err = hex.DecodeString(tvl.Nonce)
	if err != nil {
		log.Fatalf(`key: %s`, err)
	}
	copy(tvl.salt[:], v)
}

func TestNewLocalMode(t *testing.T) {
	v, err := hex.DecodeString(`707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f`)
	if err != nil {
		t.Fatal(err)
	}
	lmode := NewLocalMode([32]byte(v))
	hex := lmode.Hex()
	test.Assert(t, `LocalMode hex length`, 64, len(hex))
}

func TestLocalMode_Encrypt(t *testing.T) {
	logp := `TestLocalMode_Encrypt`

	localb, err := os.ReadFile(`testdata/local.json`)
	if err != nil {
		t.Fatalf(`%s: %s`, logp, err)
	}
	listCase := []testVectorLocal{}
	err = json.Unmarshal(localb, &listCase)
	if err != nil {
		t.Fatalf(`%s: %s`, logp, err)
	}

	for _, tc := range listCase {
		tc.init()
		lmode := NewLocalMode(tc.key)
		got, err := lmode.encrypt(tc.salt, []byte(tc.Payload),
			[]byte(tc.Footer), []byte(tc.Implicit))
		if err != nil {
			if tc.ExpectFail {
				continue
			}
			t.Fatalf(`%s: %s`, logp, err)
		}
		if tc.Token == string(got) {
			continue
		}
		if tc.ExpectFail {
			continue
		}
		t.Fatalf(`%s: token not match`, tc.Name)
	}
}

func TestLocalMode_Decrypt(t *testing.T) {
	logp := `TestLocalMode_Decrypt`

	localb, err := os.ReadFile(`testdata/local.json`)
	if err != nil {
		t.Fatalf(`%s: %s`, logp, err)
	}
	listCase := []testVectorLocal{}
	err = json.Unmarshal(localb, &listCase)
	if err != nil {
		t.Fatalf(`%s: %s`, logp, err)
	}

	for _, tc := range listCase {
		tc.init()
		lmode := NewLocalMode(tc.key)
		gotPlain, gotFooter, err := lmode.Decrypt(tc.Token, []byte(tc.Implicit))
		if err != nil {
			if tc.ExpectFail {
				continue
			}
			t.Fatalf(`%s: %s`, logp, err)
		}
		if tc.Payload != string(gotPlain) && tc.ExpectFail {
			continue
		}
		if tc.Footer != string(gotFooter) && tc.ExpectFail {
			continue
		}
		test.Assert(t, tc.Name+` payload`, tc.Payload, string(gotPlain))
		test.Assert(t, tc.Name+` footer`, tc.Footer, string(gotFooter))
	}
}
