// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package pasetov4

import (
	"encoding/hex"
	"fmt"
	"log"
)

func ExampleLocalMode() {
	secret := `707172737475767778797a7b7c7d7e7f808182838485868788898a8b8c8d8e8f`
	key, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatal(err)
	}
	lmode := NewLocalMode([32]byte(key))
	plain := []byte(`{"data":"Hello, secret!"}`)
	footer := []byte(`{"kid":1000}`)
	token, err := lmode.Encrypt(plain, footer, nil)
	if err != nil {
		log.Fatal(err)
	}

	gotPlain, gotFooter, err := lmode.Decrypt(token, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(gotPlain))
	fmt.Println(string(gotFooter))

	// Output:
	// {"data":"Hello, secret!"}
	// {"kid":1000}
}
