// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package google

// Feed define content of Google contacts feed metadata.
//
// Some of the fields are disabled for speed.
type Feed struct {
	// Ignored fields for speedup.

	// XMLNS           string     `json:"xmlns,omitempty"`
	// XMLNSOpenSearch string     `json:"xmlns$openSearch,omitempty"`
	// XMLNSGContact   string     `json:"xmlns$gContact,omitempty"`
	// XMLNSBatch      string     `json:"xmlns$batch,omitempty"`
	// XMLNSGD         string     `json:"xmlns$gd,omitempty"`
	// GDEtag          string     `json:"gd$etag,omitempty"`
	// Id              GD         `json:"id,omitempty"`
	// Updated         GD         `json:"updated,omitempty"`
	// Categories      []Category `json:"category,omitempty"`
	// Title           GD         `json:"title,omitempty"`
	// Links           []Link     `json:"link,omitempty"`
	// Authors         []Author   `json:"author,omitempty"`
	// Generator       Generator  `json:"generator,omitempty"`
	// StartIndex      GD         `json:"openSearch$startIndex,omitempty"`
	// ItemsPerPage    GD         `json:"openSearch$itemsPerPage,omitempty"`

	TotalResult GD        `json:"openSearch$totalResults,omitempty"`
	Contacts    []Contact `json:"entry,omitempty"`
}
