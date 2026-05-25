// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package http

import (
	"encoding/json"
	"errors"
	"net/http"

	liberrors "kilabit.info/pakakeh.go/lib/errors"
	"kilabit.info/pakakeh.go/lib/mlog"
)

// CallbackErrorHandler defines the function type that can be used to handle
// an error returned from [Endpoint.Call].
// By default, if [Endpoint.Call] is nil, it will use [DefaultErrorHandler].
type CallbackErrorHandler func(epr *EndpointRequest)

// DefaultErrorHandler handles the error from [Endpoint.Call].
// This is the default error handler if [Endpoint.ErrorHandler] is not set.
//
// If [EndpointRequest.Error] is an instance of [*liberrors.E], it will set
// the HTTP status code to [liberrors.E.Code] and response body to
// [liberrors.E.Message].
// Otherwise, it will set to HTTP status code to
// [http.StatusInternalServerError] with message "internal server error".
//
// The response set the HTTP header Content-Type based on
// [Endpoint.ResponseType], unless it set to [ResponseTypeNone].
//
// For [ResponseTypeBinary], [ResponseTypeHTML], [ResponseTypePlain], and
// [ResponseTypeXML] it will write the error message in the body as plain
// text.
// For [ResponseTypeJSON] it will writes the following body,
//
//	{"code":<E.Code>, "message":"<E.Message>"}
func DefaultErrorHandler(epr *EndpointRequest) {
	logp := `DefaultErrorHandler`
	errInternal := &liberrors.E{}

	if errors.As(epr.Error, &errInternal) {
		if errInternal.Code <= 0 || errInternal.Code >= 512 {
			errInternal.Code = http.StatusInternalServerError
		}
	} else {
		mlog.Errf(`%s: %s %s: %s`, logp, epr.HTTPRequest.Method,
			epr.HTTPRequest.URL.Path, epr.Error)

		errInternal = liberrors.Internal(epr.Error)
	}

	var body []byte
	var err error
	contentType := epr.Endpoint.ResponseType.String()

	switch epr.Endpoint.ResponseType {
	case ResponseTypeBinary, ResponseTypeHTML, ResponseTypePlain, ResponseTypeXML:
		body = []byte(errInternal.Error())
	case ResponseTypeJSON:
		body, err = json.Marshal(errInternal)
		if err != nil {
			mlog.Errf(`%s: %s`, logp, err)
			// keep going with probably empty body.
		}
	}

	if contentType != `` {
		epr.HTTPWriter.Header().Set(HeaderContentType, contentType)
	}
	epr.HTTPWriter.WriteHeader(errInternal.Code)

	_, err = epr.HTTPWriter.Write(body)
	if err != nil {
		mlog.Errf(`%s: %s`, logp, err)
	}
}
