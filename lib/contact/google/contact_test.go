// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

import (
	"encoding/json"
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/contact"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

const (
	sampleContact = "testdata/contact.json"
)

func parseContact(t *testing.T) (contact *contact.Record) {
	googleContact := &Contact{}

	jsonb, err := os.ReadFile(sampleContact)
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(jsonb, googleContact)
	if err != nil {
		t.Fatal(err)
	}

	return googleContact.Decode()
}

func TestDecode(t *testing.T) {
	exp := &contact.Record{
		Name: contact.Name{
			Given:  "Test",
			Middle: "Middle",
			Family: "Last",
			Prefix: "Prefix",
			Suffix: "Suffix",
		},
		Birthday: &contact.Date{
			Day:   "30",
			Month: "01",
			Year:  "1980",
		},
		Anniversary: &contact.Date{
			Day:   "20",
			Month: "11",
			Year:  "2016",
		},
		Addresses: []contact.Address{
			{
				Type:        "home",
				POBox:       "40124",
				Street:      "Jl. Tubagus Ismail VI",
				City:        "Bandung",
				StateOrProv: "Jabar",
				PostalCode:  "40124",
				Country:     "Indonesia",
			},
			{
				Type:   "work",
				Street: "Perumahan Delima Cikutra",
			},
		},
		Emails: []contact.Email{{
			Type:    "home",
			Address: "first.tester@proofn.com",
		}, {
			Type:    "work",
			Address: "work@proofn.com",
		}},
		Phones: []contact.Phone{{
			Type:   "mobile",
			Number: "856123456789",
		}, {
			Type:   "work",
			Number: "2233445566",
		}, {
			Type:   "home",
			Number: "9999999",
		}, {
			Type:   "main",
			Number: "8888888",
		}},
		Links: []string{
			"https://www.proofn.com",
		},
		Company:  "Myabuy",
		JobTitle: "Devops",
	}

	gotContact := parseContact(t)

	test.Assert(t, "Name", exp.Name, gotContact.Name)
	test.Assert(t, "Birthday", exp.Birthday, gotContact.Birthday)
	test.Assert(t, "Addresses", exp.Addresses, gotContact.Addresses)
	test.Assert(t, "Anniversary", exp.Anniversary, gotContact.Anniversary)
	test.Assert(t, "Emails", exp.Emails, gotContact.Emails)
	test.Assert(t, "Phones", exp.Phones, gotContact.Phones)
	test.Assert(t, "Links", exp.Links, gotContact.Links)
	test.Assert(t, "Company", exp.Company, gotContact.Company)
	test.Assert(t, "JobTitle", exp.JobTitle, gotContact.JobTitle)
}
