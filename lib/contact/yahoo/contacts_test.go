// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package yahoo

import (
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

const (
	sampleContacts = "testdata/contacts.json"
)

func TestImportFromJSON(t *testing.T) {
	contactsb, err := os.ReadFile(sampleContacts)
	if err != nil {
		t.Fatal(err)
	}

	contacts, err := ImportFromJSON(contactsb)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Len", 54, len(contacts))
}
