// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package contact

// Date define contact's with type "birthday" and "anniversary".
type Date struct {
	Day   string `json:"day"`
	Month string `json:"month"`
	Year  string `json:"year"`
}

// String will return the string representation of date object in `YYYY-MM-DD`
// format.
func (date *Date) String() (r string) {
	if date.Year == "" {
		return
	}

	r = date.Year + "-" + date.Month + "-" + date.Day

	return
}

// VCardString will return the string representation of date object in VCard
// format: `YYYYMMDD` or "--MMDD" if year is empty.
func (date *Date) VCardString() (r string) {
	if date.Day == "" || date.Month == "" {
		return
	}

	if date.Year == "" {
		r = "--" + date.Month + date.Day
	} else {
		r = date.Year + date.Month + date.Day
	}

	return
}
