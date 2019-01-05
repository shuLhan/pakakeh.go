// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/errors"
	"github.com/shuLhan/share/lib/strings"
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

func (ep *Endpoint) call(res http.ResponseWriter, req *http.Request,
	evaluators []Evaluator,
) {
	var (
		e       error
		reqBody []byte
	)

	switch ep.RequestType {
	case RequestTypeForm:
		e = req.ParseForm()

	case RequestTypeQuery:
		e = req.ParseForm()

	case RequestTypeMultipartForm:
		e = req.ParseMultipartForm(0)

	case RequestTypeJSON:
		e = req.ParseForm()
		if e != nil {
			res.WriteHeader(http.StatusBadRequest)
			return
		}
		reqBody, e = ioutil.ReadAll(req.Body)
	}
	if e != nil {
		log.Printf("endpoint.call: %d %s %s %s\n",
			http.StatusBadRequest, req.Method, req.URL.Path, e)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if debug.Value > 0 {
		log.Printf("> request body: %s\n", reqBody)
	}

	for _, eval := range evaluators {
		e = eval(req, reqBody)
		if e != nil {
			ep.error(res, e)
			return
		}
	}

	if ep.Eval != nil {
		e = ep.Eval(req, reqBody)
		if e != nil {
			ep.error(res, e)
			return
		}
	}

	rspb, e := ep.Call(req, reqBody)
	if e != nil {
		log.Printf("endpoint.call: %d %s %s %s\n",
			http.StatusInternalServerError,
			req.Method, req.URL.Path, e)
		ep.error(res, e)
		return
	}

	switch ep.ResponseType {
	case ResponseTypeNone:
		res.WriteHeader(http.StatusNoContent)
		return
	case ResponseTypeBinary:
		res.Header().Set(ContentType, ContentTypeBinary)
	case ResponseTypeJSON:
		res.Header().Set(ContentType, ContentTypeJSON)
	case ResponseTypePlain:
		res.Header().Set(ContentType, ContentTypePlain)
	}

	res.WriteHeader(http.StatusOK)

	_, e = res.Write(rspb)
	if e != nil {
		log.Printf("endpoint.call: %s %s %s\n", req.Method, req.URL.Path, e)
	}
}

func (ep *Endpoint) error(res http.ResponseWriter, e error) {
	se, ok := e.(*errors.E)
	if !ok {
		se = &errors.E{
			Code:    http.StatusInternalServerError,
			Message: e.Error(),
		}
	} else {
		if se.Code == 0 {
			se.Code = http.StatusInternalServerError
		}
	}

	res.WriteHeader(se.Code)
	res.Header().Set(ContentType, ContentTypeJSON)

	rsp := fmt.Sprintf(`{"code":%d,"message":"%s"}`, se.Code,
		strings.JSONEscape(se.Message))

	_, e = res.Write([]byte(rsp))
	if e != nil {
		log.Println("endpoint.error: ", e)
	}
}
