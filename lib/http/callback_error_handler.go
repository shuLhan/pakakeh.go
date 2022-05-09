// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"errors"
	"net/http"

	liberrors "github.com/shuLhan/share/lib/errors"
	"github.com/shuLhan/share/lib/mlog"
)

// CallbackErrorHandler define the function that can be used to handle an
// error returned from Endpoint.Call.
// By default, if Endpoint.Call is nil, it will use DefaultErrorHandler.
type CallbackErrorHandler func(epr *EndpointRequest)

// DefaultErrorHandler define the default function that will called to handle
// the error returned from Callback function, if the Endpoint.ErrorHandler is
// not defined.
//
// First, it will check if error instance of *errors.E. If its true, it will
// use the Code value for HTTP status code, otherwise if its zero or invalid,
// it will set to http.StatusInternalServerError.
//
// Second, it will set the HTTP header Content-Type to "application/json" and
// write the response body as JSON format,
//
//	{"code":<HTTP_STATUS_CODE>, "message":<err.Error()>}
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
		mlog.Errf("%s: %s %s: %s", logp, epr.HttpRequest.Method,
			epr.HttpRequest.URL.Path, epr.Error)

		errInternal = liberrors.Internal(epr.Error)
	}

	epr.HttpWriter.Header().Set(HeaderContentType, ContentTypeJSON)
	epr.HttpWriter.WriteHeader(errInternal.Code)

	jsonb, err = json.Marshal(errInternal)
	if err != nil {
		mlog.Errf("%s: json.Marshal: %s", logp, err)
		return
	}

	_, err = epr.HttpWriter.Write(jsonb)
	if err != nil {
		mlog.Errf("%s: Write: %s", logp, err)
	}
}
