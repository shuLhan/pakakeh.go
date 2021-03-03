// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	liberrors "github.com/shuLhan/share/lib/errors"
)

//
// CallbackErrorHandler define the function that can be used to handle an
// error returned from Endpoint.Call.
// By default, if Endpoint.Call is nil, it will use DefaultErrorHandler.
//
type CallbackErrorHandler func(epr *EndpointRequest)

//
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
//
func DefaultErrorHandler(epr *EndpointRequest) {
	errInternal := &liberrors.E{}
	if errors.As(epr.Error, &errInternal) {
		if errInternal.Code <= 0 || errInternal.Code >= 512 {
			errInternal.Code = http.StatusInternalServerError
		}
	} else {
		log.Printf("DefaultErrorHandler: %d %s %s %s\n",
			http.StatusInternalServerError,
			epr.HttpRequest.Method, epr.HttpRequest.URL.Path, epr.Error)

		errInternal = liberrors.Internal(epr.Error)
	}

	epr.HttpWriter.Header().Set(HeaderContentType, ContentTypeJSON)
	epr.HttpWriter.WriteHeader(errInternal.Code)

	rsp, err := json.Marshal(errInternal)
	if err != nil {
		log.Println("DefaultErrorHandler: " + err.Error())
		return
	}

	_, err = epr.HttpWriter.Write(rsp)
	if err != nil {
		log.Println("DefaultErrorHandler: " + err.Error())
	}
}
