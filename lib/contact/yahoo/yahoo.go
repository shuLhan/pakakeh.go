// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package yahoo implement user's contacts import using Yahoo API.
//
// # Reference
//
// - https://developer.yahoo.com/social/rest_api_guide/contacts-resource.html
package yahoo

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/shuLhan/share/lib/contact"
)

const (
	// List of APIs
	apiContactsURL    = "https://social.yahooapis.com/v1/user/"
	apiContactsSuffix = "/contacts?format=json&count=max"
)

// ImportFromJSON will parse JSON input and return list of Contact on success.
//
// On fail it will return nil and error.
func ImportFromJSON(jsonb []byte) (contacts []*contact.Record, err error) {
	root := &Root{}

	err = json.Unmarshal(jsonb, root)
	if err != nil {
		return
	}

	for _, ycontact := range root.Contacts.Contact {
		contact := ycontact.Decode()
		contacts = append(contacts, contact)
	}

	return
}

// ImportWithOAuth get Yahoo contacts using OAuth HTTP client.
func ImportWithOAuth(client *http.Client, guid string) (contacts []*contact.Record, err error) {
	api := apiContactsURL + guid + apiContactsSuffix
	req, err := http.NewRequest(http.MethodGet, api, nil)
	if err != nil {
		return
	}

	res, err := client.Do(req)
	if err != nil {
		return
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	err = res.Body.Close()
	if err != nil {
		return
	}

	contacts, err = ImportFromJSON(resBody)

	return
}
