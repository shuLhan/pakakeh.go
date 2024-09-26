// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	liberrors "git.sr.ht/~shulhan/pakakeh.go/lib/errors"
	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
	libhttptest "git.sr.ht/~shulhan/pakakeh.go/lib/test/httptest"
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
		expError: `RegisterEndpoint: ` + ErrEndpointAmbiguous.Error(),
	}, {
		desc:          "With unknown path",
		reqURL:        testServerURL,
		expStatusCode: http.StatusNotFound,
	}, {
		desc:           "With known path and subtree root",
		reqURL:         testServerURL + `/delete/`,
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
		reqURL:        testServerURL + `/delete/none?k=v`,
		expStatusCode: http.StatusOK,
	}, {
		desc: "With response type binary",
		ep: &Endpoint{
			Method:       RequestMethodDelete,
			Path:         "/delete/bin",
			ResponseType: ResponseTypeBinary,
			Call:         cbPlain,
		},
		reqURL:         testServerURL + `/delete/bin?k=v`,
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypeBinary,
		expBody:        "map[k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc:           "With response type plain",
		reqURL:         testServerURL + `/delete?k=v`,
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
		reqURL:         testServerURL + `/delete/json?k=v`,
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
		reqURL:         testServerURL + `/delete/1/x?k=v`,
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypePlain,
		expBody:        "map[id:[1] k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc:           "With duplicate key in query",
		reqURL:         testServerURL + `/delete/1/x?id=v`,
		expStatusCode:  http.StatusOK,
		expContentType: ContentTypePlain,
		expBody:        "map[id:[1]]\nmap[]\n<nil>\n",
	}}

	var err error
	for _, c := range cases {
		t.Log(c.desc)

		if c.ep != nil {
			err = testServer.RegisterEndpoint(*c.ep)
			if err != nil {
				test.Assert(t, `error`, c.expError, err.Error())
				continue
			}
		}

		if len(c.reqURL) == 0 {
			continue
		}

		var (
			ctx = context.Background()
			req *http.Request
		)

		req, err = http.NewRequestWithContext(ctx, http.MethodDelete, c.reqURL, nil)
		if err != nil {
			t.Fatal(err)
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

var testEvaluator = func(req *http.Request, _ []byte) error {
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
		reqURL:        testServerURL + `/evaluate`,
		expStatusCode: http.StatusBadRequest,
	}, {
		desc:          "With valid evaluate",
		reqURL:        testServerURL + `/evaluate?k=v`,
		expStatusCode: http.StatusOK,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var (
			ctx = context.Background()
			req *http.Request
			err error
		)

		req, err = http.NewRequestWithContext(ctx, http.MethodDelete, c.reqURL, nil)
		if err != nil {
			t.Fatal(err)
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
		reqURL:        testServerURL,
		expStatusCode: http.StatusOK,
		expBody:       "<html><body>Hello, world!</body></html>\n",
	}, {
		desc:          "With known path",
		reqURL:        testServerURL + `/index.js`,
		expStatusCode: http.StatusOK,
		expBody:       "var a = \"Hello, world!\"\n",
	}, {
		desc:          "With known path and subtree root",
		reqURL:        testServerURL + `/get/`,
		expStatusCode: http.StatusOK,
		expBody:       "map[]\nmap[]\n<nil>\n",
	}, {
		desc:          "With known path",
		reqURL:        testServerURL + `/get?k=v`,
		expStatusCode: http.StatusOK,
		expBody:       "map[k:[v]]\nmap[]\n<nil>\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var (
			ctx = context.Background()
			req *http.Request
			err error
		)

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, c.reqURL, nil)
		if err != nil {
			t.Fatal(err)
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
		reqURL:           testServerURL + `/`,
		expStatusCode:    http.StatusOK,
		expContentType:   []string{"text/html; charset=utf-8"},
		expContentLength: []string{"40"},
	}, {
		desc:           "With registered GET and subtree root",
		reqURL:         testServerURL + `/api/`,
		expStatusCode:  http.StatusOK,
		expContentType: []string{ContentTypeJSON},
	}, {
		desc:           "With registered GET",
		reqURL:         testServerURL + `/api?k=v`,
		expStatusCode:  http.StatusOK,
		expContentType: []string{ContentTypeJSON},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var (
			ctx = context.Background()
			req *http.Request
			err error
		)

		req, err = http.NewRequestWithContext(ctx, http.MethodHead, c.reqURL, nil)
		if err != nil {
			t.Fatal(err)
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
		reqURL:        testServerURL + `/`,
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PATCH and subtree root",
		reqURL:        testServerURL + `/patch/`,
		expStatusCode: http.StatusOK,
		expBody:       "map[]\nmap[]\n<nil>\n",
	}, {
		desc:          "With registered PATCH and query",
		reqURL:        testServerURL + `/patch?k=v`,
		expStatusCode: http.StatusOK,
		expBody:       "map[k:[v]]\nmap[]\n<nil>\n",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var ctx = context.Background()

		req, e := http.NewRequestWithContext(ctx, http.MethodPatch, c.reqURL, nil)
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
		reqURL:        testServerURL + `/`,
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered POST and subtree root",
		reqURL:        testServerURL + `/post/`,
		expStatusCode: http.StatusOK,
		expBody:       "map[]\nmap[]\n<nil>\n",
	}, {
		desc:          "With registered POST and query",
		reqURL:        testServerURL + `/post?k=v`,
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

		var ctx = context.Background()

		req, e := http.NewRequestWithContext(ctx, http.MethodPost, c.reqURL, &buf)
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
		reqURL:        testServerURL + `/`,
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With registered PUT and subtree root",
		reqURL:        testServerURL + `/put/`,
		expStatusCode: http.StatusOK,
	}, {
		desc:          "With registered PUT and query",
		reqURL:        testServerURL + `/put?k=v`,
		expStatusCode: http.StatusOK,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var ctx = context.Background()

		req, e := http.NewRequestWithContext(ctx, http.MethodPut, c.reqURL, nil)
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
		reqURL:        testServerURL + `/`,
		expStatusCode: http.StatusOK,
		expAllow:      "GET, HEAD, OPTIONS",
	}, {
		desc:          "With registered PATCH and subtree root",
		reqURL:        testServerURL + `/options/`,
		expStatusCode: http.StatusOK,
		expAllow:      "DELETE, OPTIONS, PATCH",
	}, {
		desc:          "With registered PATCH and query",
		reqURL:        testServerURL + `/options?k=v`,
		expStatusCode: http.StatusOK,
		expAllow:      "DELETE, OPTIONS, PATCH",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var ctx = context.Background()

		req, err := http.NewRequestWithContext(ctx, http.MethodOptions, c.reqURL, nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()

		gotAllow := res.Header.Get("Allow")

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode)
		test.Assert(t, "Allow", c.expAllow, gotAllow)
	}
}

func TestStatusError(t *testing.T) {
	var (
		cbError = func(_ *EndpointRequest) ([]byte, error) {
			return nil, &liberrors.E{
				Code:    http.StatusLengthRequired,
				Message: `Length required`,
			}
		}
		cbNoCode = func(_ *EndpointRequest) ([]byte, error) {
			return nil, liberrors.Internal(nil)
		}
		cbCustomErr = func(_ *EndpointRequest) ([]byte, error) {
			return nil, errors.New(`Custom error`)
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
		reqURL:        testServerURL + `/error/no-body?k=v`,
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"message":"Length required","code":411}`,
	}, {
		desc:          "With registered error binary",
		reqURL:        testServerURL + `/error/binary?k=v`,
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"message":"Length required","code":411}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        testServerURL + `/error/plain?k=v`,
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"message":"Length required","code":411}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        testServerURL + `/error/json?k=v`,
		expStatusCode: http.StatusLengthRequired,
		expBody:       `{"message":"Length required","code":411}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        testServerURL + `/error/no-code?k=v`,
		expStatusCode: http.StatusInternalServerError,
		expBody:       `{"message":"internal server error","name":"ERR_INTERNAL","code":500}`,
	}, {
		desc:          "With registered error plain",
		reqURL:        testServerURL + `/error/custom?k=v`,
		expStatusCode: http.StatusInternalServerError,
		expBody:       `{"message":"internal server error","name":"ERR_INTERNAL","code":500}`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var ctx = context.Background()

		req, e := http.NewRequestWithContext(ctx, http.MethodPost, c.reqURL, nil)
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
		var ctx = context.Background()

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, testServerURL+c.reqPath, nil)
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
		err = res.Body.Close()
		if err != nil {
			t.Fatalf("%s: %s", c.desc, err)
		}

		test.Assert(t, "response body", c.expResBody, string(gotBody))
	}
}

func TestServer_handleDelete(t *testing.T) {
	type testCase struct {
		tag string
		req libhttptest.SimulateRequest
	}

	var (
		srv = &Server{}
		err error
	)

	err = srv.RegisterEndpoint(Endpoint{
		Method:       RequestMethodDelete,
		Path:         `/a/b/c/:d/e`,
		RequestType:  RequestTypeNone,
		ResponseType: ResponseTypePlain,
		Call: func(epr *EndpointRequest) ([]byte, error) {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, `Request.Form=%v`, epr.HTTPRequest.Form)
			return buf.Bytes(), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var tdata *test.Data

	tdata, err = test.LoadData(`testdata/handleDelete_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var listCase = []testCase{{
		tag: `valid`,
		req: libhttptest.SimulateRequest{
			Method: http.MethodDelete,
			Path:   `/a/b/c/dddd/e`,
		},
	}}
	var (
		c      testCase
		result *libhttptest.SimulateResult
		tag    string
		exp    string
		got    []byte
	)
	for _, c = range listCase {
		result, err = libhttptest.Simulate(srv.ServeHTTP, &c.req)
		if err != nil {
			t.Fatal(err)
		}

		got, err = result.DumpRequest(nil)
		if err != nil {
			t.Fatal(err)
		}
		tag = c.tag + `:request_body`
		exp = string(tdata.Output[tag])
		test.Assert(t, tag, exp, string(got))

		got, err = result.DumpResponse(nil)
		if err != nil {
			t.Fatal(err)
		}
		tag = c.tag + `:response_body`
		exp = string(tdata.Output[tag])
		test.Assert(t, tag, exp, string(got))
	}
}

func TestServer_handleRange(t *testing.T) {
	var (
		clOpts = ClientOptions{
			ServerURL: testServerURL,
		}
		cl          = NewClient(clOpts)
		skipHeaders = []string{HeaderDate, HeaderETag}

		listTestData []*test.Data
		tdata        *test.Data
		res          *ClientResponse
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

		var req = ClientRequest{
			Path:   `/index.html`,
			Header: header,
		}

		res, err = cl.Get(req)
		if err != nil {
			t.Fatal(err)
		}

		// Replace boundary with fixed string.
		var params map[string]string

		_, params, _ = mime.ParseMediaType(res.HTTPResponse.Header.Get(HeaderContentType))

		var (
			boundary      = params[`boundary`]
			fixedBoundary = `1b4df158039f7cce`
		)

		var (
			tag = `http_headers`
			exp = tdata.Output[tag]
			got = dumpHTTPResponse(res.HTTPResponse, skipHeaders)
		)

		if len(boundary) != 0 {
			got = strings.ReplaceAll(got, boundary, fixedBoundary)
		}
		test.Assert(t, tag, string(exp), got)

		tag = `http_body`
		exp = tdata.Output[tag]

		// Replace the response body CRLF with LF.
		res.Body = bytes.ReplaceAll(res.Body, []byte("\r\n"), []byte("\n"))

		if len(boundary) != 0 {
			res.Body = bytes.ReplaceAll(res.Body, []byte(boundary), []byte(fixedBoundary))
		}

		test.Assert(t, tag, string(exp), string(res.Body))

		tag = `all_body`
		exp = tdata.Output[tag]
		got = dumpMultipartBody(res.HTTPResponse)

		test.Assert(t, tag, string(exp), got)
	}
}

func TestServer_handleRange_HEAD(t *testing.T) {
	var (
		clOpts = ClientOptions{
			ServerURL: testServerURL,
		}
		cl = NewClient(clOpts)

		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/server/head_range_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		ctx     = context.Background()
		url     = testServerURL + `/index.html`
		httpReq *http.Request
	)

	httpReq, err = http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	var httpRes *http.Response

	httpRes, err = cl.Client.Do(httpReq)
	if err != nil {
		t.Fatal(err)
	}
	err = httpRes.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	var (
		skipHeaders = []string{HeaderDate, HeaderETag}
		got         = dumpHTTPResponse(httpRes, skipHeaders)
		tag         = `http_headers`
		exp         = tdata.Output[tag]
	)
	test.Assert(t, tag, string(exp), got)
}

// Test HTTP Range request on big file using Range.
//
// When server receive,
//
//	GET /big
//	Range: bytes=0-
//
// and the requested resources is quite larger, where writing all content of
// file result in i/o timeout, it is best practice [1][2] if the server
// write only partial content and let the client continue with the
// subsequent Range request.
//
// In above case the server should response with,
//
//	HTTP/1.1 206 Partial content
//	Content-Range: bytes 0-<limit>/<size>
//	Content-Length: <limit>
//
// Where limit is maximum packet that is reasonable [3] for most of the
// client.
//
// [1]: https://stackoverflow.com/questions/63614008/how-best-to-respond-to-an-open-http-range-request
// [2]: https://bugzilla.mozilla.org/show_bug.cgi?id=570755
// [3]: https://docs.aws.amazon.com/whitepapers/latest/s3-optimizing-performance-best-practices/use-byte-range-fetches.html
func TestServerHandleRangeBig(t *testing.T) {
	var (
		pathBig     = `/big`
		tempDir     = t.TempDir()
		filepathBig = filepath.Join(tempDir, pathBig)
		bigSize     = 10485760 // 10MB
	)

	createBigFile(t, filepathBig, int64(bigSize))

	var (
		serverAddress = `127.0.0.1:22672`
		srv           *Server
	)

	srv = runServerFS(t, serverAddress, tempDir)
	defer func() {
		var errStop = srv.Stop(100 * time.Millisecond)
		if errStop != nil {
			log.Fatal(errStop)
		}
	}()

	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/server/range_big_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		clOpts = ClientOptions{
			ServerURL: `http://` + serverAddress,
		}
		cl = NewClient(clOpts)

		tag         = `HEAD /big`
		skipHeaders = []string{HeaderDate}

		res     *ClientResponse
		gotResp string
	)

	var req = ClientRequest{
		Path: pathBig,
	}

	res, err = cl.Head(req)
	if err != nil {
		t.Fatal(err)
	}

	gotResp = dumpHTTPResponse(res.HTTPResponse, skipHeaders)

	test.Assert(t, tag, string(tdata.Output[tag]), gotResp)
	test.Assert(t, tag+`- response body size`, 0, len(res.Body))

	var headers = http.Header{}

	headers.Set(HeaderRange, `bytes=0-`)

	req = ClientRequest{
		Path:   pathBig,
		Header: headers,
	}

	res, err = cl.Get(req)
	if err != nil {
		t.Fatal(err)
	}

	gotResp = dumpHTTPResponse(res.HTTPResponse, skipHeaders)
	tag = `GET /big:Range=0-`
	test.Assert(t, tag, string(tdata.Output[tag]), gotResp)
	test.Assert(t, tag+`- response body size`, DefRangeLimit, len(res.Body))
}

func TestServer_RegisterHandleFunc(t *testing.T) {
	var (
		serverOpts = ServerOptions{}

		server *Server
		err    error
	)
	server, err = NewServer(serverOpts)
	if err != nil {
		t.Fatal(err)
	}
	server.RegisterHandleFunc(`/no/method`, testHandleFunc)
	server.RegisterHandleFunc(`PUT /book/:id`, testHandleFunc)

	type testCase struct {
		request *http.Request
		tag     string
	}
	var listCase = []testCase{{
		tag:     `GET /no/method`,
		request: mustHTTPRequest(`GET`, `/no/method`, nil),
	}, {
		tag:     `POST /no/method`,
		request: mustHTTPRequest(`POST`, `/no/method`, nil),
	}, {
		tag:     `PUT /book/1`,
		request: mustHTTPRequest(`PUT`, `/book/1`, nil),
	}}

	var tdata *test.Data
	tdata, err = test.LoadData(`testdata/Server_RegisterHandleFunc_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		tcase    testCase
		httpResp *http.Response
		got      []byte
		exp      string
	)
	for _, tcase = range listCase {
		var respRec = httptest.NewRecorder()
		server.ServeHTTP(respRec, tcase.request)

		httpResp = respRec.Result()

		got, err = httputil.DumpResponse(httpResp, true)
		if err != nil {
			t.Fatal(err)
		}
		got = bytes.ReplaceAll(got, []byte("\r"), []byte(""))
		exp = string(tdata.Output[tcase.tag])
		test.Assert(t, tcase.tag, exp, string(got))
	}
}

func testHandleFunc(httpwriter http.ResponseWriter, httpreq *http.Request) {
	var (
		rawb []byte
		err  error
	)
	rawb, err = httputil.DumpRequest(httpreq, true)
	if err != nil {
		log.Fatalf(`%s: %s`, httpreq.URL, err)
	}
	httpwriter.Write(rawb)
	fmt.Fprintf(httpwriter, `Form: %+v`, httpreq.Form)
}

func TestServer_RegisterHandleFunc_duplicate(t *testing.T) {
	var (
		serverOpts = ServerOptions{}

		server *Server
		err    error
	)
	server, err = NewServer(serverOpts)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		var msg = recover()
		test.Assert(t, `recover on duplicate pattern`,
			`RegisterHandleFunc: RegisterEndpoint: ambigous endpoint "/no/method"`,
			msg,
		)
	}()

	server.RegisterHandleFunc(`/no/method`,
		func(httpwriter http.ResponseWriter, httpreq *http.Request) {
			return
		},
	)
	server.RegisterHandleFunc(`GET /no/method`,
		func(httpwriter http.ResponseWriter, httpreq *http.Request) {
			return
		},
	)
}

func createBigFile(t *testing.T, path string, size int64) {
	var (
		fbig *os.File
		err  error
	)

	fbig, err = os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	err = fbig.Truncate(size)
	if err != nil {
		t.Fatal(err)
	}

	err = fbig.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func mustHTTPRequest(method, url string, body []byte) (httpreq *http.Request) {
	var (
		reqbody = bytes.NewBuffer(body)
		err     error
	)

	httpreq, err = http.NewRequest(method, url, io.NopCloser(reqbody))
	if err != nil {
		panic(err.Error())
	}
	return httpreq
}

func runServerFS(t *testing.T, address, dir string) (srv *Server) {
	var (
		mfsOpts = &memfs.Options{
			Root:        dir,
			MaxFileSize: -1,
		}

		mfs *memfs.MemFS
		err error
	)

	mfs, err = memfs.New(mfsOpts)
	if err != nil {
		t.Fatal(err)
	}

	// Set the file modification time for predictable result.
	var (
		pathBigModTime = time.Date(2024, 1, 1, 1, 1, 1, 0, time.UTC)
		nodeBig        *memfs.Node
	)

	nodeBig, err = mfs.Get(`/big`)
	if err != nil {
		t.Fatal(err)
	}

	nodeBig.SetModTime(pathBigModTime)

	var (
		srvOpts = ServerOptions{
			Memfs:   mfs,
			Address: address,
		}
	)

	srv, err = NewServer(srvOpts)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		var err2 = srv.Start()
		if err2 != nil {
			log.Fatal(err2)
		}
	}()

	err = libnet.WaitAlive(`tcp`, address, 1*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	return srv
}
