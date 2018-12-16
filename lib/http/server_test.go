// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var ( // nolint: gochecknoglobals
	testServer *Server // nolint: gochecknoglobals
	client     = &http.Client{}

	cbNone = func(req *http.Request, reqBody []byte) ([]byte, error) {
		return nil, nil
	}

	cbPlain = func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf("%s\n", req.Form)
		s += fmt.Sprintf("%s\n", req.PostForm)
		s += fmt.Sprintf("%v\n", req.MultipartForm)
		s += fmt.Sprintf("%s", reqBody)
		return []byte(s), nil
	}

	cbJSON = func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf(`{
"form": "%s",
"multipartForm": "%v",
"body": %q
}`, req.Form, req.MultipartForm, reqBody)
		return []byte(s), nil
	}
)

func TestMain(m *testing.M) {
	var e error

	conn := &http.Server{
		Addr: "127.0.0.1:8080",
	}

	testServer, e = NewServer("testdata", conn)
	if e != nil {
		log.Fatal(e)
	}

	go func() {
		e = testServer.Start()
		if e != nil {
			log.Fatal(e)
		}
	}()

	os.Exit(m.Run())
}

func TestRegisterDelete(t *testing.T) {
	cases := []struct {
		desc          string
		reqURL        string
		resType       ResponseType
		expStatusCode int
		expBody       string
	}{{
		desc:          "With unknown path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With known path and subtree root",
		reqURL:        "http://127.0.0.1:8080/delete/",
		expStatusCode: http.StatusNotFound,
	}, {
		desc:          "With response type none",
		reqURL:        "http://127.0.0.1:8080/delete?k=v",
		resType:       ResponseTypeNone,
		expStatusCode: http.StatusNoContent,
	}, {
		desc:          "With response type binary",
		reqURL:        "http://127.0.0.1:8080/delete?k=v",
		resType:       ResponseTypeBinary,
		expStatusCode: http.StatusOK,
		expBody:       "map[k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc:          "With response type plain",
		reqURL:        "http://127.0.0.1:8080/delete?k=v",
		resType:       ResponseTypePlain,
		expStatusCode: http.StatusOK,
		expBody:       "map[k:[v]]\nmap[]\n<nil>\n",
	}, {
		desc:          "With response type JSON",
		reqURL:        "http://127.0.0.1:8080/delete?k=v",
		resType:       ResponseTypeJSON,
		expStatusCode: http.StatusOK,
		expBody: `{
"form": "map[k:[v]]",
"multipartForm": "<nil>",
"body": ""
}`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		switch c.resType {
		case ResponseTypeNone:
			testServer.RegisterDelete("/delete",
				ResponseTypeNone, cbNone)
		case ResponseTypeBinary:
			testServer.RegisterDelete("/delete",
				ResponseTypeBinary, cbPlain)
		case ResponseTypeJSON:
			testServer.RegisterDelete("/delete",
				ResponseTypeJSON, cbJSON)
		default:
			testServer.RegisterDelete("/delete",
				ResponseTypePlain, cbPlain)
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

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		if c.expStatusCode != http.StatusOK {
			continue
		}

		test.Assert(t, "Body", c.expBody, string(body), true)

		var expContentType string
		gotContentType := res.Header.Get(contentType)

		switch c.resType {
		case ResponseTypeBinary:
			expContentType = contentTypeBinary
		case ResponseTypeJSON:
			expContentType = contentTypeJSON
		default:
			expContentType = contentTypePlain
		}

		test.Assert(t, "Content-Type", expContentType, gotContentType,
			true)
	}
}

func TestRegisterGet(t *testing.T) {
	testServer.RegisterGet("/get", ResponseTypePlain, cbPlain)

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
	testServer.RegisterGet("/api", ResponseTypeJSON, cbNone)

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
		expContentType: []string{contentTypeJSON},
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
			res.Header[contentType], true)
		test.Assert(t, "Header.ContentLength", c.expContentLength,
			res.Header[contentLength], true)
	}
}

func TestRegisterPatch(t *testing.T) {
	testServer.RegisterPatch("/patch", RequestTypeQuery,
		ResponseTypePlain, cbPlain)

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
	testServer.RegisterPost("/post", RequestTypeForm, ResponseTypePlain,
		cbPlain)

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
`,
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var buf bytes.Buffer
		_, _ = buf.WriteString(c.reqBody)

		req, e := http.NewRequest(http.MethodPost, c.reqURL, &buf)
		if e != nil {
			t.Fatal(e)
		}

		req.Header.Set(contentType, contentTypeForm)

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
	testServer.RegisterPut("/put", RequestTypeForm, cbPlain)

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
	testServer.RegisterDelete("/options", ResponseTypePlain, cbPlain)
	testServer.RegisterPatch("/options", RequestTypeQuery,
		ResponseTypePlain, cbPlain)

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

		req, e := http.NewRequest(http.MethodOptions, c.reqURL, nil)
		if e != nil {
			t.Fatal(e)
		}

		res, e := client.Do(req)
		if e != nil {
			t.Fatal(e)
		}

		gotAllow := res.Header.Get("Allow")

		test.Assert(t, "StatusCode", c.expStatusCode, res.StatusCode,
			true)
		test.Assert(t, "Allow", c.expAllow, gotAllow, true)
	}
}
