// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

package main

import (
	"fmt"
	"log"

	"git.sr.ht/~shulhan/pakakeh.go/lib/systemd"
)

func main() {
	lns, err := systemd.Listeners(true)
	if err != nil {
		log.Fatal(err)
	}
	for _, ln := range lns {
		fmt.Printf(`listener network address: %s`, ln.Addr())
	}
}
