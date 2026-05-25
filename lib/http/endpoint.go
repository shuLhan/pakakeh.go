// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

package http

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"

	liberrors "kilabit.info/pakakeh.go/lib/errors"
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

	// The type of request.
	// By setting this field, server will reject any request that does not
	// match based on Content-Type with 400 Bad Request, except when set
	// to [RequestTypeNone] and [RequestTypeQuery] it will accept any
	// content-type.
	//
	// If set to [RequestTypeForm], [RequestTypeQuery], or
	// [RequestTypeJSON] server call [http.Request.ParseForm] to parse the
	// query parameters or body for [http.Request.Form]
	//
	// If set to [RequestTypeMultipartForm] server call
	// [http.Request.ParseMultipartForm].
	//
	// This field is optional, default to RequestTypeNone.
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
	logp := `Endpoint.call`

	epr := &EndpointRequest{
		Endpoint:    ep,
		HTTPWriter:  res,
		HTTPRequest: req,
	}

	switch ep.RequestType {
	case RequestTypeNone, RequestTypeQuery:
		// Allow any content-type.
	default:
		contentType := req.Header.Get(HeaderContentType)
		if !strings.HasPrefix(contentType, ep.RequestType.String()) {
			epr.Error = &liberrors.E{
				Code:    http.StatusBadRequest,
				Message: `invalid content type`,
			}
			ep.ErrorHandler(epr)
			return
		}
	}

	var err error

	epr.RequestBody, err = io.ReadAll(req.Body)
	if err != nil {
		mlog.Errf("%s: ReadAll: %s", logp, err)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(epr.RequestBody))

	switch ep.RequestType {
	case RequestTypeNone, RequestTypeHTML, RequestTypeXML:
		// NOOP.

	case RequestTypeForm, RequestTypeQuery, RequestTypeJSON:
		err = req.ParseForm()

	case RequestTypeMultipartForm:
		err = req.ParseMultipartForm(0)
	}
	if err != nil {
		mlog.Errf("%s: %s %s: request parse: %s", logp, req.Method, req.URL.Path, err)
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

	var responseBody []byte

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
			mlog.Errf("%s: %s %s: response write: %s", logp, req.Method, req.URL.Path, err)
			break
		}
		nwrite += n
	}
}
