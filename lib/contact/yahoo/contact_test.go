// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package yahoo

import (
	"os"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/contact"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

const (
	sampleContact = "testdata/contact.json"
)

func parseSampleJSON(t *testing.T, input string) (contact *contact.Record) {
	jsonb, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}

	contact, err = ParseJSON(jsonb)
	if err != nil {
		t.Fatal(err)
	}

	return
}

func TestParseJSON(t *testing.T) {
	exp := &contact.Record{
		Name: contact.Name{
			Given:  "Test",
			Middle: "Middle",
			Family: "Proofn",
		},
		Birthday: &contact.Date{
			Day:   "24",
			Month: "1",
			Year:  "1980",
		},
		Emails: []contact.Email{{
			Address: "test@proofn.com",
		}},
		Phones: []contact.Phone{{
			Type:   "home",
			Number: "084-563-21",
		}, {
			Type:   "mobile",
			Number: "084-563-20",
		}, {
			Type:   "work",
			Number: "084-563-23",
		}},
		Links: []string{
			"www.proofn.com",
		},
		Company:  "Myabuy",
		JobTitle: "Tester",
	}

	gotContact := parseSampleJSON(t, sampleContact)

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
