// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package google

import (
	"git.sr.ht/~shulhan/pakakeh.go/lib/contact"
)

// Contact define a single Google contact data.
//
// Some of the fields are disabled for speed.
type Contact struct {
	// Ignored fields for speedup.

	// ID         GD
	// ETag       string     `json:"gd$etag,omitempty"`
	// Updated    GD         `json:"updated,omitempty"`
	// Edited     GD         `json:"app$edited,omitempty"`
	// Categories []Category `json:"category,omitempty"`
	// Title      GD         `json:"title,omitempty"`
	// Links      []Link     `json:"link,omitempty"`

	Name      Name      `json:"gd$name,omitempty"`
	Birthday  Birthday  `json:"gContact$birthday,omitempty"`
	Orgs      []Org     `json:"gd$organization,omitempty"`
	Emails    []Email   `json:"gd$email,omitempty"`
	Phones    []Phone   `json:"gd$phoneNumber,omitempty"`
	Addresses []Address `json:"gd$structuredPostalAddress,omitempty"`
	Events    []Event   `json:"gContact$event,omitempty"`
	Websites  []Link    `json:"gContact$website,omitempty"`
}

func (gc *Contact) decodeOrg(contact *contact.Record) {
	if len(gc.Orgs) == 0 {
		return
	}

	contact.Company = gc.Orgs[0].Name.Value
	contact.JobTitle = gc.Orgs[0].JobTitle.Value
}

func (gc *Contact) decodeEmails(to *contact.Record) {
	for _, email := range gc.Emails {
		decodedEmail := contact.Email{
			Type:    ParseRel(email.Rel),
			Address: email.Address,
		}
		to.Emails = append(to.Emails, decodedEmail)
	}
}

func (gc *Contact) decodePhones(to *contact.Record) {
	for _, phone := range gc.Phones {
		decodedPhone := contact.Phone{
			Type:   ParseRel(phone.Rel),
			Number: phone.Number,
		}
		to.Phones = append(to.Phones, decodedPhone)
	}
}

func (gc *Contact) decodeAddresses(to *contact.Record) {
	for _, adr := range gc.Addresses {
		decAdr := contact.Address{
			Type:        ParseRel(adr.Rel),
			POBox:       adr.POBox.Value,
			Street:      adr.Street.Value,
			City:        adr.City.Value,
			StateOrProv: adr.StateOrProv.Value,
			PostalCode:  adr.PostalCode.Value,
			Country:     adr.Country.Value,
		}

		to.Addresses = append(to.Addresses, decAdr)
	}
}

func (gc *Contact) decodeEvents(to *contact.Record) {
	for _, event := range gc.Events {
		if event.Rel == contact.TypeAnniversary {
			to.SetAnniversary(event.When.Start)
		}
	}
}

func (gc *Contact) decodeWebsites(to *contact.Record) {
	for _, site := range gc.Websites {
		to.Links = append(to.Links, site.HRef)
	}
}

// Decode will convert Google's Contact to our Contact format.
func (gc *Contact) Decode() (to *contact.Record) {
	to = &contact.Record{
		Name: contact.Name{
			Given:  gc.Name.First.Value,
			Middle: gc.Name.Middle.Value,
			Family: gc.Name.Last.Value,
			Prefix: gc.Name.Prefix.Value,
			Suffix: gc.Name.Suffix.Value,
		},
	}

	to.SetBirthday(gc.Birthday.When)

	gc.decodeOrg(to)
	gc.decodeEmails(to)
	gc.decodePhones(to)
	gc.decodeAddresses(to)
	gc.decodeEvents(to)
	gc.decodeWebsites(to)

	return
}
