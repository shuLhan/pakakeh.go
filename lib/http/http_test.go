// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/memfs"
	libstrings "github.com/shuLhan/share/lib/strings"
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
				TryDirect:   true,
			},
		},
		HandleFS: handleFS,
		Address:  serverAddress,
	}

	testServerUrl = fmt.Sprintf("http://" + serverAddress)

	testServer, err = NewServer(opts)
	if err != nil {
		log.Fatal(err)
	}

	registerEndpoints()

	go func() {
		var errStart = testServer.Start()
		if errStart != nil {
			log.Fatal(errStart)
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

// handleFS authenticate the request to Memfs using cookie.
//
// If the node does not start with "/auth/" it will return true.
//
// If the node path is start with "/auth/" and cookie name "sid" exist
// with value "authz" it will return true;
// otherwise it will redirect to "/" and return false.
func handleFS(node *memfs.Node, res http.ResponseWriter, req *http.Request) bool {
	var (
		lowerPath = strings.ToLower(node.Path)

		cookieSid *http.Cookie
		err       error
	)
	if strings.HasPrefix(lowerPath, "/auth/") {
		cookieSid, err = req.Cookie("sid")
		if err != nil {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return false
		}
		if cookieSid.Value != "authz" {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return false
		}
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

// dumpHttpResponse write headers ordered by key in ascending with option to
// skip certain header keys.
func dumpHttpResponse(httpRes *http.Response, skipHeaders []string) string {
	var (
		keys []string
		hkey string
	)
	for hkey = range httpRes.Header {
		if libstrings.IsContain(skipHeaders, hkey) {
			continue
		}
		keys = append(keys, hkey)
	}
	sort.Strings(keys)

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s\n", httpRes.Proto, httpRes.Status)
	for _, hkey = range keys {
		fmt.Fprintf(&sb, "%s: %s\n", hkey, httpRes.Header.Get(hkey))
	}
	return sb.String()
}

// dumpMultipartBody Concatenate each multipart body into one string.
// If the the Content-Type header is not multipart, it will return all the
// body.
func dumpMultipartBody(httpRes *http.Response) string {
	var (
		logp        = `dumpMultipartBody`
		contentType = httpRes.Header.Get(`Content-Type`)

		mediaType string
		params    map[string]string
		err       error
	)

	mediaType, params, err = mime.ParseMediaType(contentType)
	if err != nil {
		log.Fatalf(`%s: ParseMediaType: %s`, logp, err)
	}

	var body []byte

	if !strings.HasPrefix(mediaType, `multipart/`) {
		body, err = io.ReadAll(httpRes.Body)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return ``
			}
			log.Fatalf(`%s: ReadAll httpRes.Body: %s`, logp, err)
		}
		return string(body)
	}

	var (
		reader *multipart.Reader
		part   *multipart.Part
		sb     strings.Builder
	)

	reader = multipart.NewReader(httpRes.Body, params[`boundary`])
	for {
		part, err = reader.NextPart()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatalf(`%s: NextPart: %s`, logp, err)
		}
		body, err = io.ReadAll(part)
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			log.Fatalf(`%s: ReadAll part: %s`, logp, err)
		}
		sb.Write(body)
	}
	return sb.String()
}
