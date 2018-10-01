// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

import (
	"io/ioutil"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

const (
	sampleContacts = "testdata/contacts.json"
)

func TestImportFromJSON(t *testing.T) {
	jsonb, err := ioutil.ReadFile(sampleContacts)
	if err != nil {
		t.Fatal(err)
	}

	contacts, err := ImportFromJSON(jsonb)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Len", 55, len(contacts), true)
}
