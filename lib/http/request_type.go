// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// RequestType define type of HTTP request.
type RequestType string

// List of valid request type.
const (
	RequestTypeNone          RequestType = ``
	RequestTypeQuery         RequestType = `query`
	RequestTypeForm          RequestType = `form-urlencoded`
	RequestTypeHTML          RequestType = `html`
	RequestTypeMultipartForm RequestType = `form-data`
	RequestTypeJSON          RequestType = `json`
	RequestTypeXML           RequestType = `xml`
)

// String return the string representation of request type as in
// "Content-Type" header.
// For RequestTypeNone or RequestTypeQuery it will return an empty string.
func (rt RequestType) String() string {
	switch rt {
	case RequestTypeNone, RequestTypeQuery:
		return ``
	case RequestTypeForm:
		return ContentTypeForm
	case RequestTypeHTML:
		return ContentTypeHTML
	case RequestTypeMultipartForm:
		return ContentTypeMultipartForm
	case RequestTypeJSON:
		return ContentTypeJSON
	case RequestTypeXML:
		return ContentTypeXML
	}
	return ``
}
