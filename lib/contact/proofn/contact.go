// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package proofn

import (
	"strings"

	"github.com/shuLhan/share/lib/contact"
)

//
// Contact define Proofn contact format.
//
type Contact struct {
	Title      string `json:"title,omitempty"`
	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Suffix     string `json:"suffix,omitempty"`
	FullName   string `json:"fullName,omitempty"`
	Gender     int32  `json:"gender,omitempty"`
	Birthday   string `json:"birthday,omitempty"`

	Email       string `json:"email,omitempty"`
	EmailTitle  string `json:"emailTitle,omitempty"`
	Email2      string `json:"email2,omitempty"`
	Email2Title string `json:"email2Title,omitempty"`
	Email3      string `json:"email3,omitempty"`
	Email3Title string `json:"email3Title,omitempty"`

	HomeAddress  string `json:"homeAddress,omitempty"`
	HomeCity     string `json:"homeCity,omitempty"`
	HomeProvince string `json:"homeProvince,omitempty"`
	HomeCountry  string `json:"homeCountry,omitempty"`
	HomeZIP      string `json:"homeZIP,omitempty"`
	HomePhone    string `json:"homePhone,omitempty"`

	WorkAddress  string `json:"workAddress,omitempty"`
	WorkCity     string `json:"workCity,omitempty"`
	WorkProvince string `json:"workProvince,omitempty"`
	WorkCountry  string `json:"workCountry,omitempty"`
	WorkZIP      string `json:"workZIP,omitempty"`
	WorkPhone    string `json:"workPhone,omitempty"`

	Office       string `json:"office,omitempty"`
	JobTitle     string `json:"jobTitle,omitempty"`
	Profession   string `json:"profession,omitempty"`
	Department   string `json:"department,omitempty"`
	ManagersName string `json:"managersName,omitempty"`

	DialCode       string `json:"dialCode,omitempty"`
	PhoneNumber    string `json:"phoneNumber,omitempty"`
	MobilePhone    string `json:"mobilePhone,omitempty"`
	AlternatePhone string `json:"alternatePhone,omitempty"`
	CompanyPhone   string `json:"companyPhone,omitempty"`

	SpouseType  int32    `json:"spouseType,omitempty"`
	SpouseName  string   `json:"spouseName,omitempty"`
	Anniversary string   `json:"anniversary,omitempty"`
	KidsName    []string `json:"kidsName,omitempty"`

	TopicsToAvoid   string   `json:"topicsToAvoid,omitempty"`
	TopicsToMention string   `json:"topicsToMention,omitempty"`
	FollowUpTopic   string   `json:"followUpTopic,omitempty"`
	Notes           []string `json:"notes,omitempty"`

	IsVIP    bool `json:"isVIP,omitempty"`
	IsActive bool `json:"isActive,omitempty"`

	Avatar *Avatar `json:"avatar,omitempty"`
}

//
// SetAddresses will set contact address using value from list of contact
// Address.
//
func (c *Contact) SetAddresses(addresses []contact.Address) {
	for _, adr := range addresses {
		switch adr.Type {
		case contact.TypeWork:
			c.WorkAddress = adr.Street
			c.WorkCity = adr.City
			c.WorkProvince = adr.StateOrProv
			c.WorkCountry = adr.Country
			c.WorkZIP = adr.PostalCode

		default:
			c.HomeAddress = adr.Street
			c.HomeCity = adr.City
			c.HomeProvince = adr.StateOrProv
			c.HomeCountry = adr.Country
			c.HomeZIP = adr.PostalCode
		}
	}
}

//
// SetEmails will set Proofn contact email based on type on gonctacts email.
//
func (c *Contact) SetEmails(emails []contact.Email) {
	for _, email := range emails {
		switch email.Type {
		case contact.TypeWork:
			c.Email2 = email.Address
			c.Email2Title = email.Type
		default:
			c.Email = email.Address
			c.EmailTitle = email.Type
		}
	}
}

//
// SetFullName will reset the contact full name based on values in title,
// first, middle, last name, and suffix.
//
func (c *Contact) SetFullName() {
	c.FullName = c.Title

	if c.FirstName != "" {
		c.FullName += " " + c.FirstName
	}
	if c.MiddleName != "" {
		c.FullName += " " + c.MiddleName
	}
	if c.LastName != "" {
		c.FullName += " " + c.LastName
	}
	if c.Suffix != "" {
		c.FullName += " " + c.Suffix
	}

	c.FullName = strings.TrimSpace(c.FullName)
}

//
// SetName will set Proofn contact name according to value in contact Name.
//
func (c *Contact) SetName(name *contact.Name) {
	if name == nil {
		return
	}

	c.FirstName = strings.TrimSpace(name.Given)
	c.MiddleName = strings.TrimSpace(name.Middle)
	c.LastName = strings.TrimSpace(name.Family)
	c.Title = strings.TrimSpace(name.Prefix)
	c.Suffix = strings.TrimSpace(name.Suffix)

	c.SetFullName()
}

//
// SetPhones will set Proofn contact phone based on contact Phone type and
// number.
//
func (c *Contact) SetPhones(phones []contact.Phone) {
	for _, phone := range phones {
		switch phone.Type {
		case contact.TypeWork:
			c.WorkPhone = phone.Number

		case contact.TypeMobile:
			c.MobilePhone = phone.Number

		default:
			c.PhoneNumber = phone.Number
		}
	}
}
