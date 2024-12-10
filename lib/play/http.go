// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"encoding/json"
	"log"
	"net/http"

	liberrors "git.sr.ht/~shulhan/pakakeh.go/lib/errors"
	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
)

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
	httpresw.Header().Set(libhttp.HeaderContentType, libhttp.ContentTypeJSON)
	httpresw.WriteHeader(resp.Code)
	httpresw.Write(rawb)
}
