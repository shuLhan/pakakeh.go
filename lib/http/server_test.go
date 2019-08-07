// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/shuLhan/share/lib/errors"
	"github.com/shuLhan/share/lib/test"
)

func TestRegisterDelete(t *testing.T) {
	cases := []struct {
		desc          string
		reqURL        string
		ep            *Endpoint
		expStatusCode int
		expBody       string
	}{{
		desc:   "With unknown path",
		reqURL: "http://127.0.0.1:8080/",
		ep: &Endpoint{
			Path:         "/delete",
			ResponseType: ResponseTypePlain,
			Call:         cbPlain,
		},
		expStatusCode: http.StatusNotFound,
	}, {
		desc:   "With known path and subtree root",
		reqURL: "http://127.0.0.1:8080/delete/",
		ep: &Endpoint{
			Path:         "/delete",
			ResponseType: ResponseTypePlain,
			Call:         cbPlain,
		},
		expStatusCode: http.StatusNotFound,
	}, {
		desc:   "With response type none",
		reqURL: "http://127.0.0.1:8080/delete?k=v",
		ep: &Endpoint{
			Path:         "/delete",
			ResponseType: ResponseTypeNone,
			Call:         cbNone,
		},
		expStatusCode: http.StatusNoContent,
	}, {
		desc:   "With response type binary",
		reqURL: "http://127.0.0.1:8080/delete?k=v",
		ep: &Endpoint{
			Path:         "/delete",
			ResponseType: ResponseTypeBinary,
			Call:         cbPlain,
		},
		expStatusCode: http.StatusOK,
		expBody:       "map[k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc:   "With response type plain",
		reqURL: "http://127.0.0.1:8080/delete?k=v",
		ep: &Endpoint{
			Path:         "/delete",
			ResponseType: ResponseTypePlain,
			Call:         cbPlain,
		},
		expStatusCode: http.StatusOK,
		expBody:       "map[k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc:   "With response type JSON",
		reqURL: "http://127.0.0.1:8080/delete?k=v",
		ep: &Endpoint{
			Path:         "/delete",
			ResponseType: ResponseTypeJSON,
			Call:         cbJSON,
		},
		expStatusCode: http.StatusOK,
		expBody: `{
"form": "map[k:[v]]",
"multipartForm": "<nil>",
"body": ""
}`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		testServer.RegisterDelete(c.ep)

		req, e := http.NewRequest(http.MethodDelete, c.reqURL, nil)
		if e != nil {
			t.Fatal(e)
		}

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}

		e = res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		if c.expStatusCode != http.StatusOK {
			continue
		}

		test.Assert(t, "Body", c.expBody, string(body), true)

		var expContentType string
		gotContentType := res.Header.Get(ContentType)

		switch c.ep.ResponseType {
		case ResponseTypeBinary:
			expContentType = ContentTypeBinary
		case ResponseTypeJSON:
			expContentType = ContentTypeJSON
		default:
			expContentType = ContentTypePlain
		}

		test.Assert(t, "Content-Type", expContentType, gotContentType,
			true)
	}
}

//nolint:gochecknoglobals
var testEvaluator = func(req *http.Request, reqBody []byte) error {
	k := req.Form.Get("k")

	if len(k) == 0 {
		return &errors.E{
			Code:    http.StatusBadRequest,
			Message: "Missing input value for k",
		}
	}

	return nil
}

func TestRegisterEvaluator(t *testing.T) {
	epEvaluate := &Endpoint{
		Path:         "/evaluate",
		ResponseType: ResponseTypeJSON,
		Call:         cbPlain,
	}

	testServer.RegisterDelete(epEvaluate)
	testServer.RegisterEvaluator(testEvaluator)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
	}{{
		desc:          "With invalid evaluate",
		reqURL:        "http://127.0.0.1:8080/evaluate",
		expStatusCode: http.StatusBadRequest,
	}, {
		desc:          "With valid evaluate",
		reqURL:        "http://127.0.0.1:8080/evaluate?k=v",
		expStatusCode: http.StatusOK,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		req, e := http.NewRequest(http.MethodDelete, c.reqURL, nil)
		if e != nil {
			t.Fatal(e)
		}

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		_, e = ioutil.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}

		e = res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
	}
}

