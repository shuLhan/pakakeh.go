// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

// Package yahoo implement user's contacts import using Yahoo API.
//
// # Reference
//
// - https://developer.yahoo.com/social/rest_api_guide/contacts-resource.html
package yahoo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"git.sr.ht/~shulhan/pakakeh.go/lib/contact"
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
	var (
		ctx = context.Background()
		api = apiContactsURL + guid + apiContactsSuffix
		req *http.Request
	)

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, api, nil)
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
