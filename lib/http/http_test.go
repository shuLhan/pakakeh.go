// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

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
		HandleFSAuth: handleFSAuth,
		Address:      serverAddress,
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

	time.Sleep(400 * time.Millisecond) // Wait for server to be ready.

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

// handleFSAuth authenticate the request to Memfs using cookie.
// It will return true if request path is "/auth/" and cookie name "sid" exist
// with value "authz".
func handleFSAuth(req *http.Request) bool {
	var (
		lowerPath = strings.ToLower(req.URL.Path)

		cookieSid *http.Cookie
		err       error
	)
	log.Printf("handleFSAuth: %s", lowerPath)
	if !strings.HasPrefix(lowerPath, "/auth/") {
		return true
	}
	cookieSid, err = req.Cookie("sid")
	if err != nil {
		return false
	}
	if cookieSid.Value != "authz" {
		return false
	}
	return true
}

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
