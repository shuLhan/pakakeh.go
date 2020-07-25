// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/errors"
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
	reqBody, e := ioutil.ReadAll(req.Body)
	if e != nil {
		log.Printf("endpoint.call: " + e.Error())
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

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

	if debug.Value >= 2 {
		log.Printf("> request body: %s\n", reqBody)
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
		e = eval(req, reqBody)
		if e != nil {
			ep.error(res, req, e)
			return
		}
	}

	if ep.Eval != nil {
		e = ep.Eval(req, reqBody)
		if e != nil {
			ep.error(res, req, e)
			return
		}
	}

	rspb, e := ep.Call(res, req, reqBody)
	if e != nil {
		ep.error(res, req, e)
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
	for nwrite < len(rspb) {
		n, err := res.Write(rspb[nwrite:])
		if err != nil {
			log.Printf("endpoint.call: %s %s %s\n", req.Method,
				req.URL.Path, e)
			break
		}
		nwrite += n
	}
}

func (ep *Endpoint) error(res http.ResponseWriter, req *http.Request, err error) {
	errInternal := &errors.E{}
	if stderrors.As(err, &errInternal) {
		if errInternal.Code <= 0 || errInternal.Code >= 512 {
			errInternal.Code = http.StatusInternalServerError
		}
	} else {
		log.Printf("endpoint.call: %d %s %s %s\n",
			http.StatusInternalServerError,
			req.Method, req.URL.Path, err)

		errInternal = errors.Internal(err)
	}

	res.WriteHeader(errInternal.Code)
	res.Header().Set(HeaderContentType, ContentTypeJSON)

	rsp, err := json.Marshal(errInternal)
	if err != nil {
		log.Println("endpoint.error: ", err)
		return
	}

	_, err = res.Write(rsp)
	if err != nil {
		log.Println("endpoint.error: ", err)
	}
}
