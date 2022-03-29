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
		ep             *Endpoint
		desc           string
		reqURL         string
		expContentType string
		expBody        string
		expError       string
		expStatusCode  int
	}{{
		desc: "With new endpoint",
		ep: &Endpoint{
			Method:       RequestMethodDelete,
			Path:         "/delete",
			ResponseType: ResponseTypePlain,
			Call:         cbPlain,
		},
	}, {
		desc: "With duplicate endpoint",
		ep: &Endpoint{
			Method:       RequestMethodDelete,
			Path:         "/delete",
			ResponseType: ResponseTypePlain,
			Call:         cbPlain,
		},
		expError: ErrEndpointAmbiguous.Error(),
	}, {
		desc:          "With unknown path",
		reqURL:        testServerUrl,
		expStatusCode: http.StatusNotFound,
	}, {
		desc:           "With known path and subtree root",
		reqURL:         testServerUrl + "/delete/",
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypePlain,
		expBody:        "map[]\nmap[]\n<nil>\n",
	}, {
		desc: "With response type none",
		ep: &Endpoint{
			Method:       RequestMethodDelete,
			Path:         "/delete/none",
			ResponseType: ResponseTypeNone,
			Call:         cbNone,
		},
		reqURL:        testServerUrl + "/delete/none?k=v",
		expStatusCode: http.StatusNoContent,
	}, {
		desc: "With response type binary",
		ep: &Endpoint{
			Method:       RequestMethodDelete,
			Path:         "/delete/bin",
			ResponseType: ResponseTypeBinary,
			Call:         cbPlain,
		},
		reqURL:         testServerUrl + "/delete/bin?k=v",
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypeBinary,
		expBody:        "map[k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc:           "With response type plain",
		reqURL:         testServerUrl + "/delete?k=v",
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypePlain,
		expBody:        "map[k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc: "With response type JSON",
		ep: &Endpoint{
			Method:       RequestMethodDelete,
			Path:         "/delete/json",
			ResponseType: ResponseTypeJSON,
			Call:         cbJSON,
		},
		reqURL:         testServerUrl + "/delete/json?k=v",
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypeJSON,
		expBody: `{
"form": "map[k:[v]]",
"multipartForm": "<nil>",
"body": ""
}`,
	}, {
		desc: "With ambigous path",
		ep: &Endpoint{
			Method:       RequestMethodDelete,
			Path:         "/delete/:id",
			ResponseType: ResponseTypePlain,
			Call:         cbPlain,
		},
		expError: ErrEndpointAmbiguous.Error(),
	}, {
		desc: "With key",
		ep: &Endpoint{
			Method:       RequestMethodDelete,
			Path:         "/delete/:id/x",
			ResponseType: ResponseTypePlain,
			Call:         cbPlain,
		},
		reqURL:         testServerUrl + "/delete/1/x?k=v",
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypePlain,
		expBody:        "map[id:[1] k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc:           "With duplicate key in query",
		reqURL:         testServerUrl + "/delete/1/x?id=v",
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypePlain,
		expBody:        "map[id:[1]]\nmap[]\n<nil>\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		err := testServer.RegisterEndpoint(c.ep)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		if len(c.reqURL) == 0 {
			continue
		}

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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)

		if c.expStatusCode != http.StatusOK {
			continue
		}

		test.Assert(t, "Body", c.expBody, string(body))

		gotContentType := res.Header.Get(HeaderContentType)

		test.Assert(t, "Content-Type", c.expContentType, gotContentType)
	}
}

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

	err := testServer.registerDelete(epEvaluate)
	if err != nil {
		t.Fatal(err)
	}

	testServer.RegisterEvaluator(testEvaluator)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
	}{{
		desc:          "With invalid evaluate",
		reqURL:        testServerUrl + "/evaluate",
		expStatusCode: http.StatusBadRequest,
	}, {
		desc:          "With valid evaluate",
		reqURL:        testServerUrl + "/evaluate?k=v",
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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
	}
}

func TestRegisterGet(t *testing.T) {
	testServer.evals = nil

	epGet := &Endpoint{
		Path:         "/get",
		ResponseType: ResponseTypePlain,
		Call:         cbPlain,
	}

	err := testServer.registerGet(epGet)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc          string
		reqURL        string
		expBody       string
		expStatusCode int
	}{{
		desc:          "With root path",
		reqURL:        testServerUrl,
		expStatusCode: http.StatusOK,
		expBody:       "<html><body>Hello, world!</body></html>\n",
	}, {
		desc:          "With known path",
		reqURL:        testServerUrl + "/index.js",
		expStatusCode: http.StatusOK,
		expBody:       "var a = \"Hello, world!\"\n",
	}, {
		desc:          "With known path and subtree root",
		reqURL:        testServerUrl + "/get/",
		expStatusCode: http.StatusOK,
		expBody:       "map[]\nmap[]\n<nil>\n",
	}, {
		desc:          "With known path",
		reqURL:        testServerUrl + "/get?k=v",
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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
		test.Assert(t, "Body", c.expBody, string(body))
	}
}

