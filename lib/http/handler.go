// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/shuLhan/share/lib/debug"
)

type handler struct {
	reqType RequestType
	resType ResponseType
	cb      Callback
}

func (h *handler) call(res http.ResponseWriter, req *http.Request) {
	var (
		e       error
		reqBody []byte
	)

	switch h.reqType {
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
		log.Printf("handler.call: %d %s %s %s\n",
			http.StatusBadRequest, req.Method, req.URL.Path, e)
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	if debug.Value > 0 {
		log.Printf("> request body: %s\n", reqBody)
	}

	rspb, e := h.cb(req, reqBody)
	if e != nil {
		log.Printf("handler.call: %d %s %s %s\n",
			http.StatusInternalServerError,
			req.Method, req.URL.Path, e)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch h.resType {
	case ResponseTypeNone:
		res.WriteHeader(http.StatusNoContent)
		return
	case ResponseTypeBinary:
		res.Header().Set(contentType, contentTypeBinary)
	case ResponseTypeJSON:
		res.Header().Set(contentType, contentTypeJSON)
	case ResponseTypePlain:
		res.Header().Set(contentType, contentTypePlain)
	}

	res.WriteHeader(http.StatusOK)

	_, e = res.Write(rspb)
	if e != nil {
		log.Printf("handler.call: %s %s %s\n", req.Method, req.URL.Path, e)
	}
}
