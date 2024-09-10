// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package websocket

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

const (
	_qKeyTicket = "ticket"
)

func createClient(t *testing.T, endpoint string) (cl *Client) {
	cl = &Client{
		Endpoint: endpoint,
	}

	var err = cl.parseURI()
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
	type testCase struct {
		desc     string
		req      *http.Request
		query    url.Values
		expError string
	}

	var (
		wsURL *url.URL
		err   error
	)

	wsURL, err = url.ParseRequestURI(_testWSAddr)
	if err != nil {
		t.Fatal(err)
	}

	var cases = []testCase{{
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
		expError: "invalid server accept key",
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
		expError: "400 invalid HTTP method",
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
		expError: "400 bad request: header length is less than minimum",
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
		expError: "400 bad request: header length is less than minimum",
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
		expError: "400 invalid Connection header",
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
		expError: "400 bad request: header length is less than minimum",
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
		expError: "400 bad request: header length is less than minimum",
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
		expError: "400 bad request: header length is less than minimum",
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
		expError: "400 unsupported Sec-WebSocket-Version",
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
		expError: "400 Missing authorization",
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
		expError: "400 Missing authorization",
	}}

	var (
		bb   bytes.Buffer
		c    testCase
		cl   *Client
		v    []string
		path string
		k    string
		x    int
	)

	for _, c = range cases {
		t.Log(c.desc)

		bb.Reset()
		cl = createClient(t, _testWSAddr)
		path = c.req.URL.EscapedPath() + "?" + c.query.Encode()

		fmt.Fprintf(&bb, "%s %s HTTP/1.1\r\n", c.req.Method, path)

		for k, v = range c.req.Header {
			for x = range v {
				fmt.Fprintf(&bb, "%s: %s\r\n", k, v[x])
			}
		}

		fmt.Fprintf(&bb, "\r\n")

		_, err = cl.doHandshake("", bb.Bytes())
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
		}
	}
}

func TestServer_Health(t *testing.T) {
	var (
		ctx = context.Background()
		url = `http://` + _testAddr + `/status`

		httpReq *http.Request
		err     error
	)

	httpReq, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	var res *http.Response

	res, err = http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()

	test.Assert(t, "/status response code", http.StatusOK, res.StatusCode)
}

// TestServer_upgrader test to make sure that server upgrade does not block
// other requests.
func TestServer_upgrader_nonblocking(t *testing.T) {
	var (
		err error
	)

	// Open new connection that does not send anything, that will trigger
	// the server Accept and continue to Recv.
	_, err = net.Dial(`tcp`, _testAddr)
	if err != nil {
		t.Fatal(err)
	}

	// Create new client that send text.
	// The client should able to receive response without waiting the
	// above connection for timeout.
	var (
		qtext = make(chan []byte, 1)
		cl    = &Client{
			Endpoint: _testEndpointAuth,
			HandleText: func(_ *Client, frame *Frame) (err error) {
				qtext <- frame.Payload()
				return nil
			},
		}
	)

	err = cl.Connect()
	if err != nil {
		t.Fatal(err)
	}

	var (
		msg = []byte(`hello world`)
		got []byte
	)

	err = cl.SendText(msg)
	if err != nil {
		t.Fatal(err)
	}

	got = <-qtext
	test.Assert(t, `SendText`, msg, got)
}