func TestRegisterGet(t *testing.T) {
	epGet := &Endpoint{
		Path:         "/get",
		ResponseType: ResponseTypePlain,
		Call:         cbPlain,
	}
	testServer.RegisterGet(epGet)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expBody       string
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusOK,
		expBody:       "<html><body>Hello, world!</body></html>\n",
	}, {
		desc:          "With known path",
		reqURL:        "http://127.0.0.1:8080/index.js",
		expStatusCode: http.StatusOK,
		expBody:       "var a = \"Hello, world!\"\n",
	}, {
		desc:          "With known path and subtree root",
		reqURL:        "http://127.0.0.1:8080/get/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With known path",
		reqURL:        "http://127.0.0.1:8080/get?k=v",
		expStatusCode: http.StatusOK,
		expBody:       "map[k:[v]]\nmap[]\n<nil>\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		req, e := http.NewRequest(http.MethodGet, c.reqURL, nil)
		if e != nil {
			t.Fatal(e)
		}

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}

		e = res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		test.Assert(t, "Body", c.expBody, string(body), true)
	}
}

func TestRegisterHead(t *testing.T) {
	epAPI := &Endpoint{
		Path:         "/api",
		ResponseType: ResponseTypeJSON,
		Call:         cbNone,
	}
	testServer.RegisterGet(epAPI)

	cases := []struct {
		desc             string
		reqURL           string
		expStatusCode    int
		expBody          string
		expContentType   []string
		expContentLength []string
	}{{
		desc:             "With root path",
		reqURL:           "http://127.0.0.1:8080/",
		expStatusCode:    http.StatusOK,
		expContentType:   []string{"text/html; charset=utf-8"},
		expContentLength: []string{"40"},
	}, {
		desc:          "With registered GET and subtree root",
		reqURL:        "http://127.0.0.1:8080/api/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:           "With registered GET",
		reqURL:         "http://127.0.0.1:8080/api?k=v",
		expStatusCode:  http.StatusOK,
		expContentType: []string{ContentTypeJSON},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		req, e := http.NewRequest(http.MethodHead, c.reqURL, nil)
		if e != nil {
			t.Fatal(e)
		}

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}

		e = res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		test.Assert(t, "Body", c.expBody, string(body), true)
		test.Assert(t, "Header.ContentType", c.expContentType,
			res.Header[ContentType], true)
		test.Assert(t, "Header.ContentLength", c.expContentLength,
			res.Header[ContentLength], true)
	}
}

func TestRegisterPatch(t *testing.T) {
	ep := &Endpoint{
		Path:         "/patch",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbPlain,
	}
	testServer.RegisterPatch(ep)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expBody       string
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PATCH and subtree root",
		reqURL:        "http://127.0.0.1:8080/patch/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PATCH and query",
		reqURL:        "http://127.0.0.1:8080/patch?k=v",
		expStatusCode: http.StatusOK,
		expBody:       "map[k:[v]]\nmap[]\n<nil>\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		req, e := http.NewRequest(http.MethodPatch, c.reqURL, nil)
		if e != nil {
			t.Fatal(e)
		}

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}

		e = res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		test.Assert(t, "Body", c.expBody, string(body), true)
	}
}

func TestRegisterPost(t *testing.T) {
	ep := &Endpoint{
		Path:         "/post",
		RequestType:  RequestTypeForm,
		ResponseType: ResponseTypePlain,
		Call:         cbPlain,
	}

	testServer.RegisterPost(ep)

	cases := []struct {
		desc          string
		reqURL        string
		reqBody       string
		expStatusCode int
		expBody       string
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered POST and subtree root",
		reqURL:        "http://127.0.0.1:8080/post/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered POST and query",
		reqURL:        "http://127.0.0.1:8080/post?k=v",
		reqBody:       "k=vv",
		expStatusCode: http.StatusOK,
		expBody: `map[k:[vv v]]
map[k:[vv]]
<nil>
k=vv`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var buf bytes.Buffer
		_, _ = buf.WriteString(c.reqBody)

		req, e := http.NewRequest(http.MethodPost, c.reqURL, &buf)
		if e != nil {
			t.Fatal(e)
		}

		req.Header.Set(ContentType, ContentTypeForm)

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}

		e = res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		test.Assert(t, "Body", c.expBody, string(body), true)
	}
}

