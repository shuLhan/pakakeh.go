// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

//
// DoHClient client for DNS over HTTPS.
//
type DoHClient struct {
	addr    *url.URL
	headers http.Header
	req     *http.Request
	query   url.Values
	conn    *http.Client

	// w hold the ResponseWriter on receiver side.
	w http.ResponseWriter
	// responded is a channel to signal the underlying receiver that the
	// response has ready to be send to client.
	responded chan bool
}

//
// NewDoHClient will create new DNS client with HTTP connection.
//
func NewDoHClient(nameserver string, allowInsecure bool) (*DoHClient, error) {
	nsURL, err := url.Parse(nameserver)
	if err != nil {
		return nil, err
	}

	if nsURL.Scheme != "https" {
		err = fmt.Errorf("DoH name server must be HTTPS")
		return nil, err
	}

	tr := &http.Transport{
		MaxIdleConns:    1,
		IdleConnTimeout: 30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: allowInsecure,
		},
	}

	cl := &DoHClient{
		addr: nsURL,
		headers: http.Header{
			"accept": []string{
				"application/dns-message",
			},
		},
		query: nsURL.Query(),
		conn: &http.Client{
			Transport: tr,
			Timeout:   clientTimeout,
		},
	}

	cl.req = &http.Request{
		Method:     http.MethodGet,
		URL:        nsURL,
		Proto:      "HTTP/2",
		ProtoMajor: 2,
		ProtoMinor: 0,
		Header:     cl.headers,
		Body:       nil,
		Host:       nsURL.Hostname(),
	}

	return cl, nil
}

//
// Close all idle connections.
//
func (cl *DoHClient) Close() error {
	cl.conn.Transport.(*http.Transport).CloseIdleConnections()
	return nil
}

//
// Lookup DNS records based on MessageQuestion Name and Type, in synchronous
// mode.
// The MessageQuestion Class default to IN.
//
// It will return an error if the Name is empty.
//
func (cl *DoHClient) Lookup(q MessageQuestion, allowRecursion bool) (res *Message, err error) {
	if len(q.Name) == 0 {
		return nil, fmt.Errorf("Lookup: empty question name")
	}
	if q.Type == 0 {
		q.Type = RecordTypeA
	}
	if q.Class == 0 {
		q.Class = RecordClassIN
	}

	msg := NewMessage()

	// No ID.
	// HTTP correlates the request and response, thus eliminating
	// the need for the ID in a media type such as
	// "application/dns-message".
	// The use of a varying DNS ID can cause semantically equivalent DNS
	// queries to be cached separately.
	// -- RFC8484 4.1
	msg.Header.IsRD = allowRecursion
	msg.Question = q

	_, err = msg.Pack()
	if err != nil {
		return nil, fmt.Errorf("Lookup: %w", err)
	}

	res, err = cl.Get(msg)
	if err != nil {
		return nil, fmt.Errorf("Lookup: %w", err)
	}

	return res, err
}

//
// Post send query to name server using HTTP POST and return the response
// as unpacked message.
//
func (cl *DoHClient) Post(msg *Message) (*Message, error) {
	cl.req.Method = http.MethodPost
	cl.req.Body = ioutil.NopCloser(bytes.NewReader(msg.packet))
	cl.req.URL.RawQuery = ""

	httpRes, err := cl.conn.Do(cl.req)
	if err != nil {
		cl.req.Body.Close()
		return nil, err
	}
	cl.req.Body.Close()

	res := NewMessage()

	res.packet, err = ioutil.ReadAll(httpRes.Body)
	httpRes.Body.Close()
	if err != nil {
		return nil, err
	}

	err = res.Unpack()

	return res, err
}

//
// Get send query to name server using HTTP GET and return the response as
// unpacked message.
//
func (cl *DoHClient) Get(msg *Message) (*Message, error) {
	q := base64.RawURLEncoding.EncodeToString(msg.packet)

	cl.query.Set("dns", q)
	cl.req.Method = http.MethodGet
	cl.req.Body = nil
	cl.req.URL.RawQuery = cl.query.Encode()

	httpRes, err := cl.conn.Do(cl.req)
	if err != nil {
		return nil, err
	}

	res := NewMessage()

	res.packet, err = ioutil.ReadAll(httpRes.Body)
	httpRes.Body.Close()
	if err != nil {
		return nil, err
	}
	if httpRes.StatusCode != 200 {
		err = fmt.Errorf("%s", string(res.packet))
		return nil, err
	}

	if len(res.packet) > 20 {
		err = res.Unpack()
		if err != nil {
			return nil, err
		}
	}

	return res, err
}

//
// Query send DNS query to name server.  This is an alias to Get method, to
// make it consistent with other DNS clients.
//
func (cl *DoHClient) Query(msg *Message) (*Message, error) {
	return cl.Get(msg)
}

//
// RemoteAddr return client remote nameserver address.
//
func (cl *DoHClient) RemoteAddr() string {
	return cl.addr.String()
}

//
// SetRemoteAddr set the remote address for sending the packet.
//
func (cl *DoHClient) SetRemoteAddr(addr string) (err error) {
	cl.addr, err = url.Parse(addr)
	if err != nil {
		return
	}

	cl.query = cl.addr.Query()

	return
}

//
// SetTimeout set the timeout for sending and receiving packet.
//
func (cl *DoHClient) SetTimeout(t time.Duration) {
	cl.conn.Timeout = t
}

//
// Write the raw DNS response message to active connection.
// This method is only used by server to write the response of query to
// client.
//
func (cl *DoHClient) Write(packet []byte) (n int, err error) {
	n, err = cl.w.Write(packet)
	if err != nil {
		cl.responded <- false
		return
	}
	cl.responded <- true
	return
}

//
// waitResponse wait for http.ResponseWriter being called by server.
// This method is to prevent the function that process the HTTP request
// terminated and write empty response.
//
func (cl *DoHClient) waitResponse() {
	success, ok := <-cl.responded
	if !success || !ok {
		cl.w.WriteHeader(http.StatusGatewayTimeout)
	}
}