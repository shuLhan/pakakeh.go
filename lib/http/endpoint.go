// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

package http

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	"kilabit.info/pakakeh.go/lib/mlog"
)

// Endpoint represents route handled by server.
// Endpoint can be set by calling [Server.RegisterEndpoint].
// Each route have their own evaluator that will be evaluated after global
// evaluators from server.
type Endpoint struct {
	// Function that handle the error from Call.
	// This field is optional, default to [DefaultErrorHandler].
	ErrorHandler CallbackErrorHandler

	// Evaluator for route that is called after global evaluators and
	// before Call.
	Eval Evaluator

	// The main function of route.
	Call Callback

	// Define HTTP method, default to GET.
	Method RequestMethod

	// Route to be served, default to "/".
	Path string

	// The type of request, default to RequestTypeNone.
	// RequestTypeNone and RequestTypeQuery accept any content-type.
	// Client request with content-type that does not match with defines
	// RequestType will be rejected.
	//
	// If set to [RequestTypeForm] or [RequestTypeQuery]
	// server call [http.Request.ParseForm] to parse the request body into
	// [http.Request.Form].
	//
	// If set to [RequestTypeMultipartForm] server call
	// [http.Request.ParseMultipartForm].
	RequestType RequestType

	// Type of response, default to ResponseTypeNone.
	// If set, the content-type set based its String value.
	ResponseType ResponseType
}

func (ep *Endpoint) call(
	res http.ResponseWriter,
	req *http.Request,
	evaluators []Evaluator,
	vals map[string]string,
) {
	var (
		logp = "Endpoint.call"
		epr  = &EndpointRequest{
			Endpoint:    ep,
			HTTPWriter:  res,
			HTTPRequest: req,
		}
		responseBody []byte
		e            error
	)

	epr.RequestBody, e = io.ReadAll(req.Body)
	if e != nil {
		mlog.Errf("%s: ReadAll: %s", logp, e)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(epr.RequestBody))

	switch ep.RequestType {
	case RequestTypeNone, RequestTypeHTML, RequestTypeXML:
		// NOOP.

	case RequestTypeForm, RequestTypeQuery, RequestTypeJSON:
		e = req.ParseForm()

	case RequestTypeMultipartForm:
		e = req.ParseMultipartForm(0)
	}
	if e != nil {
		mlog.Errf("%s: %s %s: request parse: %s", logp, req.Method, req.URL.Path, e)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(vals) > 0 && req.Form == nil {
		req.Form = make(url.Values, len(vals))
	}
	for k, v := range vals {
		if len(k) > 0 && len(v) > 0 {
			req.Form.Set(k, v)
		}
	}

	for _, eval := range evaluators {
		epr.Error = eval(req, epr.RequestBody)
		if epr.Error != nil {
			ep.ErrorHandler(epr)
			return
		}
	}

	if ep.Eval != nil {
		epr.Error = ep.Eval(req, epr.RequestBody)
		if epr.Error != nil {
			ep.ErrorHandler(epr)
			return
		}
	}

	responseBody, epr.Error = ep.Call(epr)
	if epr.Error != nil {
		ep.ErrorHandler(epr)
		return
	}

	switch ep.ResponseType {
	case ResponseTypeNone:
		return
	case ResponseTypeBinary:
		res.Header().Set(HeaderContentType, ContentTypeBinary)
	case ResponseTypeHTML:
		res.Header().Set(HeaderContentType, ContentTypeHTML)
	case ResponseTypeJSON:
		res.Header().Set(HeaderContentType, ContentTypeJSON)
	case ResponseTypePlain:
		res.Header().Set(HeaderContentType, ContentTypePlain)
	case ResponseTypeXML:
		res.Header().Set(HeaderContentType, ContentTypeXML)
	}

	var nwrite int
	for nwrite < len(responseBody) {
		n, err := res.Write(responseBody[nwrite:])
		if err != nil {
			mlog.Errf("%s: %s %s: response write: %s", logp, req.Method, req.URL.Path, e)
			break
		}
		nwrite += n
	}
}
