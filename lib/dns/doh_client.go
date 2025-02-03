// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dns

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// DoHClient client for DNS over HTTPS.
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

// NewDoHClient will create new DNS client with HTTP connection.
func NewDoHClient(nameserver string, allowInsecure bool) (cl *DoHClient, err error) {
	var (
		nsURL *url.URL
		tr    *http.Transport
	)

	nsURL, err = url.Parse(nameserver)
	if err != nil {
		return nil, err
	}

	if nsURL.Scheme != "https" {
		return nil, errors.New(`DoH name server must be HTTPS`)
	}

	tr = &http.Transport{
		MaxIdleConns:    1,
		IdleConnTimeout: 30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: allowInsecure,
		},
	}

	cl = &DoHClient{
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

// Close all idle connections.
func (cl *DoHClient) Close() error {
	cl.conn.Transport.(*http.Transport).CloseIdleConnections()
	return nil
}

// Lookup DNS records based on MessageQuestion Name and Type, in synchronous
// mode.
// The MessageQuestion Class default to IN.
//
// It will return an error if the Name is empty.
func (cl *DoHClient) Lookup(q MessageQuestion, allowRecursion bool) (res *Message, err error) {
	if len(q.Name) == 0 {
		return nil, errors.New(`Lookup: empty question name`)
	}
	if q.Type == 0 {
		q.Type = RecordTypeA
	}
	if q.Class == 0 {
		q.Class = RecordClassIN
	}

	var msg = NewMessage()

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

// Post send query to name server using HTTP POST and return the response
// as unpacked message.
func (cl *DoHClient) Post(msg *Message) (res *Message, err error) {
	var (
		logp = `Post`

		httpRes *http.Response
	)

	cl.req.Method = http.MethodPost
	cl.req.Body = io.NopCloser(bytes.NewReader(msg.packet))
	cl.req.URL.RawQuery = ""

	httpRes, err = cl.conn.Do(cl.req)
	if err != nil {
		cl.req.Body.Close()
		return nil, err
	}
	cl.req.Body.Close()

	var packet []byte

	packet, err = io.ReadAll(httpRes.Body)
	httpRes.Body.Close()
	if err != nil {
		return nil, err
	}

	res, err = UnpackMessage(packet)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return res, nil
}

// Get send query to name server using HTTP GET and return the response as
// unpacked message.
func (cl *DoHClient) Get(msg *Message) (res *Message, err error) {
	var (
		logp = `Get`
		q    = base64.RawURLEncoding.EncodeToString(msg.packet)
	)

	cl.query.Set("dns", q)
	cl.req.Method = http.MethodGet
	cl.req.Body = nil
	cl.req.URL.RawQuery = cl.query.Encode()

	var httpRes *http.Response

	httpRes, err = cl.conn.Do(cl.req)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var packet []byte

	packet, err = io.ReadAll(httpRes.Body)
	httpRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	if httpRes.StatusCode != 200 {
		return nil, fmt.Errorf(`%s: %s`, logp, string(packet))
	}

	if len(packet) > 20 {
		res, err = UnpackMessage(packet)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	return res, nil
}

// Query send DNS query to name server.  This is an alias to Get method, to
// make it consistent with other DNS clients.
func (cl *DoHClient) Query(msg *Message) (*Message, error) {
	return cl.Get(msg)
}

// RemoteAddr return client remote nameserver address.
func (cl *DoHClient) RemoteAddr() string {
	return cl.addr.String()
}

// SetRemoteAddr set the remote address for sending the packet.
func (cl *DoHClient) SetRemoteAddr(addr string) (err error) {
	cl.addr, err = url.Parse(addr)
	if err != nil {
		return
	}

	cl.query = cl.addr.Query()

	return
}

// SetTimeout set the timeout for sending and receiving packet.
func (cl *DoHClient) SetTimeout(t time.Duration) {
	cl.conn.Timeout = t
}

// Write the raw DNS response message to active connection.
// This method is only used by server to write the response of query to
// client.
func (cl *DoHClient) Write(packet []byte) (n int, err error) {
	n, err = cl.w.Write(packet)
	if err != nil {
		cl.responded <- false
		return
	}
	cl.responded <- true
	return
}

// waitResponse wait for http.ResponseWriter being called by server.
// This method is to prevent the function that process the HTTP request
// terminated and write empty response.
func (cl *DoHClient) waitResponse() {
	var (
		success bool
		ok      bool
	)

	success, ok = <-cl.responded
	if !success || !ok {
		cl.w.WriteHeader(http.StatusGatewayTimeout)
	}
}
