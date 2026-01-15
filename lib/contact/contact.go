// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

// Package contact provide a library to import contact from Google, Microsoft,
// and Yahoo.
package contact

import (
	"strings"
)

// Record define a single contact entity with sane format.
type Record struct {
	Name Name

	Birthday    *Date
	Anniversary *Date

	Company  string
	JobTitle string

	Addresses []Address
	Emails    []Email
	Phones    []Phone
	Links     []string
	Notes     []string
}

// SetBirthday will set contact birthday from string format "YYYY-MM-DD" or
// "YYYY-MM-DDTHH:MM:SSZ".
func (contact *Record) SetBirthday(dateStr string) {
	if dateStr == "" {
		return
	}

	// Split by zone first, and then
	dateTimeZone := strings.Split(dateStr, "T")

	// split the date.
	dates := strings.Split(dateTimeZone[0], "-")
	if len(dates) != 3 {
		return
	}

	contact.Birthday = &Date{
		Year:  dates[0],
		Month: dates[1],
		Day:   dates[2],
	}
}

// SetAnniversary will set contact annivery from string format "YYYY-MM-DD".
func (contact *Record) SetAnniversary(dateStr string) {
	if dateStr == "" {
		return
	}

	dates := strings.Split(dateStr, "-")
	if len(dates) != 3 {
		return
	}

	contact.Anniversary = &Date{
		Year:  dates[0],
		Month: dates[1],
		Day:   dates[2],
	}
}
