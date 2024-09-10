// SPDX-FileCopyrightText: 2020 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestProtocol_generateWithTimestamp_sha1(t *testing.T) {
	type testCase struct {
		exp  string
		time int64
	}

	var (
		secretHex = `3132333435363738393031323334353637383930`
		proto     = New(CryptoHashSHA1, 8, DefTimeStep)

		secretb []byte
		err     error
		mac     hash.Hash
	)

	secretb, err = hex.DecodeString(secretHex)
	if err != nil {
		t.Fatal(err)
	}

	mac = hmac.New(sha1.New, secretb)

	var cases = []testCase{{
		time: 59,
		exp:  `94287082`,
	}, {
		time: 1111111109,
		exp:  `07081804`,
	}, {
		time: 1111111111,
		exp:  `14050471`,
	}, {
		time: 1234567890,
		exp:  `89005924`,
	}, {
		time: 2000000000,
		exp:  `69279037`,
	}, {
		time: 20000000000,
		exp:  `65353130`,
	}}

	var (
		c   testCase
		got string
	)
	for _, c = range cases {
		mac.Reset()
		got, err = proto.generateWithTimestamp(mac, c.time)
		if err != nil {
			t.Error(err)
			continue
		}
		test.Assert(t, `generateWithTimestamp`, c.exp, got)
	}
}

func TestProtocol_generateWithTimestamp_sha256(t *testing.T) {
	type testCase struct {
		exp  string
		time int64
	}

	var (
		secretHex = `3132333435363738393031323334353637383930313233343536373839303132`
		proto     = New(CryptoHashSHA256, 8, DefTimeStep)

		mac     hash.Hash
		secretb []byte
		err     error
	)

	secretb, err = hex.DecodeString(secretHex)
	if err != nil {
		t.Fatal(err)
	}

	mac = hmac.New(sha256.New, secretb)

	var cases = []testCase{{
		time: 59,
		exp:  `46119246`,
	}, {
		time: 1111111109,
		exp:  `68084774`,
	}, {
		time: 1111111111,
		exp:  `67062674`,
	}, {
		time: 1234567890,
		exp:  `91819424`,
	}, {
		time: 2000000000,
		exp:  `90698825`,
	}, {
		time: 20000000000,
		exp:  `77737706`,
	}}

	var (
		c   testCase
		got string
	)
	for _, c = range cases {
		mac.Reset()
		got, err = proto.generateWithTimestamp(mac, c.time)
		if err != nil {
			t.Error(err)
			continue
		}
		test.Assert(t, `generateWithTimestamp`, c.exp, got)
	}
}

func TestProtocol_generateWithTimestamp_sha512(t *testing.T) {
	type testCase struct {
		exp  string
		time int64
	}

	var (
		secretHex = `3132333435363738393031323334353637383930` +
			`3132333435363738393031323334353637383930` +
			`3132333435363738393031323334353637383930` +
			`31323334`
		proto = New(CryptoHashSHA512, 8, DefTimeStep)

		mac     hash.Hash
		secretb []byte
		err     error
	)

	secretb, err = hex.DecodeString(secretHex)
	if err != nil {
		t.Fatal(err)
	}

	mac = hmac.New(sha512.New, secretb)

	var cases = []testCase{{
		time: 59,
		exp:  `90693936`,
	}, {
		time: 1111111109,
		exp:  `25091201`,
	}, {
		time: 1111111111,
		exp:  `99943326`,
	}, {
		time: 1234567890,
		exp:  `93441116`,
	}, {
		time: 2000000000,
		exp:  `38618901`,
	}, {
		time: 20000000000,
		exp:  `47863826`,
	}}

	var (
		c   testCase
		got string
	)
	for _, c = range cases {
		mac.Reset()
		got, err = proto.generateWithTimestamp(mac, c.time)
		if err != nil {
			t.Error(err)
			continue
		}
		test.Assert(t, `generateWithTimestamp sha512`, c.exp, got)
	}
}

func TestProtocol_verifyWithTimestamp(t *testing.T) {
	type testCase struct {
		desc  string
		token string
		ts    int64
		steps int
		exp   bool
	}

	var (
		secretHex = `3132333435363738393031323334353637383930`
		proto     = New(CryptoHashSHA1, 8, DefTimeStep)

		mac     hash.Hash
		secretb []byte
		err     error
	)

	secretb, err = hex.DecodeString(secretHex)
	if err != nil {
		t.Fatal(err)
	}

	mac = hmac.New(sha1.New, secretb)

	var cases = []testCase{{
		desc:  `With OTP ~ timestamp`,
		ts:    2000000000,
		token: `69279037`,
		steps: 2,
		exp:   true,
	}, {
		desc:  `With previous OTP timestamp`,
		ts:    2000000000 - DefTimeStep,
		token: `40196847`,
		steps: 2,
		exp:   true,
	}, {
		desc:  `With timestamp + DefTimeStep`,
		ts:    2000000000 + DefTimeStep,
		token: `69279037`,
		steps: 2,
		exp:   true,
	}}

	var (
		c   testCase
		got bool
	)

	for _, c = range cases {
		got = proto.verifyWithTimestamp(mac, c.token, c.steps, c.ts)
		test.Assert(t, c.desc, c.exp, got)
	}
}
