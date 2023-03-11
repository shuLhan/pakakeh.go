// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	liberrors "github.com/shuLhan/share/lib/errors"
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
			if !errors.Is(ErrEndpointAmbiguous, err) {
				test.Assert(t, "error", c.expError, err.Error())
			}
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

		body, e := io.ReadAll(res.Body)
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
		return &liberrors.E{
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
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
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

		_, e = io.ReadAll(res.Body)
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
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
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

		body, e := io.ReadAll(res.Body)
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
	epAPI := &Endpoint{
		Path:         "/api",
		ResponseType: ResponseTypeJSON,
		Call:         cbNone,
	}

	err := testServer.registerGet(epAPI)
	if err != nil {
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
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

		body, e := io.ReadAll(res.Body)
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
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
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

		body, e := io.ReadAll(res.Body)
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
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
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

		body, e := io.ReadAll(res.Body)
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
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
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

		body, e := io.ReadAll(res.Body)
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
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
	}

	err = testServer.registerPatch(epPatch)
	if err != nil {
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
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
	var (
		cbError = func(epr *EndpointRequest) ([]byte, error) {
			return nil, &liberrors.E{
				Code:    http.StatusLengthRequired,
				Message: `Length required`,
			}
		}
		cbNoCode = func(epr *EndpointRequest) ([]byte, error) {
			return nil, liberrors.Internal(nil)
		}
		cbCustomErr = func(epr *EndpointRequest) ([]byte, error) {
			return nil, fmt.Errorf("Custom error")
		}

		err error
	)

	epErrNoBody := &Endpoint{
		Path:         "/error/no-body",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypeNone,
		Call:         cbError,
	}
	err = testServer.registerPost(epErrNoBody)
	if err != nil {
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
	}

	epErrBinary := &Endpoint{
		Path:         "/error/binary",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypeBinary,
		Call:         cbError,
	}
	err = testServer.registerPost(epErrBinary)
	if err != nil {
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
	}

	epErrJSON := &Endpoint{
		Path:         "/error/json",
		RequestType:  RequestTypeJSON,
		ResponseType: ResponseTypeJSON,
		Call:         cbError,
	}
	err = testServer.registerPost(epErrJSON)
	if err != nil {
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
	}

	epErrPlain := &Endpoint{
		Path:         "/error/plain",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbError,
	}
	err = testServer.registerPost(epErrPlain)
	if err != nil {
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
	}

	epErrNoCode := &Endpoint{
		Path:         "/error/no-code",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbNoCode,
	}
	err = testServer.registerPost(epErrNoCode)
	if err != nil {
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
	}

	epErrCustom := &Endpoint{
		Path:         "/error/custom",
		RequestType:  RequestTypeQuery,
		ResponseType: ResponseTypePlain,
		Call:         cbCustomErr,
	}
	err = testServer.registerPost(epErrCustom)
	if err != nil {
		if !errors.Is(ErrEndpointAmbiguous, err) {
			t.Fatal(err)
		}
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

		body, e := io.ReadAll(res.Body)
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

// TestServer_Options_HandleFS test GET on memfs with authorization.
func TestServer_Options_HandleFS(t *testing.T) {
	type testCase struct {
		cookieSid     *http.Cookie
		desc          string
		reqPath       string
		expResBody    string
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
		expResBody:    "<html><body>Hello, world!</body></html>\n",
	}, {
		desc:          "With /auth.txt",
		reqPath:       "/auth.txt",
		expStatusCode: http.StatusOK,
		expResBody:    "Hello, auth.txt!\n",
	}, {
		desc:          "With /auth path no cookie, redirected to /",
		reqPath:       "/auth",
		expStatusCode: http.StatusOK,
		expResBody:    "<html><body>Hello, world!</body></html>\n",
	}, {
		desc:    "With /auth path and cookie",
		reqPath: "/auth",
		cookieSid: &http.Cookie{
			Name:  "sid",
			Value: "authz",
		},
		expStatusCode: http.StatusOK,
		expResBody:    "<html><body>Hello, authorized world!</body></html>\n",
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
		expResBody:    "<html><body>Hello, /auth/sub!</body></html>\n",
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

		gotBody, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("%s: %s", c.desc, err)
		}

		test.Assert(t, "response body", c.expResBody, string(gotBody))
	}
}

func TestServer_handleRange(t *testing.T) {
	var (
		clOpts = &ClientOptions{
			ServerUrl: testServerUrl,
		}
		cl          = NewClient(clOpts)
		skipHeaders = []string{HeaderDate, HeaderETag}

		listTestData []*test.Data
		tdata        *test.Data
		httpRes      *http.Response
		resBody      []byte
		err          error
	)

	listTestData, err = test.LoadDataDir(`testdata/server/range/`)
	if err != nil {
		t.Fatal(err)
	}

	for _, tdata = range listTestData {
		t.Log(tdata.Name)

		var (
			header      = http.Header{}
			headerRange = tdata.Input[`header_range`]
		)

		header.Set(HeaderRange, string(headerRange))

		httpRes, resBody, err = cl.Get(`/index.html`, header, nil)
		if err != nil {
			t.Fatal(err)
		}

		var (
			tag = `http_headers`
			exp = tdata.Output[tag]
			got = dumpHttpResponse(httpRes, skipHeaders)
		)
		test.Assert(t, tag, string(exp), got)

		tag = `http_body`
		exp = tdata.Output[tag]

		// Replace the response body CRLF with LF.
		resBody = bytes.ReplaceAll(resBody, []byte("\r\n"), []byte("\n"))

		test.Assert(t, tag, string(exp), string(resBody))

		tag = `all_body`
		exp = tdata.Output[tag]
		got = dumpMultipartBody(httpRes)

		test.Assert(t, tag, string(exp), got)
	}
}

func TestServer_handleRange_HEAD(t *testing.T) {
	var (
		clOpts = &ClientOptions{
			ServerUrl: testServerUrl,
		}
		cl = NewClient(clOpts)

		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/server/head_range_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var httpRes *http.Response

	httpRes, err = cl.Client.Head(testServerUrl + `/index.html`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		skipHeaders = []string{HeaderDate, HeaderETag}
		got         = dumpHttpResponse(httpRes, skipHeaders)
		tag         = `http_headers`
		exp         = tdata.Output[tag]
	)
	test.Assert(t, tag, string(exp), got)
}