func TestRegisterPut(t *testing.T) {
	ep := &Endpoint{
		Path:        "/put",
		RequestType: RequestTypeForm,
		Call:        cbPlain,
	}

	testServer.RegisterPut(ep)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expBody       string
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PUT and subtree root",
		reqURL:        "http://127.0.0.1:8080/put/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PUT and query",
		reqURL:        "http://127.0.0.1:8080/put?k=v",
		expStatusCode: http.StatusNoContent,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		req, e := http.NewRequest(http.MethodPut, c.reqURL, nil)
		if e != nil {
			t.Fatal(e)
		}

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}

		e = res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		test.Assert(t, "Body", c.expBody, string(body), true)
	}
}

func TestServeHTTPOptions(t *testing.T) {
	epDelete := &Endpoint{
		Path:         "/options",
		ResponseType: ResponseTypePlain,
		Call:         cbPlain,
	}
	epPatch := &Endpoint{
		Path:         "/options",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbPlain,
	}

	testServer.RegisterDelete(epDelete)
	testServer.RegisterPatch(epPatch)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expAllow      string
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusOK,
		expAllow:      "GET, HEAD, OPTIONS",
	}, {
		desc:          "With registered PATCH and subtree root",
		reqURL:        "http://127.0.0.1:8080/options/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PATCH and query",
		reqURL:        "http://127.0.0.1:8080/options?k=v",
		expStatusCode: http.StatusOK,
		expAllow:      "DELETE, OPTIONS, PATCH",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		req, err := http.NewRequest(http.MethodOptions, c.reqURL, nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := client.Do(req) //nolint:bodyclose
		if err != nil {
			t.Fatal(err)
		}

		gotAllow := res.Header.Get("Allow")

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		test.Assert(t, "Allow", c.expAllow, gotAllow, true)
	}
}

func TestStatusError(t *testing.T) {
	cbError := func(res http.ResponseWriter, req *http.Request, reqBody []byte) (
		[]byte, error,
	) {
		return nil, &errors.E{
			Code:    http.StatusLengthRequired,
			Message: "Length required",
		}
	}

	cbNoCode := func(res http.ResponseWriter, req *http.Request, reqBody []byte) (
		[]byte, error,
	) {
		return nil, &errors.E{
			Message: "Internal server error",
		}
	}

	cbCustomErr := func(res http.ResponseWriter, req *http.Request, reqBody []byte) (
		[]byte, error,
	) {
		return nil, fmt.Errorf("Custom error")
	}

	epErrNoBody := &Endpoint{
		Path:         "/error/no-body",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypeNone,
		Call:         cbError,
	}
	testServer.RegisterPost(epErrNoBody)

	epErrBinary := &Endpoint{
		Path:         "/error/binary",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypeBinary,
		Call:         cbError,
	}
	testServer.RegisterPost(epErrBinary)

	epErrJSON := &Endpoint{
		Path:         "/error/json",
		RequestType:  RequestTypeJSON,
		ResponseType: ResponseTypeJSON,
		Call:         cbError,
	}
	testServer.RegisterPost(epErrJSON)

	epErrPlain := &Endpoint{
		Path:         "/error/plain",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbError,
	}
	testServer.RegisterPost(epErrPlain)

	epErrNoCode := &Endpoint{
		Path:         "/error/no-code",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbNoCode,
	}
	testServer.RegisterPost(epErrNoCode)

	epErrCustom := &Endpoint{
		Path:         "/error/custom",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbCustomErr,
	}
	testServer.RegisterPost(epErrCustom)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expBody       string
	}{{
		desc:          "With registered error no body",
		reqURL:        "http://127.0.0.1:8080/error/no-body?k=v",
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"code":411,"message":"Length required"}`,
	}, {
		desc:          "With registered error binary",
		reqURL:        "http://127.0.0.1:8080/error/binary?k=v",
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"code":411,"message":"Length required"}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        "http://127.0.0.1:8080/error/plain?k=v",
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"code":411,"message":"Length required"}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        "http://127.0.0.1:8080/error/json?k=v",
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"code":411,"message":"Length required"}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        "http://127.0.0.1:8080/error/no-code?k=v",
		expStatusCode: http.StatusInternalServerError,
		expBody:       `{"code":500,"message":"Internal server error"}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        "http://127.0.0.1:8080/error/custom?k=v",
		expStatusCode: http.StatusInternalServerError,
		expBody:       `{"code":500,"message":"Custom error"}`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		req, e := http.NewRequest(http.MethodPost, c.reqURL, nil)
		if e != nil {
			t.Fatal(e)
		}

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		body, e := ioutil.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}

		e = res.Body.Close()
		if e != nil {
			t.Fatal(e)
		}

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		test.Assert(t, "Body", c.expBody, string(body), true)
	}
}
