// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

//
// Server defines DNS server.
//
type Server struct {
	Handler     Handler
	udp         *net.UDPConn
	tcp         *net.TCPListener
	doh         *http.Server
	errListener chan error
}

func (srv *Server) init(opts *ServerOptions) (err error) {
	err = opts.init()
	if err != nil {
		return
	}

	udpAddr := opts.getUDPAddress()
	srv.udp, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("dns: error listening on UDP '%v': %s",
			udpAddr, err.Error())
	}

	tcpAddr := opts.getTCPAddress()
	srv.tcp, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return fmt.Errorf("dns: error listening on TCP '%v': %s",
			tcpAddr, err.Error())
	}

	srv.errListener = make(chan error, 1)

	return nil
}

//
// Start the server, listening and serve query from clients.
//
func (srv *Server) Start(opts *ServerOptions) (err error) {
	err = srv.init(opts)
	if err != nil {
		return
	}

	if opts.cert != nil {
		dohAddress := opts.getDoHAddress()
		go srv.serveDoH(dohAddress, opts.DoHIdleTimeout, *opts.cert,
			opts.DoHAllowInsecure)
		opts.cert = nil
	}

	go srv.serveTCP()
	go srv.serveUDP()

	return nil
}

//
// Stop the server, close all listeners.
//
func (srv *Server) Stop() {
	err := srv.udp.Close()
	if err != nil {
		log.Println("dns: error when closing UDP: " + err.Error())
	}
	err = srv.tcp.Close()
	if err != nil {
		log.Println("dns: error when closing TCP: " + err.Error())
	}
	err = srv.doh.Close()
	if err != nil {
		log.Println("dns: error when closing DoH: " + err.Error())
	}
}

//
// Wait for server to be Stop()-ed or when one of listener throw an error.
//
func (srv *Server) Wait() {
	err := <-srv.errListener
	if err != nil && err != io.EOF {
		log.Println(err)
	}

	srv.Stop()
}

//
// serveDoH listen for request over HTTPS using certificate and key
// file in parameter.  The path to request is static "/dns-query".
//
func (srv *Server) serveDoH(addr *net.TCPAddr, idleTimeout time.Duration,
	cert tls.Certificate, allowInsecure bool,
) {
	srv.doh = &http.Server{
		Addr:        addr.String(),
		IdleTimeout: idleTimeout,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{
				cert,
			},
			InsecureSkipVerify: allowInsecure, // nolint: gosec
		},
	}

	http.Handle("/dns-query", srv)

	err := srv.doh.ListenAndServeTLS("", "")
	if err != io.EOF {
		err = fmt.Errorf("dns: error on DoH: " + err.Error())
	}

	srv.errListener <- err
}

//
// serveTCP serve DNS request from TCP connection.
//
func (srv *Server) serveTCP() {
	for {
		conn, err := srv.tcp.AcceptTCP()
		if err != nil {
			if err != io.EOF {
				err = fmt.Errorf("dns: error on accepting TCP connection: " + err.Error())
			}
			srv.errListener <- err
			return
		}

		cl := &TCPClient{
			Timeout: clientTimeout,
			conn:    conn,
		}

		go srv.serveTCPClient(cl)
	}
}

//
// serveUDP serve DNS request from UDP connection.
//
func (srv *Server) serveUDP() {
	sender := &UDPClient{
		Timeout: clientTimeout,
		Conn:    srv.udp,
	}

	for {
		req := NewRequest()

		n, raddr, err := srv.udp.ReadFromUDP(req.Message.Packet)
		if err != nil {
			if err != io.EOF {
				err = fmt.Errorf("dns: error when reading from UDP: " + err.Error())
			}
			srv.errListener <- err
			return
		}

		req.Kind = ConnTypeUDP
		req.UDPAddr = raddr
		req.Message.Packet = req.Message.Packet[:n]

		req.Message.UnpackHeaderQuestion()
		req.Sender = sender

		srv.Handler.ServeDNS(req)
	}
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
	req := NewRequest()

	req.Kind = ConnTypeDoH
	req.ResponseWriter = w
	req.ChanResponded = make(chan bool, 1)

	req.Message.Packet = append(req.Message.Packet[:0], raw...)
	req.Message.UnpackHeaderQuestion()

	srv.Handler.ServeDNS(req)

	_, ok := <-req.ChanResponded
	if !ok {
		w.WriteHeader(http.StatusGatewayTimeout)
	}
}

func (srv *Server) serveTCPClient(cl *TCPClient) {
	var (
		n   int
		err error
	)
	for {
		req := NewRequest()
		for {
			n, err = cl.Recv(req.Message)
			if err == nil {
				break
			}
			if err == io.EOF {
				break
			}
			if n != 0 {
				log.Println("serveTCPClient:", err)
				req.Message.Reset()
			}
			continue
		}
		if err == io.EOF {
			break
		}

		req.Kind = ConnTypeTCP
		req.Message.UnpackHeaderQuestion()
		req.Sender = cl

		srv.Handler.ServeDNS(req)
	}

	err = cl.conn.Close()
	if err != nil {
		log.Println("serveTCPClient: conn.Close:", err)
	}
}
