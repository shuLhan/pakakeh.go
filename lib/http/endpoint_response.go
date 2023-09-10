// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import liberrors "github.com/shuLhan/share/lib/errors"

// EndpointResponse is one of the common HTTP response container that can be
// used by Server implementor.
// Its embed the lib/errors.E type to work seamlessly with Endpoint.Call
// handler for checking the returned error.
//
// If the response is paging, contains more than one item in Data, one
// can set the current status of paging in field Limit, Offset, Page, and
// Count.
//
// See the example below on how to use it with Endpoint.Call handler.
type EndpointResponse struct {
	Data interface{} `json:"data,omitempty"`

	liberrors.E

	// The Limit field contains the maximum number of records per page.
	Limit int64 `json:"limit,omitempty"`

	// The Offset field contains the start index of paging.
	// If Page values is from request then the offset can be set to
	// Page times Limit.
	Offset int64 `json:"offset,omitempty"`

	// The Page field contains the requested or current page of response.
	Page int64 `json:"page,omitempty"`

	// Count field contains the total number of records in Data.
	Count int64 `json:"count,omitempty"`

	// Total field contains the total number of all records.
	Total int64 `json:"total,omitempty"`
}

func (epr *EndpointResponse) Error() string {
	return epr.E.Error()
}

// Unwrap return the error as instance of *liberror.E.
func (epr *EndpointResponse) Unwrap() (err error) {
	return &epr.E
}
