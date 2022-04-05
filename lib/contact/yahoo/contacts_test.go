// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yahoo

import (
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
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
