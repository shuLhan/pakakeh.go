// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package google implement Google's contact API v3.
//
package google

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/shuLhan/share/lib/contact"
)

const (
	// List of APIs
	apiContactsURL = "https://www.google.com/m8/feeds/contacts/default/full?alt=json&max-results=50000&v=3.0"
)

//
// ImportFromJSON will parse JSON input and return Contacts object on success.
//
// On fail it will return nil and error.
//
func ImportFromJSON(jsonb []byte) (contacts []*contact.Record, err error) {
	root := &Root{}

	err = json.Unmarshal(jsonb, root)
	if err != nil {
		return
	}

	for _, gcontact := range root.Feed.Contacts {
		contact := gcontact.Decode()
		contacts = append(contacts, contact)
	}

	return
}

//
// ImportWithOAuth will send a request to user's contact API using OAuth
// authentication code.
//
// On success it will return pointer to OAuth token and list of contacts, with
// nil error.
//
// On fail, it will return nil token, empty contacts, and error.
//
func ImportWithOAuth(
	tokenType string,
	accessToken string,
) (
	contacts []*contact.Record,
	err error,
) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", apiContactsURL, nil)
	if err != nil {
		return
	}

	req.Header.Add("Authorization", tokenType+" "+accessToken)

	res, err := client.Do(req)
	if err != nil {
		return
	}

	resBody, err := ioutil.ReadAll(res.Body)
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