func TestRegisterHead(t *testing.T) {
	testServer.routeGets = nil

	epAPI := &Endpoint{
		Path:         "/api",
		ResponseType: ResponseTypeJSON,
		Call:         cbNone,
	}

	err := testServer.registerGet(epAPI)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc             string
		reqURL           string
		expBody          string
		expContentType   []string
		expContentLength []string
		expStatusCode    int
	}{{
		desc:             "With root path",
		reqURL:           testServerUrl + "/",
		expStatusCode:    http.StatusOK,
		expContentType:   []string{"text/html; charset=utf-8"},
		expContentLength: []string{"40"},
	}, {
		desc:           "With registered GET and subtree root",
		reqURL:         testServerUrl + "/api/",
		expStatusCode:  http.StatusOK,
		expContentType: []string{ContentTypeJSON},
	}, {
		desc:           "With registered GET",
		reqURL:         testServerUrl + "/api?k=v",
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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
		test.Assert(t, "Body", c.expBody, string(body))
		test.Assert(t, "Header.ContentType", c.expContentType, res.Header[HeaderContentType])
		test.Assert(t, "Header.ContentLength", c.expContentLength, res.Header[HeaderContentLength])
	}
}

func TestRegisterPatch(t *testing.T) {
	ep := &Endpoint{
		Path:         "/patch",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbPlain,
	}

	err := testServer.registerPatch(ep)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc          string
		reqURL        string
		expBody       string
		expStatusCode int
	}{{
		desc:          "With root path",
		reqURL:        testServerUrl + "/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PATCH and subtree root",
		reqURL:        testServerUrl + "/patch/",
		expStatusCode: http.StatusOK,
		expBody:       "map[]\nmap[]\n<nil>\n",
	}, {
		desc:          "With registered PATCH and query",
		reqURL:        testServerUrl + "/patch?k=v",
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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
		test.Assert(t, "Body", c.expBody, string(body))
	}
}

func TestRegisterPost(t *testing.T) {
	ep := &Endpoint{
		Path:         "/post",
		RequestType:  RequestTypeForm,
		ResponseType: ResponseTypePlain,
		Call:         cbPlain,
	}

	err := testServer.registerPost(ep)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc          string
		reqURL        string
		reqBody       string
		expBody       string
		expStatusCode int
	}{{
		desc:          "With root path",
		reqURL:        testServerUrl + "/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered POST and subtree root",
		reqURL:        testServerUrl + "/post/",
		expStatusCode: http.StatusOK,
		expBody:       "map[]\nmap[]\n<nil>\n",
	}, {
		desc:          "With registered POST and query",
		reqURL:        testServerUrl + "/post?k=v",
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

		req.Header.Set(HeaderContentType, ContentTypeForm)

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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
		test.Assert(t, "Body", c.expBody, string(body))
	}
}

