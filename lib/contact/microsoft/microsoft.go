// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

// Package microsoft implement Microsoft's Live contact API v1.0.
//
// # Reference
//
// - https://developer.microsoft.com/en-us/graph/docs/api-reference/v1.0/api/user_list_contacts
package microsoft

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"git.sr.ht/~shulhan/pakakeh.go/lib/contact"
)

const (
	// List of provider APIs.
	apiContactsURL = "https://graph.microsoft.com/v1.0/me/contacts"
)

// ImportFromJSON will parse Microsoft Live's JSON contact response and return
// list of contact on success.
func ImportFromJSON(jsonb []byte) (
	contacts []*contact.Record,
	err error,
) {
	root := &Root{}

	err = json.Unmarshal(jsonb, root)
	if err != nil {
		return
	}

	for _, mscontact := range root.Contacts {
		contact := mscontact.Decode()
		contacts = append(contacts, contact)
	}

	return
}

// ImportWithOAuth get Microsoft Live contacts using OAuth HTTP client.
func ImportWithOAuth(client *http.Client) (contacts []*contact.Record, err error) {
	var (
		ctx = context.Background()
		req *http.Request
	)

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, apiContactsURL, nil)
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
