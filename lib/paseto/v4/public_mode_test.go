// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

type testVectorPublic struct {
	Name         string
	PublicKey    string `json:"public-key"`
	PublicKeyPEM string `json:"public-key-pem"`
	SecretKey    string `json:"secret-key"`
	SecretKeyPEM string `json:"secret-key-pem"`
	Seed         string `json:"secret-key-seed"`
	Token        string
	Payload      string
	Footer       string
	Implicit     string `json:"implicit-assertion"`
	seed         []byte
	ExpectFail   bool `json:"expect-fail"`
}

func (tvp *testVectorPublic) init() {
	var err error
	tvp.seed, err = hex.DecodeString(tvp.Seed)
	if err != nil {
		log.Fatalf(`init: on seed: %s`, err)
	}
}

func TestPublicMode_Sign(t *testing.T) {
	logp := `TestPublicMode_Sign`

	jsonb, err := os.ReadFile(`testdata/public.json`)
	if err != nil {
		t.Fatalf(`%s: %s`, logp, err)
	}
	listCase := []testVectorPublic{}
	err = json.Unmarshal(jsonb, &listCase)
	if err != nil {
		t.Fatalf(`%s: %s`, logp, err)
	}

	for _, tc := range listCase {
		t.Run(tc.Name, func(tt *testing.T) {
			defer func() {
				if tc.ExpectFail {
					msg := recover()
					tt.Logf(`%s: expect panic: %s`, tc.Name, msg)
				}
			}()

			tc.init()
			pmode := NewPublicMode(`testsign`, [32]byte(tc.seed))
			gotToken, err := pmode.Sign([]byte(tc.Payload), []byte(tc.Footer),
				[]byte(tc.Implicit))
			if err != nil {
				if tc.ExpectFail {
					return
				}
				tt.Fatalf(`%s: %s`, logp, err)
			}
			test.Assert(tt, tc.Name, tc.Token, string(gotToken))
		})
	}
}

func TestPublicMode_Verify(t *testing.T) {
	logp := `TestPublicMode_Verify`

	jsonb, err := os.ReadFile(`testdata/public.json`)
	if err != nil {
		t.Fatalf(`%s: %s`, logp, err)
	}
	listCase := []testVectorPublic{}
	err = json.Unmarshal(jsonb, &listCase)
	if err != nil {
		t.Fatalf(`%s: %s`, logp, err)
	}

	for _, tc := range listCase {
		t.Run(tc.Name, func(tt *testing.T) {
			defer func() {
				if tc.ExpectFail {
					msg := recover()
					tt.Logf(`%s: expect panic: %s`, tc.Name, msg)
				}
			}()

			tc.init()
			pmode := NewPublicMode(`testverify`, [32]byte(tc.seed))
			gotMsg, gotFooter, err := pmode.Verify(tc.Token, []byte(tc.Implicit))
			if err != nil {
				if tc.ExpectFail {
					return
				}
				tt.Fatalf(`%s: %s`, logp, err)
			}
			test.Assert(tt, tc.Name, tc.Payload, string(gotMsg))
			test.Assert(tt, tc.Name, tc.Footer, string(gotFooter))
		})
	}
}