func TestRegisterPut(t *testing.T) {
	ep := &Endpoint{
		Path:        "/put",
		RequestType: RequestTypeForm,
		Call:        cbPlain,
	}

	err := testServer.registerPut(ep)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc          string
		reqURL        string
		expBody       string
		expStatusCode int
	}{{
		desc:          "With root path",
		reqURL:        testServerUrl + "/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PUT and subtree root",
		reqURL:        testServerUrl + "/put/",
		expStatusCode: http.StatusNoContent,
	}, {
		desc:          "With registered PUT and query",
		reqURL:        testServerUrl + "/put?k=v",
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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
		test.Assert(t, "Body", c.expBody, string(body))
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

	err := testServer.registerDelete(epDelete)
	if err != nil {
		t.Fatal(err)
	}

	err = testServer.registerPatch(epPatch)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc          string
		reqURL        string
		expAllow      string
		expStatusCode int
	}{{
		desc:          "With root path",
		reqURL:        testServerUrl + "/",
		expStatusCode: http.StatusOK,
		expAllow:      "GET, HEAD, OPTIONS",
	}, {
		desc:          "With registered PATCH and subtree root",
		reqURL:        testServerUrl + "/options/",
		expStatusCode: http.StatusOK,
		expAllow:      "DELETE, OPTIONS, PATCH",
	}, {
		desc:          "With registered PATCH and query",
		reqURL:        testServerUrl + "/options?k=v",
		expStatusCode: http.StatusOK,
		expAllow:      "DELETE, OPTIONS, PATCH",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		req, err := http.NewRequest(http.MethodOptions, c.reqURL, nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		gotAllow := res.Header.Get("Allow")

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
		test.Assert(t, "Allow", c.expAllow, gotAllow)
	}
}

func TestStatusError(t *testing.T) {
	cbError := func(epr *EndpointRequest) (
		[]byte, error,
	) {
		return nil, &errors.E{
			Code:    http.StatusLengthRequired,
			Message: "Length required",
		}
	}

	cbNoCode := func(epr *EndpointRequest) (
		[]byte, error,
	) {
		return nil, errors.Internal(nil)
	}

	cbCustomErr := func(epr *EndpointRequest) (
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
	err := testServer.registerPost(epErrNoBody)
	if err != nil {
		t.Fatal(err)
	}

	epErrBinary := &Endpoint{
		Path:         "/error/binary",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypeBinary,
		Call:         cbError,
	}
	err = testServer.registerPost(epErrBinary)
	if err != nil {
		t.Fatal(err)
	}

	epErrJSON := &Endpoint{
		Path:         "/error/json",
		RequestType:  RequestTypeJSON,
		ResponseType: ResponseTypeJSON,
		Call:         cbError,
	}
	err = testServer.registerPost(epErrJSON)
	if err != nil {
		t.Fatal(err)
	}

	epErrPlain := &Endpoint{
		Path:         "/error/plain",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbError,
	}
	err = testServer.registerPost(epErrPlain)
	if err != nil {
		t.Fatal(err)
	}

	epErrNoCode := &Endpoint{
		Path:         "/error/no-code",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbNoCode,
	}
	err = testServer.registerPost(epErrNoCode)
	if err != nil {
		t.Fatal(err)
	}

	epErrCustom := &Endpoint{
		Path:         "/error/custom",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbCustomErr,
	}
	err = testServer.registerPost(epErrCustom)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc          string
		reqURL        string
		expBody       string
		expStatusCode int
	}{{
		desc:          "With registered error no body",
		reqURL:        testServerUrl + "/error/no-body?k=v",
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"code":411,"message":"Length required"}`,
	}, {
		desc:          "With registered error binary",
		reqURL:        testServerUrl + "/error/binary?k=v",
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"code":411,"message":"Length required"}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        testServerUrl + "/error/plain?k=v",
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"code":411,"message":"Length required"}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        testServerUrl + "/error/json?k=v",
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"code":411,"message":"Length required"}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        testServerUrl + "/error/no-code?k=v",
		expStatusCode: http.StatusInternalServerError,
		expBody:       `{"code":500,"message":"internal server error","name":"ERR_INTERNAL"}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        testServerUrl + "/error/custom?k=v",
		expStatusCode: http.StatusInternalServerError,
		expBody:       `{"code":500,"message":"internal server error","name":"ERR_INTERNAL"}`,
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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
		test.Assert(t, "Body", c.expBody, string(body))
	}
}

// TestServer_HandleFS_Auth test GET on memfs with authorization.
func TestServer_HandleFS_Auth(t *testing.T) {
	type testCase struct {
		cookieSid     *http.Cookie
		desc          string
		reqPath       string
		expStatusCode int
	}

	var (
		c   testCase
		req *http.Request
		res *http.Response
		err error
	)

	cases := []testCase{{
		desc:          "With public path",
		reqPath:       "/index.html",
		expStatusCode: http.StatusOK,
	}, {
		desc:          "With /auth.txt",
		reqPath:       "/auth.txt",
		expStatusCode: http.StatusOK,
	}, {
		desc:          "With /auth path no cookie",
		reqPath:       "/auth",
		expStatusCode: http.StatusUnauthorized,
	}, {
		desc:    "With /auth path and cookie",
		reqPath: "/auth",
		cookieSid: &http.Cookie{
			Name:  "sid",
			Value: "authz",
		},
		expStatusCode: http.StatusOK,
	}, {
		desc:    "With invalid /auth path and cookie",
		reqPath: "/auth/notexist",
		cookieSid: &http.Cookie{
			Name:  "sid",
			Value: "authz",
		},
		expStatusCode: http.StatusNotFound,
	}, {
		desc:    "With /auth/sub path and cookie",
		reqPath: "/auth/sub",
		cookieSid: &http.Cookie{
			Name:  "sid",
			Value: "authz",
		},
		expStatusCode: http.StatusOK,
	}}

	for _, c = range cases {
		req, err = http.NewRequest(http.MethodGet, testServerUrl+c.reqPath, nil)
		if err != nil {
			t.Fatalf("%s: %s", c.desc, err)
		}

		if c.cookieSid != nil {
			req.AddCookie(c.cookieSid)
		}

		res, err = client.Do(req)
		if err != nil {
			t.Fatalf("%s: %s", c.desc, err)
		}

		test.Assert(t, c.desc, c.expStatusCode, res.StatusCode)
	}
}
