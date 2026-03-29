// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"encoding/hex"
	"fmt"
	"log"
)

func ExamplePublicMode() {
	secret := `b4cbfb43df4ce210727d953e4a713307fa19bb7d9f85041438d9e11b942a3774`
	seed, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatal(err)
	}
	pmode := NewPublicMode(seed)

	plain := []byte(`{"data":"signed message!"}`)
	footer := []byte(`{"kid":1000}`)
	token, err := pmode.Sign(plain, footer, nil)
	if err != nil {
		log.Fatal(err)
	}

	gotPlain, gotFooter, err := pmode.Verify(token, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(gotPlain))
	fmt.Println(string(gotFooter))

	// Output:
	// {"data":"signed message!"}
	// {"kid":1000}
}
