// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package totp_test

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/totp"
)

func ExampleProtocol_GenerateNWithTime() {
	var (
		secretHex = `3132333435363738393031323334353637383930`

		secret []byte
		err    error
	)

	secret, err = hex.DecodeString(secretHex)
	if err != nil {
		log.Fatal(err)
	}

	var (
		proto = totp.New(totp.CryptoHashSHA1, totp.DefCodeDigits, totp.DefTimeStep)
		ts    = time.Date(2024, time.January, 29, 23, 37, 0, 0, time.UTC)

		listOTP []string
	)

	listOTP, err = proto.GenerateNWithTime(ts, secret, 3)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(listOTP)
	// Output:
	// [933840 870583 802638]
}

func ExampleProtocol_GenerateWithTime() {
	var (
		secretHex = `3132333435363738393031323334353637383930`

		secret []byte
		err    error
	)

	secret, err = hex.DecodeString(secretHex)
	if err != nil {
		log.Fatal(err)
	}

	var (
		proto = totp.New(totp.CryptoHashSHA1, totp.DefCodeDigits, totp.DefTimeStep)
		ts    = time.Date(2024, time.January, 29, 23, 37, 0, 0, time.UTC)

		otp string
	)

	otp, err = proto.GenerateWithTime(ts, secret)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(otp)
	// Output:
	// 933840
}

func ExampleProtocol_Verify() {
	var (
		secretHex = `3132333435363738393031323334353637383930`

		err    error
		secret []byte
	)

	secret, err = hex.DecodeString(secretHex)
	if err != nil {
		log.Fatal(err)
	}

	var (
		proto = totp.New(totp.CryptoHashSHA1, totp.DefCodeDigits, totp.DefTimeStep)

		otp string
	)

	otp, _ = proto.Generate(secret)

	if proto.Verify(secret, otp, 1) {
		fmt.Println(`Generated token is valid.`)
	} else {
		fmt.Printf(`Generated token is not valid.`)
	}

	// Output:
	// Generated token is valid.
}
