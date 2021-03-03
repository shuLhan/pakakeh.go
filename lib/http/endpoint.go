// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/shuLhan/share/lib/debug"
)

//
// Endpoint represent route that will be handled by server.
// Each route have their own evaluator that will be evaluated after global
// evaluators from server.
//
type Endpoint struct {
	// Method contains HTTP method, default to GET.
	Method RequestMethod

	// Path contains route to be served, default to "/" if its empty.
	Path string

	// RequestType contains type of request, default to RequestTypeNone.
	RequestType RequestType

	// ResponseType contains type of request, default to ResponseTypeNone.
	ResponseType ResponseType

	// Eval define evaluator for route that will be called after global
	// evaluators and before callback.
	Eval Evaluator

	// Call is the main process of route.
	Call Callback

	// ErrorHandler define the function that will handle the error
	// returned from Call.
	ErrorHandler CallbackErrorHandler
}

//
// HTTPMethod return the string representation of HTTP method as predefined
// in "net/http".
//
func (ep *Endpoint) HTTPMethod() string {
	switch ep.Method {
	case RequestMethodGet:
		return http.MethodGet
	case RequestMethodConnect:
		return http.MethodConnect
	case RequestMethodDelete:
		return http.MethodDelete
	case RequestMethodHead:
		return http.MethodHead
	case RequestMethodOptions:
		return http.MethodOptions
	case RequestMethodPatch:
		return http.MethodPatch
	case RequestMethodPost:
		return http.MethodPost
	case RequestMethodPut:
		return http.MethodPut
	case RequestMethodTrace:
		return http.MethodTrace
	}
	return http.MethodGet
}

func (ep *Endpoint) call(
	res http.ResponseWriter,
	req *http.Request,
	evaluators []Evaluator,
	vals map[string]string,
) {
	var (
		epr = &EndpointRequest{
			Endpoint:    ep,
			HttpWriter:  res,
			HttpRequest: req,
		}
		responseBody []byte
		e            error
	)

	epr.RequestBody, e = ioutil.ReadAll(req.Body)
	if e != nil {
		log.Printf("endpoint.call: " + e.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(epr.RequestBody))

	switch ep.RequestType {
	case RequestTypeForm, RequestTypeQuery, RequestTypeJSON:
		e = req.ParseForm()

	case RequestTypeMultipartForm:
		e = req.ParseMultipartForm(0)
	}

	if e != nil {
		log.Printf("endpoint.call: %d %s %s %s\n",
			http.StatusBadRequest, req.Method, req.URL.Path, e)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if debug.Value >= 3 {
		log.Printf("> request body: %s\n", epr.RequestBody)
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
		res.WriteHeader(http.StatusNoContent)
		return
	case ResponseTypeBinary:
		res.Header().Set(HeaderContentType, ContentTypeBinary)
	case ResponseTypeJSON:
		res.Header().Set(HeaderContentType, ContentTypeJSON)
	case ResponseTypePlain:
		res.Header().Set(HeaderContentType, ContentTypePlain)
	}

	var nwrite int
	for nwrite < len(responseBody) {
		n, err := res.Write(responseBody[nwrite:])
		if err != nil {
			log.Printf("endpoint.call: %s %s %s\n", req.Method,
				req.URL.Path, e)
			break
		}
		nwrite += n
	}
}
