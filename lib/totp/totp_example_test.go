// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package totp

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"log"
)

func ExampleProtocol_Verify() {
	secretHex := "3132333435363738393031323334353637383930"

	secret, err := hex.DecodeString(secretHex)
	if err != nil {
		log.Fatal(err)
	}

	p := New(sha1.New, DefCodeDigits, DefTimeStep)
	otp, _ := p.Generate(secret)

	if p.Verify(secret, otp, 1) {
		fmt.Printf("Generated token is valid.\n")
	} else {
		fmt.Printf("Generated token is not valid.\n")
	}
	//Output:
	//Generated token is valid.
}
