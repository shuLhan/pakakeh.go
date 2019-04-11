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
	"net"
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
			InsecureSkipVerify: allowInsecure, // nolint: gosec
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
// Lookup will query the DoH server with specific type, class, and name in
// synchronous mode.
//
func (cl *DoHClient) Lookup(allowRecursion bool, qtype, qclass uint16, qname []byte) (*Message, error) {
	if len(qname) == 0 {
		return nil, nil
	}
	if qtype == 0 {
		qtype = QueryTypeA
	}
	if qclass == 0 {
		qclass = QueryClassIN
	}

	msg := NewMessage()

	msg.Header.IsRD = allowRecursion
	msg.Question.Type = qtype
	msg.Question.Class = qclass
	msg.Question.Name = append(msg.Question.Name, qname...)

	_, err := msg.Pack()
	if err != nil {
		return nil, err
	}

	res, err := cl.Get(msg)
	if err != nil {
		return nil, err
	}

	return res, err
}

//
// Post send query to name server using HTTP POST and return the response
// as unpacked message.
//
func (cl *DoHClient) Post(msg *Message) (*Message, error) {
	cl.req.Method = http.MethodPost
	cl.req.Body = ioutil.NopCloser(bytes.NewReader(msg.Packet))
	cl.req.URL.RawQuery = ""

	httpRes, err := cl.conn.Do(cl.req)
	if err != nil {
		cl.req.Body.Close()
		return nil, err
	}
	cl.req.Body.Close()

	res := NewMessage()

	packet, err := ioutil.ReadAll(httpRes.Body)
	httpRes.Body.Close()
	if err != nil {
		return nil, err
	}

	res.Packet = append(res.Packet[:0], packet...)

	err = res.Unpack()

	return res, err
}

//
// Get send query to name server using HTTP GET and return the response as
// unpacked message.
//
func (cl *DoHClient) Get(msg *Message) (*Message, error) {
	q := base64.RawURLEncoding.EncodeToString(msg.Packet)

	cl.query.Set("dns", q)
	cl.req.Method = http.MethodGet
	cl.req.Body = nil
	cl.req.URL.RawQuery = cl.query.Encode()

	httpRes, err := cl.conn.Do(cl.req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(httpRes.Body)
	httpRes.Body.Close()
	if err != nil {
		return nil, err
	}

	if httpRes.StatusCode != 200 {
		err = fmt.Errorf("%s", string(body))
		return nil, err
	}

	res := NewMessage()

	res.Packet = append(res.Packet[:0], body...)

	if len(res.Packet) > 20 {
		err = res.Unpack()
		if err != nil {
			return nil, err
		}
	}

	return res, err
}

//
// Query send DNS query to name server.  This is an alias to Get method.
// The addr parameter is unused.
//
func (cl *DoHClient) Query(msg *Message, ns net.Addr) (*Message, error) {
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
