// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package microsoft

import (
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/contact"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

const (
	sampleContacts = "testdata/contacts.json"
)

func TestImportFromJSON(t *testing.T) {
	exp := &contact.Record{
		Name: contact.Name{
			Given:  "First",
			Middle: "Middle",
			Family: "Tester",
			Prefix: "Prof.",
			Suffix: "AMD",
		},
		Birthday: &contact.Date{
			Year:  "1984",
			Month: "08",
			Day:   "14",
		},
		Addresses: []contact.Address{{
			Type:        contact.TypeHome,
			Street:      "Tubagus Ismail VI",
			City:        "Bandung",
			StateOrProv: "JABAR",
			PostalCode:  "40124",
			Country:     "Indonesia",
		}, {
			Type:   contact.TypeWork,
			Street: "Cikutra",
			City:   "Bandung",
		}},
		Emails: []contact.Email{{
			Type:    contact.TypeMain,
			Address: "first.tester@proofn.com",
		}, {
			Type:    contact.TypeHome,
			Address: "tester@proofn.com",
		}},
		Phones: []contact.Phone{{
			Type:   contact.TypeHome,
			Number: "+22808080",
		}, {
			Type:   contact.TypeMobile,
			Number: "+62856123456789",
		}, {
			Type:   contact.TypeWork,
			Number: "+22909090",
		}},
		Notes: []string{
			"This is a note.",
		},
		Company:  "Myabuy",
		JobTitle: "Tester",
	}

	jsonb, err := os.ReadFile(sampleContacts)
	if err != nil {
		t.Fatal(err)
	}

	contacts, err := ImportFromJSON(jsonb)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Len", 1, len(contacts))

	got := contacts[0]

	test.Assert(t, "", exp, got)
}
