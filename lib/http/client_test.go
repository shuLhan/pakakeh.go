// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"bytes"
	"fmt"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestClient_Download(t *testing.T) {
	var (
		logp       = `Download`
		clientOpts = ClientOptions{
			ServerURL: `http://` + testServer.Options.Address,
		}
		client = NewClient(clientOpts)

		out bytes.Buffer
		err error
	)

	cases := []struct {
		desc     string
		expError string
		req      DownloadRequest
	}{{
		desc: "With nil Output",
		req: DownloadRequest{
			ClientRequest: ClientRequest{
				Path: "/redirect/downloads",
			},
		},
		expError: fmt.Sprintf("%s: %s", logp, ErrClientDownloadNoOutput),
	}, {
		desc: "With invalid path",
		req: DownloadRequest{
			ClientRequest: ClientRequest{
				Path: "/redirect/downloads",
			},
			Output: &out,
		},
		expError: logp + `: 404 Not Found`,
	}, {
		desc: "With redirect",
		req: DownloadRequest{
			ClientRequest: ClientRequest{
				Path: "/redirect/download",
			},
			Output: &out,
		},
	}, {
		desc: "With redirect and trailing slash",
		req: DownloadRequest{
			ClientRequest: ClientRequest{
				Path: "/redirect/download/",
			},
			Output: &out,
		},
	}, {
		desc: "With direct path",
		req: DownloadRequest{
			ClientRequest: ClientRequest{
				Path: "/download",
			},
			Output: &out,
		},
	}, {
		desc: "With direct path and trailing slash",
		req: DownloadRequest{
			ClientRequest: ClientRequest{
				Path: "/download/",
			},
			Output: &out,
		},
	}}

	for _, c := range cases {
		out.Reset()

		_, err = client.Download(c.req) //nolint: bodyclose
		if err != nil {
			test.Assert(t, c.desc+`: error`, c.expError, err.Error())
			continue
		}

		test.Assert(t, c.desc, testDownloadBody, out.Bytes())
	}
}
