// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestProtocol_generateWithTimestamp_sha1(t *testing.T) {
	secretHex := "3132333435363738393031323334353637383930"

	secretb, err := hex.DecodeString(secretHex)
	if err != nil {
		t.Fatal(err)
	}

	mac := hmac.New(sha1.New, secretb)
	p := New(CryptoHashSHA1, 8, DefTimeStep)

	cases := []struct {
		time int64
		exp  string
	}{{
		time: 59,
		exp:  "94287082",
	}, {
		time: 1111111109,
		exp:  "07081804",
	}, {
		time: 1111111111,
		exp:  "14050471",
	}, {
		time: 1234567890,
		exp:  "89005924",
	}, {
		time: 2000000000,
		exp:  "69279037",
	}, {
		time: 20000000000,
		exp:  "65353130",
	}}

	for _, c := range cases {
		mac.Reset()
		got, err := p.generateWithTimestamp(mac, c.time)
		if err != nil {
			t.Error(err)
			continue
		}
		test.Assert(t, "generateWithTimestamp", c.exp, got)
	}
}

func TestProtocol_generateWithTimestamp_sha256(t *testing.T) {
	secretHex := "3132333435363738393031323334353637383930" +
		"313233343536373839303132"

	secretb, err := hex.DecodeString(secretHex)
	if err != nil {
		t.Fatal(err)
	}

	mac := hmac.New(sha256.New, secretb)
	p := New(CryptoHashSHA256, 8, DefTimeStep)

	cases := []struct {
		time int64
		exp  string
	}{{
		time: 59,
		exp:  "46119246",
	}, {
		time: 1111111109,
		exp:  "68084774",
	}, {
		time: 1111111111,
		exp:  "67062674",
	}, {
		time: 1234567890,
		exp:  "91819424",
	}, {
		time: 2000000000,
		exp:  "90698825",
	}, {
		time: 20000000000,
		exp:  "77737706",
	}}

	for _, c := range cases {
		mac.Reset()
		got, err := p.generateWithTimestamp(mac, c.time)
		if err != nil {
			t.Error(err)
			continue
		}
		test.Assert(t, "generateWithTimestamp", c.exp, got)
	}
}

func TestProtocol_generateWithTimestamp_sha512(t *testing.T) {
	secretHex := "3132333435363738393031323334353637383930" +
		"3132333435363738393031323334353637383930" +
		"3132333435363738393031323334353637383930" +
		"31323334"

	secretb, err := hex.DecodeString(secretHex)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		time int64
		exp  string
	}{{
		time: 59,
		exp:  "90693936",
	}, {
		time: 1111111109,
		exp:  "25091201",
	}, {
		time: 1111111111,
		exp:  "99943326",
	}, {
		time: 1234567890,
		exp:  "93441116",
	}, {
		time: 2000000000,
		exp:  "38618901",
	}, {
		time: 20000000000,
		exp:  "47863826",
	}}

	p := New(CryptoHashSHA512, 8, DefTimeStep)
	mac := hmac.New(sha512.New, secretb)

	for _, c := range cases {
		mac.Reset()
		got, err := p.generateWithTimestamp(mac, c.time)
		if err != nil {
			t.Error(err)
			continue
		}
		test.Assert(t, "generateWithTimestamp sha512", c.exp, got)
	}
}

func TestProtocol_verifyWithTimestamp(t *testing.T) {
	secretHex := "3132333435363738393031323334353637383930"

	secretb, err := hex.DecodeString(secretHex)
	if err != nil {
		t.Fatal(err)
	}

	mac := hmac.New(sha1.New, secretb)
	p := New(CryptoHashSHA1, 8, DefTimeStep)

	cases := []struct {
		desc  string
		ts    int64
		token string
		steps int
		exp   bool
	}{{
		desc:  "With OTP ~ timestamp",
		ts:    2000000000,
		token: "69279037",
		steps: 2,
		exp:   true,
	}, {
		desc:  "With previous OTP timestamp",
		ts:    2000000000 - DefTimeStep,
		token: "40196847",
		steps: 2,
		exp:   true,
	}, {
		desc:  "With timestamp + DefTimeStep",
		ts:    2000000000 + DefTimeStep,
		token: "69279037",
		steps: 2,
		exp:   true,
	}}

	for _, c := range cases {
		got := p.verifyWithTimestamp(mac, c.token, c.steps, c.ts)
		test.Assert(t, c.desc, c.exp, got)
	}
}
