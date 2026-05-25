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

// DefaultErrorHandler handles the error from [Endpoint.Call], if the
// [Endpoint.ErrorHandler] is not defined.
//
// It will check if [EndpointRequest.Error] is an instance of *[liberrors.E].
// If true, it will use the [liberrors.E.Code] value for HTTP status code,
// otherwise if its zero or invalid, it will set to
// [http.StatusInternalServerError].
//
// The response set the HTTP header Content-Type to "application/json"
// and write the body as JSON format,
//
//	{"code":<E.Code>, "message":"<E.Message | err.Error()>"}
func DefaultErrorHandler(epr *EndpointRequest) {
	var (
		logp        = "DefaultErrorHandler"
		errInternal = &liberrors.E{}

		jsonb []byte
		err   error
	)

	if errors.As(epr.Error, &errInternal) {
		if errInternal.Code <= 0 || errInternal.Code >= 512 {
			errInternal.Code = http.StatusInternalServerError
		}
	} else {
		mlog.Errf(`%s: %s %s: %s`, logp, epr.HTTPRequest.Method,
			epr.HTTPRequest.URL.Path, epr.Error)

		errInternal = liberrors.Internal(epr.Error)
	}

	epr.HTTPWriter.Header().Set(HeaderContentType, ContentTypeJSON)
	epr.HTTPWriter.WriteHeader(errInternal.Code)

	jsonb, err = json.Marshal(errInternal)
	if err != nil {
		mlog.Errf("%s: json.Marshal: %s", logp, err)
		return
	}

	_, err = epr.HTTPWriter.Write(jsonb)
	if err != nil {
		mlog.Errf("%s: Write: %s", logp, err)
	}
}
