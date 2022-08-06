// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package totp

import (
	"encoding/hex"
	"fmt"
	"log"
)

func ExampleProtocol_Verify() {
	var (
		secretHex = `3132333435363738393031323334353637383930`
		proto     = New(CryptoHashSHA1, DefCodeDigits, DefTimeStep)

		otp    string
		err    error
		secret []byte
	)

	secret, err = hex.DecodeString(secretHex)
	if err != nil {
		log.Fatal(err)
	}

	otp, _ = proto.Generate(secret)

	if proto.Verify(secret, otp, 1) {
		fmt.Println(`Generated token is valid.`)
	} else {
		fmt.Printf(`Generated token is not valid.`)
	}

	// Output:
	// Generated token is valid.
}
