// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>

package http

import liberrors "kilabit.info/pakakeh.go/lib/errors"

// EndpointResponse defines common container that can be used by Server
// implementor.
// Its embed the [liberrors.E] to work seamlessly with [Endpoint.Call]
// handler for checking the returned error.
//
// If the response is paging, contains more than one item in Data, one
// can set the current status of paging in field Limit, Offset, Page, and
// Count.
//
// See the example below on how to use it with [Endpoint.Call] handler.
type EndpointResponse struct {
	Data any `json:"data,omitempty"`

	liberrors.E

	// For paging, return the maximum number of records per page.
	Limit int64 `json:"limit,omitempty"`

	// For paging, the start index of paging.
	// If Page values is from request then the offset is set to
	// (Page * Limit).
	Offset int64 `json:"offset,omitempty"`

	// For paging, return the requested or current page of response.
	Page int64 `json:"page,omitempty"`

	// For paging, return the total number of records in Data.
	Count int64 `json:"count,omitempty"`

	// For paging, return the total number of all records.
	Total int64 `json:"total,omitempty"`
}

// Error implements the builtin error interface that returns string from
// [*liberrors.E.Message]
func (epr *EndpointResponse) Error() string {
	return epr.E.Error()
}

// Unwrap returns the error as instance of [*liberrors.E].
func (epr *EndpointResponse) Unwrap() (err error) {
	return &epr.E
}
