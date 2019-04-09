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
	"os"
	"path/filepath"
	"strings"
)

//
// Server defines DNS server.
//
// Caches
//
// There are two type of answer: local and non-local.
// Local answer is a DNS record that is loaded from hosts file or master
// zone file.
// Non-local answer is a DNS record that is received from parent name
// servers.
//
// Server caches the DNS answers in two storages: map and list.
// The map caches store local and non local answers, using domain name as a
// key and list of answers as value,
//
//	domain-name -> [{A,IN,...},{AAAA,IN,...}]
//
// The list caches store non-local answers, ordered by last accessed time,
// it is used to prune least frequently accessed answers.
// Local caches will never get pruned.
//
type Server struct {
	opts        *ServerOptions
	errListener chan error
	caches      *caches

	udp *net.UDPConn
	tcp *net.TCPListener
	doh *http.Server

	requestq chan *Request
}

//
// NewServer create and initialize server using the options and a .handler.
//
func NewServer(opts *ServerOptions) (srv *Server, err error) {
	err = opts.init()
	if err != nil {
		return nil, err
	}

	srv = &Server{
		opts:     opts,
		requestq: make(chan *Request, 512),
	}

	udpAddr := opts.getUDPAddress()
	srv.udp, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("dns: error listening on UDP '%v': %s",
			udpAddr, err.Error())
	}

	tcpAddr := opts.getTCPAddress()
	srv.tcp, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("dns: error listening on TCP '%v': %s",
			tcpAddr, err.Error())
	}

	srv.errListener = make(chan error, 1)
	srv.caches = newCaches(opts.PruneDelay, opts.PruneThreshold)

	return srv, nil
}

//
// LoadHostsDir populate caches with DNS record from hosts formatted files in
// directory "dir".
//
func (srv *Server) LoadHostsDir(dir string) {
	if len(dir) == 0 {
		return
	}

	d, err := os.Open(dir)
	if err != nil {
		log.Println("dns: LoadHostsDir: ", err)
		return
	}

	fis, err := d.Readdir(0)
	if err != nil {
		log.Println("dns: LoadHostsDir: ", err)
		err = d.Close()
		if err != nil {
			log.Println("dns: LoadHostsDir: ", err)
		}
		return
	}

	for x := 0; x < len(fis); x++ {
		if fis[x].IsDir() {
			continue
		}

		hostsFile := filepath.Join(dir, fis[x].Name())

		srv.LoadHostsFile(hostsFile)
	}

	err = d.Close()
	if err != nil {
		log.Println("dns: LoadHostsDir: ", err)
	}
}

//
// LoadHostsFile populate caches with DNS record from hosts formatted file.
//
func (srv *Server) LoadHostsFile(path string) {
	if len(path) == 0 {
		fmt.Println("dns: loading system hosts file")
	} else {
		fmt.Printf("dns: loading hosts file '%s'\n", path)
	}

	msgs, err := HostsLoad(path)
	if err != nil {
		log.Println("dns: LoadHostsFile: " + err.Error())
	}

	srv.populateCaches(msgs)
}

//
// LoadMasterDir populate caches with DNS record from master (zone) formatted
// files in directory "dir".
//
func (srv *Server) LoadMasterDir(dir string) {
	if len(dir) == 0 {
		return
	}

	d, err := os.Open(dir)
	if err != nil {
		log.Println("dns: LoadMasterDir: ", err)
		return
	}

	fis, err := d.Readdir(0)
	if err != nil {
		log.Println("dns: LoadMasterDir: ", err)
		err = d.Close()
		if err != nil {
			log.Println("dns: LoadMasterDir: ", err)
		}
		return
	}

	for x := 0; x < len(fis); x++ {
		if fis[x].IsDir() {
			continue
		}

		masterFile := filepath.Join(dir, fis[x].Name())

		srv.LoadMasterFile(masterFile)
	}

	err = d.Close()
	if err != nil {
		log.Println("dns: LoadMasterDir: error closing directory:", err)
	}
}

//
// LostMasterFile populate caches with DNS record from master (zone) formatted
// file.
//
func (srv *Server) LoadMasterFile(path string) {
	fmt.Printf("dns: loading master file '%s'\n", path)

	msgs, err := MasterLoad(path, "", 0)
	if err != nil {
		log.Println("dns: LoadMasterFile: " + err.Error())
	}

	srv.populateCaches(msgs)
}

//
// populateCaches add list of message to caches.
//
func (srv *Server) populateCaches(msgs []*Message) {
	var (
		n        int
		inserted bool
		isLocal  = true
	)

	for x := 0; x < len(msgs); x++ {
		an := newAnswer(msgs[x], isLocal)
		inserted = srv.caches.upsert(an)
		if inserted {
			n++
		}
		msgs[x] = nil
	}

	fmt.Printf("dns: %d out of %d records cached\n", n, len(msgs))
}

//
// Start the server, listening and serve query from clients.
//
func (srv *Server) Start() {
	go srv.processRequest()

	if srv.opts.DoHCertificate != nil {
		go srv.serveDoH()
	}

	go srv.serveTCP()
	go srv.serveUDP()
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

	close(srv.requestq)
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
func (srv *Server) serveDoH() {
	srv.doh = &http.Server{
		Addr:        srv.opts.getDoHAddress().String(),
		IdleTimeout: srv.opts.DoHIdleTimeout,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{
				*srv.opts.DoHCertificate,
			},
			InsecureSkipVerify: srv.opts.DoHAllowInsecure, // nolint: gosec
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

		srv.requestq <- req
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

	srv.requestq <- req

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

		srv.requestq <- req
	}

	err = cl.conn.Close()
	if err != nil {
		log.Println("serveTCPClient: conn.Close:", err)
	}
}

//
// processRequest from client.
//
func (srv *Server) processRequest() {
	var (
		resp []byte
	)

	for req := range srv.requestq {
		ans, an := srv.caches.get(string(req.Message.Question.Name),
			req.Message.Question.Type,
			req.Message.Question.Class)

		if ans == nil {
			req.Message.SetResponseCode(RCodeErrName)
		}
		if an == nil {
			req.Message.SetQuery(false)
			req.Message.SetAuthorativeAnswer(true)
			resp = req.Message.Packet
		} else {
			an.msg.SetID(req.Message.Header.ID)
			resp = an.get()
		}

		switch req.Kind {
		case ConnTypeUDP, ConnTypeTCP:
			if req.Sender != nil {
				_, err := req.Sender.Send(resp, req.UDPAddr)
				if err != nil {
					log.Println("dns: processRequest: Sender.Send:" + err.Error())
				}
			}

		case ConnTypeDoH:
			if req.ResponseWriter != nil {
				_, err := req.ResponseWriter.Write(resp)
				if err != nil {
					log.Println("dns: processRequest: ResponseWriter.Write:", err)
				}
				req.ChanResponded <- true
			}
		}
	}
}
