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
	chRes   chan *http.Response
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
		chRes: make(chan *http.Response, 1),
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
	cl.conn.CloseIdleConnections()
	return nil
}

func (cl *DoHClient) Lookup(qtype, qclass uint16, qname []byte) (*Message, error) {
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
	if err != nil {
		httpRes.Body.Close()
		return nil, err
	}

	res.Packet = append(res.Packet[:0], packet...)

	httpRes.Body.Close()

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

	if httpRes.StatusCode != 200 {
		body, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			return nil, err
		}
		err = fmt.Errorf("%s", string(body))
		return nil, err
	}

	res := NewMessage()

	packet, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		httpRes.Body.Close()
		return nil, err
	}

	res.Packet = append(res.Packet[:0], packet...)

	httpRes.Body.Close()

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
// Recv read response from channel.
//
func (cl *DoHClient) Recv(msg *Message) (int, error) {
	httpRes := <-cl.chRes

	if httpRes.StatusCode != 200 {
		body, err := ioutil.ReadAll(httpRes.Body)
		if err != nil {
			return 0, err
		}
		err = fmt.Errorf("%s", string(body))
		return 0, err
	}

	packet, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		httpRes.Body.Close()
		return 0, err
	}

	msg.Packet = append(msg.Packet[:0], packet...)

	httpRes.Body.Close()

	if len(msg.Packet) > 20 {
		err = msg.Unpack()
		if err != nil {
			return 0, err
		}
	}

	return len(msg.Packet), nil
}

//
// RemoteAddr return client remote nameserver address.
//
func (cl *DoHClient) RemoteAddr() string {
	return cl.addr.String()
}

//
// Send DNS message to name server using Get method.  Since HTTP client is
// synchronous, the response is forwarded to channel to be consumed by Recv().
//
func (cl *DoHClient) Send(msg *Message, ns net.Addr) (int, error) {
	packet := base64.RawURLEncoding.EncodeToString(msg.Packet)

	cl.query.Set("dns", packet)
	cl.req.Method = http.MethodGet
	cl.req.Body = nil
	cl.req.URL.RawQuery = cl.query.Encode()

	httpRes, err := cl.conn.Do(cl.req)
	if err != nil {
		return 0, err
	}

	cl.chRes <- httpRes

	return len(msg.Packet), nil
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
