// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	liberrors "git.sr.ht/~shulhan/pakakeh.go/lib/errors"
	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
)

// HTTPHandleFormat define the HTTP handler for formating Go code.
func HTTPHandleFormat(httpresw http.ResponseWriter, httpreq *http.Request) {
	var (
		logp = `HTTPHandleFormat`
		resp = libhttp.EndpointResponse{}

		req     Request
		rawbody []byte
		err     error
	)

	var contentType = httpreq.Header.Get(libhttp.HeaderContentType)
	if contentType != libhttp.ContentTypeJSON {
		resp.Code = http.StatusUnsupportedMediaType
		resp.Name = `ERR_CONTENT_TYPE`
		goto out
	}

	rawbody, err = io.ReadAll(httpreq.Body)
	if err != nil {
		resp.Code = http.StatusInternalServerError
		resp.Name = `ERR_INTERNAL`
		resp.Message = err.Error()
		goto out
	}

	err = json.Unmarshal(rawbody, &req)
	if err != nil {
		resp.Code = http.StatusBadRequest
		resp.Name = `ERR_BAD_REQUEST`
		resp.Message = err.Error()
		goto out
	}

	rawbody, err = Format(req)
	if err != nil {
		resp.Code = http.StatusUnprocessableEntity
		resp.Name = `ERR_CODE`
		resp.Message = err.Error()
		goto out
	}

	resp.Code = http.StatusOK
	resp.Data = string(rawbody)
out:
	rawbody, err = json.Marshal(resp)
	if err != nil {
		log.Printf(`%s: %s`, logp, err)
		resp.Code = http.StatusInternalServerError
	}
	httpresw.Header().Set(libhttp.HeaderContentType,
		libhttp.ContentTypeJSON)
	httpresw.WriteHeader(resp.Code)
	httpresw.Write(rawbody)
}

// HTTPHandleRun define the HTTP handler for running Go code.
// Each client is identified by unique cookie, so if two Run requests come
// from the same client, the previous Run will be cancelled.
func HTTPHandleRun(httpresw http.ResponseWriter, httpreq *http.Request) {
	var (
		logp = `HTTPHandleRun`

		req  *Request
		resp *libhttp.EndpointResponse
		rawb []byte
		err  error
	)

	req, resp = readRequest(httpreq)
	if resp != nil {
		goto out
	}

	rawb, err = Run(req)
	if err != nil {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: err.Error(),
				Name:    `ERR_INTERNAL`,
				Code:    http.StatusInternalServerError,
			},
		}
		goto out
	}

	http.SetCookie(httpresw, req.cookieSid)
	resp = &libhttp.EndpointResponse{}
	resp.Code = http.StatusOK
	resp.Data = string(rawb)
out:
	rawb, err = json.Marshal(resp)
	if err != nil {
		log.Printf(`%s: %s`, logp, err)
		resp.Code = http.StatusInternalServerError
	}
	httpresw.Header().Set(libhttp.HeaderContentType,
		libhttp.ContentTypeJSON)
	httpresw.WriteHeader(resp.Code)
	httpresw.Write(rawb)
}

func readRequest(httpreq *http.Request) (
	req *Request,
	resp *libhttp.EndpointResponse,
) {
	var contentType = httpreq.Header.Get(libhttp.HeaderContentType)
	if contentType != libhttp.ContentTypeJSON {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: `invalid content type`,
				Name:    `ERR_CONTENT_TYPE`,
				Code:    http.StatusUnsupportedMediaType,
			},
		}
		return nil, resp
	}

	var (
		rawbody []byte
		err     error
	)

	rawbody, err = io.ReadAll(httpreq.Body)
	if err != nil {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: err.Error(),
				Name:    `ERR_INTERNAL`,
				Code:    http.StatusInternalServerError,
			},
		}
		return nil, resp
	}

	err = json.Unmarshal(rawbody, &req)
	if err != nil {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: err.Error(),
				Name:    `ERR_BAD_REQUEST`,
				Code:    http.StatusBadRequest,
			},
		}
		return nil, resp
	}

	req.cookieSid, err = httpreq.Cookie(cookieNameSid)
	if err != nil {
		// Ignore the error if cookie is not exist, we wiil generate
		// one later.
	}
	return req, nil
}

// HTTPHandleTest define the HTTP handler for testing Go code.
// Each client is identified by unique cookie, so if two Run requests come
// from the same client, the previous Test will be cancelled.
func HTTPHandleTest(httpresw http.ResponseWriter, httpreq *http.Request) {
	var (
		logp = `HTTPHandleTest`

		treq *Request
		resp *libhttp.EndpointResponse
		rawb []byte
		err  error
	)

	treq, resp = readRequest(httpreq)
	if resp != nil {
		goto out
	}

	rawb, err = Test(treq)
	if err != nil {
		resp = &libhttp.EndpointResponse{
			E: liberrors.E{
				Message: err.Error(),
				Name:    `ERR_INTERNAL`,
				Code:    http.StatusInternalServerError,
			},
		}
		goto out
	}

	http.SetCookie(httpresw, treq.cookieSid)
	resp = &libhttp.EndpointResponse{}
	resp.Code = http.StatusOK
	resp.Data = string(rawb)
out:
	rawb, err = json.Marshal(resp)
	if err != nil {
		log.Printf(`%s: %s`, logp, err)
		resp.Code = http.StatusInternalServerError
	}
	httpresw.Header().Set(libhttp.HeaderContentType,
		libhttp.ContentTypeJSON)
	httpresw.WriteHeader(resp.Code)
	httpresw.Write(rawb)
}
