// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// ResponseType define the content type for HTTP response.
type ResponseType string

// List of valid response type.
const (
	ResponseTypeNone   ResponseType = ``
	ResponseTypeBinary ResponseType = `binary`
	ResponseTypeHTML   ResponseType = `html`
	ResponseTypeJSON   ResponseType = `json`
	ResponseTypePlain  ResponseType = `plain`
	ResponseTypeXML    ResponseType = `xml`
)

// String return the string representation of ResponseType as in
// "Content-Type" header.
// For ResponseTypeNone it will return an empty string.
func (restype ResponseType) String() string {
	switch restype {
	case ResponseTypeNone:
		return ``
	case ResponseTypeBinary:
		return ContentTypeBinary
	case ResponseTypeHTML:
		return ContentTypeHTML
	case ResponseTypeJSON:
		return ContentTypeJSON
	case ResponseTypePlain:
		return ContentTypePlain
	case ResponseTypeXML:
		return ContentTypeXML
	}
	return ``
}
