// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yahoo

import (
	"io/ioutil"
	"testing"

	"github.com/shuLhan/share/lib/contact"
	"github.com/shuLhan/share/lib/test"
)

const (
	sampleContact = "testdata/contact.json"
)

func parseSampleJSON(t *testing.T, input string) (contact *contact.Record) {
	jsonb, err := ioutil.ReadFile(input)
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

	test.Assert(t, "Name", exp.Name, gotContact.Name, true)
	test.Assert(t, "Birthday", exp.Birthday, gotContact.Birthday, true)
	test.Assert(t, "Addresses", exp.Addresses, gotContact.Addresses, true)
	test.Assert(t, "Anniversary", exp.Anniversary, gotContact.Anniversary, true)
	test.Assert(t, "Emails", exp.Emails, gotContact.Emails, true)
	test.Assert(t, "Phones", exp.Phones, gotContact.Phones, true)
	test.Assert(t, "Links", exp.Links, gotContact.Links, true)
	test.Assert(t, "Company", exp.Company, gotContact.Company, true)
	test.Assert(t, "JobTitle", exp.JobTitle, gotContact.JobTitle, true)
}
