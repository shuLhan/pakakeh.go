// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package microsoft

import (
	"github.com/shuLhan/share/lib/contact"
)

// Contact define Microsoft Live's contact format.
//
// Some of the fields are disabled for speed up.
type Contact struct {
	//ETag string `json:"@odata.etag,omitempty"`
	//Id   string `json:"id,omitempty"`
	//Created string `json:"createdDateTime,omitempty"`
	//LastModified string `json:"lastModifiedDateTime,omitempty"`
	//ChangeKey string `json:"changeKey,omitempty"`
	//Categories []string `json:"categories,omitempty"`
	//ParentFolderID string `json:"parentFolderId,omitempty"`
	//FileAs string `json:"fileAs,omitempty"`

	Birthday string `json:"birthday,omitempty"`

	DisplayName string `json:"displayName,omitempty"`
	GivenName   string `json:"givenName,omitempty"`
	Initials    string `json:"initials,omitempty"`
	MiddleName  string `json:"middleName,omitempty"`
	NickName    string `json:"nickName,omitempty"`
	SurName     string `json:"surname,omitempty"`
	Title       string `json:"title,omitempty"`
	Generation  string `json:"generation,omitempty"`

	//YomiGivenName string `json:"yomiGivenName,omitempty"`
	//YomiSurname string `json:"yomiSurname,omitempty"`
	//YomiCompanyName string `json:"yomiCompanyName,omitempty"`

	IMAddresses []string `json:"imAddresses,omitempty"`

	JobTitle         string `json:"jobTitle,omitempty"`
	Company          string `json:"companyName,omitempty"`
	Department       string `json:"department,omitempty"`
	OfficeLocation   string `json:"officeLocation,omitempty"`
	Profession       string `json:"profession,omitempty"`
	BusinessHomePage string `json:"businessHomePage,omitempty"`
	AssistantName    string `json:"assistantName,omitempty"`
	Manager          string `json:"manager,omitempty"`

	HomePhones     []string `json:"homePhones,omitempty"`
	MobilePhone    string   `json:"mobilePhone,omitempty"`
	BusinessPhones []string `json:"businessPhones,omitempty"`

	SpouseName    string   `json:"spouseName,omitempty"`
	PersonalNotes string   `json:"personalNotes,omitempty"`
	Children      []string `json:"children,omitempty"`

	Emails []Email `json:"emailAddresses,omitempty"`

	HomeAddress     Address `json:"homeAddress,omitempty"`
	BusinessAddress Address `json:"businessAddress,omitempty"`
	OtherAddress    Address `json:"otherAddress,omitempty"`
}

func (c *Contact) decodeEmails(to *contact.Record) {
	var flag string

	for x, email := range c.Emails {
		switch x {
		case 0:
			flag = contact.TypeMain
		case 1:
			flag = contact.TypeHome
		case 2:
			flag = contact.TypeWork
		default:
			flag = contact.TypeOther
		}

		to.Emails = append(to.Emails, contact.Email{
			Type:    flag,
			Address: email.Address,
		})
	}
}

func (c *Contact) decodePhones(to *contact.Record) {
	if len(c.HomePhones) > 0 {
		to.Phones = append(to.Phones, contact.Phone{
			Type:   contact.TypeHome,
			Number: c.HomePhones[0],
		})
	}

	if len(c.MobilePhone) > 0 {
		to.Phones = append(to.Phones, contact.Phone{
			Type:   contact.TypeMobile,
			Number: c.MobilePhone,
		})
	}

	if len(c.BusinessPhones) > 0 {
		to.Phones = append(to.Phones, contact.Phone{
			Type:   contact.TypeWork,
			Number: c.BusinessPhones[0],
		})
	}
}

func (c *Contact) decodeLinks(to *contact.Record) {
	if len(c.IMAddresses) > 0 {
		to.Links = append(to.Links, c.IMAddresses...)
	}

	if len(c.BusinessHomePage) > 0 {
		to.Links = append(to.Links, c.BusinessHomePage)
	}
}

func (c *Contact) decodeNotes(to *contact.Record) {
	if len(c.PersonalNotes) > 0 {
		to.Notes = append(to.Notes, c.PersonalNotes)
	}
}

// Decode will convert Microsoft's Contact to our Contact format.
func (c *Contact) Decode() (to *contact.Record) {
	to = &contact.Record{
		Name: contact.Name{
			Given:  c.GivenName,
			Middle: c.MiddleName,
			Family: c.SurName,
			Prefix: c.Title,
			Suffix: c.Generation,
		},
		Addresses: []contact.Address{{
			Type:        "home",
			Street:      c.HomeAddress.Street,
			City:        c.HomeAddress.City,
			StateOrProv: c.HomeAddress.State,
			PostalCode:  c.HomeAddress.PostalCode,
			Country:     c.HomeAddress.Country,
		}, {
			Type:        "work",
			Street:      c.BusinessAddress.Street,
			City:        c.BusinessAddress.City,
			StateOrProv: c.BusinessAddress.State,
			PostalCode:  c.BusinessAddress.PostalCode,
			Country:     c.BusinessAddress.Country,
		}},
		Company:  c.Company,
		JobTitle: c.JobTitle,
	}

	to.SetBirthday(c.Birthday)

	c.decodeEmails(to)
	c.decodePhones(to)
	c.decodeLinks(to)
	c.decodeNotes(to)

	return to
}
