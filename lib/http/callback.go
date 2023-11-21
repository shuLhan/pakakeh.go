// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

// Callback define a type of function for handling registered handler.
//
// The function will have the query URL, request multipart form data,
// and request body ready to be used in [EndpointRequest.HttpRequest] and
// [EndpointRequest.RequestBody] fields.
//
// The [EndpointRequest.HttpWriter] can be used to write custom header or to
// write cookies but should not be used to write response body.
//
// The error return type should be an instance of *errors.E, with E.Code
// define the HTTP status code.
// If error is not nil and not *[liberrors.E], server will response with
// [http.StatusInternalServerError] status code.
type Callback func(req *EndpointRequest) (resBody []byte, err error)
