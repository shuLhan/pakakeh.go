package http

import liberrors "github.com/shuLhan/share/lib/errors"

//
// EndpointResponse is one of the common HTTP response container that can be
// used by Server implementor.
// Its embed the lib/errors.E type to work seamlessly with Endpoint.Call
// handler for checking the returned error.
//
// If the response is paging, contains more than one item in data, one
// can set the current status of paging in field Offset, Limit, and Count.
//
// The Offset field contains the start index of paging.
// The Limit field contains the maximum number of records per page.
// The Count field contains the total number of records.
//
// See the example below on how to use it with Endpoint.Call handler.
//
type EndpointResponse struct {
	liberrors.E
	Data   interface{} `json:"data,omitempty"`
	Offset int64       `json:"offset,omitempty"`
	Limit  int64       `json:"limit,omitempty"`
	Count  int64       `json:"count,omitempty"`
}

//
// Unwrap return the error as instance of *liberror.E.
//
func (epr *EndpointResponse) Unwrap() (err error) {
	return &epr.E
}
