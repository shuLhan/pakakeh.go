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

var testServer *Server // nolint: gochecknoglobals

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
	cb := func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf("%s\n", req.Form)
		s += fmt.Sprintf("%v\n", req.MultipartForm)
		s += fmt.Sprintf("%s", reqBody)
		return []byte(s), nil
	}

	client := &http.Client{}

	testServer.RegisterDelete("/delete", ResponseTypePlain, cb)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expBody       []byte
	}{{
		desc:          "With unknown path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With known path and subtree root",
		reqURL:        "http://127.0.0.1:8080/delete/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With known path",
		reqURL:        "http://127.0.0.1:8080/delete?k=v",
		expStatusCode: http.StatusOK,
		expBody:       []byte("map[k:[v]]\n<nil>\n"),
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
		test.Assert(t, "Body", string(c.expBody), string(body), true)
	}
}

func TestRegisterGet(t *testing.T) {
	cb := func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf("%s\n", req.Form)
		s += fmt.Sprintf("%v\n", req.MultipartForm)
		s += fmt.Sprintf("%s", reqBody)
		return []byte(s), nil
	}

	client := &http.Client{}

	testServer.RegisterGet("/get", ResponseTypePlain, cb)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expBody       []byte
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusOK,
		expBody:       []byte("<html><body>Hello, world!</body></html>\n"),
	}, {
		desc:          "With known path and subtree root",
		reqURL:        "http://127.0.0.1:8080/get/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With known path",
		reqURL:        "http://127.0.0.1:8080/get?k=v",
		expStatusCode: http.StatusOK,
		expBody:       []byte("map[k:[v]]\n<nil>\n"),
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
		test.Assert(t, "Body", string(c.expBody), string(body), true)
	}
}

func TestRegisterHead(t *testing.T) {
	cb := func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		return
	}

	client := &http.Client{}

	testServer.RegisterGet("/api", ResponseTypeJSON, cb)

	cases := []struct {
		desc             string
		reqURL           string
		expStatusCode    int
		expBody          []byte
		expContentType   []string
		expContentLength []string
	}{{
		desc:             "With root path",
		reqURL:           "http://127.0.0.1:8080/",
		expStatusCode:    http.StatusOK,
		expBody:          []byte{},
		expContentType:   []string{"text/html; charset=utf-8"},
		expContentLength: []string{"40"},
	}, {
		desc:          "With registered GET and subtree root",
		reqURL:        "http://127.0.0.1:8080/api/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:           "With registered GET",
		reqURL:         "http://127.0.0.1:8080/api?k=v",
		expStatusCode:  http.StatusOK,
		expBody:        []byte{},
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
		test.Assert(t, "Body", string(c.expBody), string(body), true)
		test.Assert(t, "Header.ContentType", c.expContentType,
			res.Header[contentType], true)
		test.Assert(t, "Header.ContentLength", c.expContentLength,
			res.Header[contentLength], true)
	}
}

func TestRegisterPatch(t *testing.T) {
	cb := func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf("%s\n", req.Form)
		s += fmt.Sprintf("%v\n", req.MultipartForm)
		s += fmt.Sprintf("%s", reqBody)
		return []byte(s), nil
	}

	client := &http.Client{}

	testServer.RegisterPatch("/patch", RequestTypeQuery, ResponseTypePlain, cb)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expBody       []byte
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With registered PATCH and subtree root",
		reqURL:        "http://127.0.0.1:8080/patch/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With registered PATCH and query",
		reqURL:        "http://127.0.0.1:8080/patch?k=v",
		expStatusCode: http.StatusOK,
		expBody:       []byte("map[k:[v]]\n<nil>\n"),
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
		test.Assert(t, "Body", string(c.expBody), string(body), true)
	}
}

func TestRegisterPost(t *testing.T) {
	cb := func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf("%s\n", req.Form)
		s += fmt.Sprintf("%s\n", req.PostForm)
		s += fmt.Sprintf("%v\n", req.MultipartForm)
		s += fmt.Sprintf("%s", reqBody)
		return []byte(s), nil
	}

	client := &http.Client{}

	testServer.RegisterPost("/post", RequestTypeForm, ResponseTypePlain, cb)

	cases := []struct {
		desc          string
		reqURL        string
		reqBody       []byte
		expStatusCode int
		expBody       []byte
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With registered POST and subtree root",
		reqURL:        "http://127.0.0.1:8080/post/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With registered POST and query",
		reqURL:        "http://127.0.0.1:8080/post?k=v",
		reqBody:       []byte("k=vv"),
		expStatusCode: http.StatusOK,
		expBody: []byte(`map[k:[vv v]]
map[k:[vv]]
<nil>
`),
	}}

	for _, c := range cases {
		t.Log(c.desc)

		var buf bytes.Buffer
		_, _ = buf.Write(c.reqBody)

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
		test.Assert(t, "Body", string(c.expBody), string(body), true)
	}
}

func TestRegisterPut(t *testing.T) {
	cb := func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf("%s\n", req.Form)
		s += fmt.Sprintf("%v\n", req.MultipartForm)
		s += fmt.Sprintf("%s", reqBody)
		return []byte(s), nil
	}

	client := &http.Client{}

	testServer.RegisterPut("/put", RequestTypeForm, cb)

	cases := []struct {
		desc          string
		reqURL        string
		expStatusCode int
		expBody       []byte
	}{{
		desc:          "With root path",
		reqURL:        "http://127.0.0.1:8080/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With registered PUT and subtree root",
		reqURL:        "http://127.0.0.1:8080/put/",
		expStatusCode: http.StatusNotFound,
		expBody:       []byte{},
	}, {
		desc:          "With registered PUT and query",
		reqURL:        "http://127.0.0.1:8080/put?k=v",
		expStatusCode: http.StatusNoContent,
		expBody:       []byte{},
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
		test.Assert(t, "Body", string(c.expBody), string(body), true)
	}
}

func TestServeHTTPOptions(t *testing.T) {
	cb := func(req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf("%s\n", req.Form)
		s += fmt.Sprintf("%v\n", req.MultipartForm)
		s += fmt.Sprintf("%s", reqBody)
		return []byte(s), nil
	}

	client := &http.Client{}

	testServer.RegisterDelete("/options", ResponseTypePlain, cb)
	testServer.RegisterPatch("/options", RequestTypeQuery,
		ResponseTypePlain, cb)

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
