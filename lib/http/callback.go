// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package http

// Callback defines function type for handling request in [Endpoint.Call].
//
// The function will have the query URL, request multipart form data,
// and request body ready to be used in [EndpointRequest.HTTPRequest] and
// [EndpointRequest.RequestBody] fields.
//
// The [EndpointRequest.HTTPWriter] can be used to write custom header or to
// write cookies but should not be used to write response body.
//
// The error return type should be an instance of *[liberrors.E], with
// [liberrors.E.Code] define the HTTP status code.
// If error is not nil and not *[liberrors.E], server will response with
// [http.StatusInternalServerError] status code.
type Callback func(req *EndpointRequest) (resBody []byte, err error)
