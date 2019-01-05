// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	libmemfs "github.com/shuLhan/share/lib/memfs"
)

var ( // nolint: gochecknoglobals
	testServer *Server // nolint: gochecknoglobals
	client     = &http.Client{}

	cbNone = func(res http.ResponseWriter, req *http.Request, reqBody []byte) (
		[]byte, error,
	) {
		return nil, nil
	}

	cbPlain = func(res http.ResponseWriter, req *http.Request, reqBody []byte) (
		resBody []byte, e error,
	) {
		s := fmt.Sprintf("%s\n", req.Form)
		s += fmt.Sprintf("%s\n", req.PostForm)
		s += fmt.Sprintf("%v\n", req.MultipartForm)
		s += fmt.Sprintf("%s", reqBody)
		return []byte(s), nil
	}

	cbJSON = func(res http.ResponseWriter, req *http.Request, reqBody []byte) (
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
	var err error

	conn := &http.Server{
		Addr: "127.0.0.1:8080",
	}

	// Testing handleFS with large size.
	libmemfs.MaxFileSize = 30

	testServer, err = NewServer("testdata", conn)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err = testServer.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	os.Exit(m.Run())
}
