// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

import (
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

const (
	sampleContacts = "testdata/contacts.json"
)

func TestImportFromJSON(t *testing.T) {
	jsonb, err := os.ReadFile(sampleContacts)
	if err != nil {
		t.Fatal(err)
	}

	contacts, err := ImportFromJSON(jsonb)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Len", 55, len(contacts))
}
