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
	"time"

	libnet "github.com/shuLhan/share/lib/net"
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

func parseAddress(address string) (udp *net.UDPAddr, tcp, doh *net.TCPAddr, err error) {
	ip, port, err := libnet.ParseIPPort(address, DefaultPort)
	if err != nil {
		return
	}

	udp = &net.UDPAddr{
		IP:   ip,
		Port: int(port),
	}

	tcp = &net.TCPAddr{
		IP:   ip,
		Port: int(port),
	}

	doh = &net.TCPAddr{
		IP:   ip,
		Port: 8443,
	}

	return
}

//
// ListenAndServe run DNS server, listening on UDP and TCP connection.
//
func (srv *Server) ListenAndServe(address, certFile, keyFile string, allowInsecure bool) error {
	var err error

	udpAddr, tcpAddr, dohAddr, err := parseAddress(address)
	if err != nil {
		return err
	}

	cherr := make(chan error, 1)

	go func() {
		err = srv.ListenAndServeTCP(tcpAddr)
		if err != nil {
			cherr <- err
		}
	}()
	go func() {
		err = srv.ListenAndServeUDP(udpAddr)
		if err != nil {
			cherr <- err
		}
	}()
	if len(certFile) > 0 && len(keyFile) > 0 {
		go func() {
			err = srv.ListenAndServeDoH(dohAddr, certFile, keyFile, allowInsecure)
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
func (srv *Server) ListenAndServeDoH(address *net.TCPAddr, certFile, keyFile string, allowInsecure bool) error {
	srv.doh = &http.Server{
		Addr:        address.String(),
		IdleTimeout: 120 * time.Second,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: allowInsecure,
		},
	}

	http.Handle("/dns-query", srv)

	err := srv.doh.ListenAndServeTLS(certFile, keyFile)

	return err
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
	req := _requestPool.Get().(*Request)
	req.Reset()
	req.ChanMessage = make(chan *Message, 1)
	req.Message.Packet = append(req.Message.Packet[:0], raw...)
	req.Message.UnpackHeaderQuestion()

	srv.Handler.ServeDNS(req)

	timeout := time.NewTicker(clientTimeout)
	for {
		select {
		case res := <-req.ChanMessage:
			_, err := w.Write(res.Packet)
			if err != nil {
				log.Printf("! handleDoHRequest: %s\n", err)
			}
			goto out

		case <-timeout.C:
			w.WriteHeader(http.StatusGatewayTimeout)
			goto out
		}
	}
out:
	timeout.Stop()
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
			req = _requestPool.Get().(*Request)
		}
		req.Reset()

		n, req.UDPAddr, err = srv.udp.ReadFromUDP(req.Message.Packet)
		if err != nil {
			log.Println(err)
			continue
		}

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
			req = _requestPool.Get().(*Request)
		}
		req.Reset()

		for {
			n, err = cl.Recv(req.Message)
			if err == nil {
				break
			}
			if err != nil {
				if err == io.EOF {
					_requestPool.Put(req)
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

//
// FreeRequest put the request back to the pool.
//
func (srv *Server) FreeRequest(req *Request) {
	_requestPool.Put(req)
}
