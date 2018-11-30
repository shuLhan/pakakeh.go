// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"crypto/tls"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
)

//
// Server defines DNS server.
//
type Server struct {
	Handler Handler
	udp     *net.UDPConn
	tcp     *net.TCPListener
	doh     *http.Server
}

//
// ListenAndServe run DNS server on UDP, TCP, or DNS over HTTP (DoH).
//
func (srv *Server) ListenAndServe(opts *ServerOptions) error {
	err := opts.parse()
	if err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		err = srv.ListenAndServeTCP(opts.getTCPAddress())
		if err != nil {
			cherr <- err
		}
	}()

	go func() {
		err = srv.ListenAndServeUDP(opts.getUDPAddress())
		if err != nil {
			cherr <- err
		}
	}()

	if len(opts.DoHCert) > 0 && len(opts.DoHCertKey) > 0 {
		go func() {
			err = srv.ListenAndServeDoH(opts)
			if err != nil {
				cherr <- err
			}
		}()
	}

	err = <-cherr

	return err
}

//
// ListenAndServeDoH listen for request over HTTPS using certificate and key
// file in parameter.  The path to request is static "/dns-query".
//
func (srv *Server) ListenAndServeDoH(opts *ServerOptions) error {
	if opts.ip == nil {
		err := opts.parse()
		if err != nil {
			return err
		}
	}

	srv.doh = &http.Server{
		Addr:        opts.getDoHAddress().String(),
		IdleTimeout: opts.DoHIdleTimeout,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: opts.DoHAllowInsecure, // nolint: gosec
		},
	}

	http.Handle("/dns-query", srv)

	return srv.doh.ListenAndServeTLS(opts.DoHCert, opts.DoHCertKey)
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hdr := w.Header()
	hdr.Set(dohHeaderKeyContentType, dohHeaderValDNSMessage)

	hdrAcceptValue := r.Header.Get(dohHeaderKeyAccept)
	if len(hdrAcceptValue) == 0 {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	hdrAcceptValue = strings.ToLower(hdrAcceptValue)
	if hdrAcceptValue != dohHeaderValDNSMessage {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	if r.Method == http.MethodGet {
		srv.handleDoHGet(w, r)
		return
	}
	if r.Method == http.MethodPost {
		srv.handleDoHPost(w, r)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (srv *Server) handleDoHGet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	msgBase64 := q.Get("dns")

	if len(msgBase64) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	raw, err := base64.RawURLEncoding.DecodeString(msgBase64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srv.handleDoHRequest(raw, w)
}

func (srv *Server) handleDoHPost(w http.ResponseWriter, r *http.Request) {
	raw, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srv.handleDoHRequest(raw, w)
}

func (srv *Server) handleDoHRequest(raw []byte, w http.ResponseWriter) {
	req := AllocRequest()

	req.Kind = ConnTypeDoH
	req.ResponseWriter = w
	req.ChanResponded = make(chan bool, 1)

	req.Message.Packet = append(req.Message.Packet[:0], raw...)
	req.Message.UnpackHeaderQuestion()

	srv.Handler.ServeDNS(req)

	_, ok := <-req.ChanResponded
	if ok {
		FreeRequest(req)
	} else {
		w.WriteHeader(http.StatusGatewayTimeout)
	}
}

//
// ListenAndServeTCP listen for request with TCP socket.
//
func (srv *Server) ListenAndServeTCP(tcpAddr *net.TCPAddr) error {
	var err error

	srv.tcp, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	for {
		conn, err := srv.tcp.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		cl := &TCPClient{
			Timeout: clientTimeout,
			conn:    conn,
		}

		go srv.serveTCPClient(cl)
	}
}

//
// ListenAndServeUDP listen for request with UDP socket.
//
func (srv *Server) ListenAndServeUDP(udpAddr *net.UDPAddr) error {
	var (
		n   int
		err error
		req *Request
	)

	srv.udp, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	sender := &UDPClient{
		Timeout: clientTimeout,
		conn:    srv.udp,
	}

	for {
		if req == nil {
			req = AllocRequest()
		}

		n, req.UDPAddr, err = srv.udp.ReadFromUDP(req.Message.Packet)
		if err != nil {
			log.Println(err)
			continue
		}

		req.Kind = ConnTypeUDP
		req.Message.Packet = req.Message.Packet[:n]

		req.Message.UnpackHeaderQuestion()
		req.Sender = sender

		srv.Handler.ServeDNS(req)
		req = nil
	}
}

func (srv *Server) serveTCPClient(cl *TCPClient) {
	var (
		n   int
		err error
		req *Request
	)
	for {
		if req == nil {
			req = AllocRequest()
		}

		for {
			n, err = cl.Recv(req.Message)
			if err == nil {
				break
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				if n != 0 {
					log.Println("serveTCPClient:", err)
					req.Message.Reset()
				}
				continue
			}
		}
		if err == io.EOF {
			break
		}

		req.Kind = ConnTypeTCP
		req.Message.UnpackHeaderQuestion()
		req.Sender = cl

		srv.Handler.ServeDNS(req)
		req = nil
	}

	err = cl.conn.Close()
	if err != nil {
		log.Println("serveTCPClient: conn.Close:", err)
	}
}
