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

	"github.com/shuLhan/share/lib/memfs"
)

var (
	testServer    *Server
	testServerUrl string

	client = &http.Client{}

	cbNone = func(epr *EndpointRequest) ([]byte, error) {
		return nil, nil
	}

	cbPlain = func(epr *EndpointRequest) (resBody []byte, e error) {
		s := fmt.Sprintf("%s\n", epr.HttpRequest.Form)
		s += fmt.Sprintf("%s\n", epr.HttpRequest.PostForm)
		s += fmt.Sprintf("%v\n", epr.HttpRequest.MultipartForm)
		s += string(epr.RequestBody)
		return []byte(s), nil
	}

	cbJSON = func(epr *EndpointRequest) (resBody []byte, e error) {
		s := fmt.Sprintf(`{
"form": "%s",
"multipartForm": "%v",
"body": %q
}`, epr.HttpRequest.Form, epr.HttpRequest.MultipartForm, epr.RequestBody)
		return []byte(s), nil
	}
)

func TestMain(m *testing.M) {
	var (
		serverAddress = "127.0.0.1:14832"
		err           error
	)

	opts := &ServerOptions{
		Memfs: &memfs.MemFS{
			Opts: &memfs.Options{
				Root:        "./testdata",
				MaxFileSize: 30,
				Development: true,
			},
		},
		Address: serverAddress,
	}

	testServerUrl = fmt.Sprintf("http://" + serverAddress)

	testServer, err = NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	registerEndpoints()

	go func() {
		err := testServer.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	status := m.Run()

	err = testServer.Stop(0)
	if err != nil {
		log.Println("testServer.Stop: ", err)
	}

	os.Exit(status)
}

var (
	testDownloadBody []byte
)

func registerEndpoints() {
	var err error

	testDownloadBody, err = os.ReadFile("client.go")
	if err != nil {
		log.Fatalf("TestMain: %s", err)
	}

	// Endpoint to test the client Download().
	err = testServer.RegisterEndpoint(&Endpoint{
		Path:         "/download",
		ResponseType: ResponseTypePlain,
		Call: func(epr *EndpointRequest) ([]byte, error) {
			return testDownloadBody, nil
		},
	})
	if err != nil {
		log.Fatalf("TestMain: %s", err)
	}

	// Endpoint to test the client Download() with HTTP 302 redirect.
	err = testServer.RegisterEndpoint(&Endpoint{
		Path:         "/redirect/download",
		ResponseType: ResponseTypePlain,
		Call: func(epr *EndpointRequest) ([]byte, error) {
			http.Redirect(epr.HttpWriter, epr.HttpRequest, "/download", http.StatusFound)
			return nil, nil
		},
	})
	if err != nil {
		log.Fatalf("TestMain: %s", err)
	}
}
