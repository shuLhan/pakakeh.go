// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package yahoo

import (
	"encoding/json"

	"github.com/shuLhan/share/lib/contact"
)

// Contact define the contact item in response.
type Contact struct {
	Fields []Field `json:"fields"`

	// Ignored fields for speedup.
	//ID           int        `json:"id"`
	//IsConnection bool       `json:"isConnection"`
	//Error        int        `json:"error"`
	//RestoredID   int        `json:"restoredId"`
	//Categories []Category `json:"categories"`
	//Meta
}

// Decode will convert the interface value in each field into its struct
// representation.
func (c *Contact) Decode() (to *contact.Record) {
	to = &contact.Record{}

	for x, field := range c.Fields {
		field.Decode(to)

		// Clear the Value to minimize memory usage.
		c.Fields[x].Value = nil
	}

	return
}

// ParseJSON will parse JSON input and return contact.Record object on
// success.
//
// On fail it will return nil and error.
func ParseJSON(jsonb []byte) (to *contact.Record, err error) {
	ycontact := &Contact{}

	err = json.Unmarshal(jsonb, ycontact)
	if err != nil {
		return
	}

	to = ycontact.Decode()

	return
}
