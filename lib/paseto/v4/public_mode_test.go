// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"testing"
	"testing/cryptotest"
	"testing/synctest"

	"git.sr.ht/~shulhan/pakakeh.go/lib/paseto"
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

func TestPublicMode_Pack(t *testing.T) {
	cryptotest.SetGlobalRandom(t, 1)

	listcase := []struct {
		desc          string
		expToken      string
		expRawPayload string
		expRawFooter  string
		msg           paseto.Message
	}{{
		desc:          `Empty`,
		expToken:      `v4.public.e316aAoBcV59pCXGKmxwuKl050jOA9gP9F5qCFWllfEHUzzEWqRtZDK29vC07QoCJZEd-6T_65tybe4uJRuErwwE`,
		expRawPayload: `{}`,
	}, {
		desc: `FooterOnly`,
		msg: paseto.Message{
			Footer: paseto.Footer{
				Data: `footer.data`,
			},
		},
		expToken:      `v4.public.e32fZEGqzzkbqRiMbrlQnsacHTa1RjMBuE5cgciyHTEQiUIwh67S106_rGBROlkJKu71axKQURDKqjLMS1E6liQF.eyJkYXRhIjoiZm9vdGVyLmRhdGEifQ`,
		expRawPayload: `{}`,
		expRawFooter:  `{"data":"footer.data"}`,
	}, {
		desc: `PayloadAndFooter`,
		msg: paseto.Message{
			Payload: paseto.Payload{
				Data: []byte{0x73, 0x68, 0x75, 0x6c},
			},
			Footer: paseto.Footer{
				Data: `footer.data`,
			},
		},
		expToken:      `v4.public.eyJkYXRhIjoiYzJoMWJBPT0ifXIYsFygQyakjTQ2j500FIxcPKrE3B3G1Sid7NAvvZ9_yqAZgIhhMu4ojqiewIPw81G-ZvX7GO321Ofgvv8_Zgw.eyJkYXRhIjoiZm9vdGVyLmRhdGEifQ`,
		expRawPayload: `{"data":"c2h1bA=="}`,
		expRawFooter:  `{"data":"footer.data"}`,
	}}

	seed, err := hex.DecodeString(`b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`)
	if err != nil {
		t.Fatal(err)
	}
	pmode := NewPublicMode(``, [32]byte(seed))

	synctest.Test(t, func(t *testing.T) {
		for _, tc := range listcase {
			gotToken, err := pmode.Pack(tc.msg, nil)
			if err != nil {
				t.Fatalf(`%s: %s`, tc.desc, err)
			}
			test.Assert(t, tc.desc, tc.expToken, gotToken)

			msg := paseto.Message{}
			err = UnpackPublicToken(&msg, gotToken, nil)
			if err != nil {
				t.Fatalf(`%s: %s`, tc.desc, err)
			}
			test.Assert(t, `RawPayload`, tc.expRawPayload, string(msg.RawPayload))
			test.Assert(t, `RawFooter`, tc.expRawFooter, string(msg.RawFooter))
		}
	})
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
