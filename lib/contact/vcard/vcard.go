// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package vcard implement RFC6350 for encoding and decoding VCard formatted
// data.
//
package vcard

import (
	"github.com/shuLhan/share/lib/contact"
)

//
// VCard define vcard 4.0 data structure.
//
type VCard struct {
	UID          string
	Source       []string
	Kind         string
	Fn           string
	N            contact.Name
	Nickname     []string
	Photo        []Resource
	Bday         contact.Date
	Anniversary  contact.Date
	Gender       Gender
	Adr          []contact.Address
	Tel          []contact.Phone
	Email        []contact.Email
	Impp         []Messaging
	Lang         []string
	TZ           string
	Geo          []GeoLocation
	Title        []string
	Role         []string
	Logo         []Resource
	Org          []string
	Related      []Relation
	Categories   []string
	Note         []string
	ProdID       string
	Sound        []Resource
	ClientPIDMap string
	Key          []Resource
}
