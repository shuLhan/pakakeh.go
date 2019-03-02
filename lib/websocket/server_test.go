// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websocket

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestNewServer(t *testing.T) {
	cases := []struct {
		desc   string
		port   int
		expErr string
	}{{
		desc:   "With invalid port",
		port:   -1,
		expErr: "invalid argument",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_, err := NewServer(c.port)
		if err != nil {
			test.Assert(t, "err", c.expErr, err.Error(), true)
			continue
		}
	}
}

func createClient(t *testing.T, endpoint string) (cl *Client) {
	cl = &Client{}

	err := cl.parseURI(endpoint)
	if err != nil {
		t.Fatal(err)
		return
	}

	err = cl.open()
	if err != nil {
		t.Fatal(err)
		return
	}

	return
}

func TestServerHandshake(t *testing.T) {
	wsURL, err := url.ParseRequestURI(_testWSAddr)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc        string
		req         *http.Request
		query       url.Values
		expKey      string
		expRespCode int
	}{{
		desc: "With valid request and authorization",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{_hdrValConnectionUpgrade},
				_hdrKeyUpgrade:    []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:      []string{_testHdrValWSKey},
				_hdrKeyWSVersion:  []string{_hdrValWSVersion},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expKey:      _testHdrValWSAccept,
		expRespCode: http.StatusSwitchingProtocols,
	}, {
		desc: "Without GET",
		req: &http.Request{
			Method: http.MethodPost,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{_hdrValConnectionUpgrade},
				_hdrKeyUpgrade:    []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:      []string{_testHdrValWSKey},
				_hdrKeyWSVersion:  []string{_hdrValWSVersion},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "Without HTTP header Host",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyUpgrade:   []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:     []string{_testHdrValWSKey},
				_hdrKeyWSVersion: []string{_hdrValWSVersion},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "Without HTTP header Connection",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:      []string{"127.0.0.1"},
				_hdrKeyUpgrade:   []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:     []string{_testHdrValWSKey},
				_hdrKeyWSVersion: []string{_hdrValWSVersion},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "With invalid HTTP header Connection",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{"upgraade"},
				_hdrKeyUpgrade:    []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:      []string{_testHdrValWSKey},
				_hdrKeyWSVersion:  []string{_hdrValWSVersion},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "Without HTTP header Upgrade",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{_hdrValConnectionUpgrade},
				_hdrKeyWSKey:      []string{_testHdrValWSKey},
				_hdrKeyWSVersion:  []string{_hdrValWSVersion},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "Without HTTP header 'Sec-Websocket-Key'",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{_hdrValConnectionUpgrade},
				_hdrKeyUpgrade:    []string{_hdrValUpgradeWS},
				_hdrKeyWSVersion:  []string{_hdrValWSVersion},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "Without HTTP header 'Sec-Websocket-Version'",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{_hdrValConnectionUpgrade},
				_hdrKeyUpgrade:    []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:      []string{_testHdrValWSKey},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "With unsupported websocket version",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{_hdrValConnectionUpgrade},
				_hdrKeyUpgrade:    []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:      []string{_testHdrValWSKey},
				_hdrKeyWSVersion:  []string{"11"},
			},
		},
		query: url.Values{
			_qKeyTicket: []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "Without authorization",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{_hdrValConnectionUpgrade},
				_hdrKeyUpgrade:    []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:      []string{_testHdrValWSKey},
				_hdrKeyWSVersion:  []string{_hdrValWSVersion},
			},
		},
		expRespCode: http.StatusBadRequest,
	}, {
		desc: "Without invalid HTTP header 'Authorization'",
		req: &http.Request{
			Method: http.MethodGet,
			URL:    wsURL,
			Header: http.Header{
				_hdrKeyHost:       []string{"127.0.0.1"},
				_hdrKeyConnection: []string{_hdrValConnectionUpgrade},
				_hdrKeyUpgrade:    []string{_hdrValUpgradeWS},
				_hdrKeyWSKey:      []string{_testHdrValWSKey},
				_hdrKeyWSVersion:  []string{_hdrValWSVersion},
			},
		},
		query: url.Values{
			"Basic": []string{_testExternalJWT},
		},
		expRespCode: http.StatusBadRequest,
	}}

	var bb bytes.Buffer

	for _, c := range cases {
		t.Log(c.desc)

		bb.Reset()
		cl := createClient(t, _testWSAddr)
		path := c.req.URL.EscapedPath() + "?" + c.query.Encode()

		fmt.Fprintf(&bb, "%s %s HTTP/1.1\r\n", c.req.Method, path)

		for k, v := range c.req.Header {
			for x := range v {
				fmt.Fprintf(&bb, "%s: %s\r\n", k, v[x])
			}
		}

		fmt.Fprintf(&bb, "\r\n")

		c := c
		handleHandshake := func(ctx context.Context, resp []byte) (err error) {
			httpBuf := bufio.NewReader(bytes.NewBuffer(resp))

			httpRes, err := http.ReadResponse(httpBuf, nil)
			if err != nil {
				t.Fatal(err)
				return
			}

			test.Assert(t, "expRespCode", c.expRespCode, httpRes.StatusCode, true)

			return
		}

		err = cl.Send(context.Background(), bb.Bytes(), handleHandshake)
		if err != nil {
			t.Fatal(err)
		}
	}
}
