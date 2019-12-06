// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
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
	"sync"
	"time"

	"github.com/shuLhan/share/lib/debug"
)

const (
	asFallback = "fallback"
	asPrimary  = "primary"
)

//
// Server defines DNS server.
//
// Services
//
// The server will listening for DNS over TLS only if certificates file is
// exist and valid.
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
// Debugging
//
// If debug.Value is set to value greater than 1, server will print each
// processed request, forward, and response.
// The debug information prefixed with single character to differentiate
// single action,
//
//	< : incoming request from client
//	> : the answer is sent to client
//	! : no answer found on cache and the query is not recursive, or
//	    response contains error code
//	^ : request is forwarded to parent name server
//      * : request is dropped from queue
//	~ : answer exist on cache but its expired
//	- : answer is pruned from caches
//	+ : new answer is added to caches
//	# : the expired answer is renewed and updated on caches
//
// Following the prefix is connection type, parent name server address,
// message ID, and question.
//
type Server struct {
	opts        *ServerOptions
	errListener chan error
	caches      *caches

	tlsConfig *tls.Config

	udp *net.UDPConn
	tcp *net.TCPListener
	dot net.Listener
	doh *http.Server

	requestq  chan *request
	primaryq  chan *request
	fallbackq chan *request

	fwLocker   sync.Mutex
	fwStoppers []chan bool
	fwGroup    *sync.WaitGroup // fwGroup maintain reference counting for all forwarders.
	fwn        int             // Number of forwarders currently running.
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
		opts:      opts,
		requestq:  make(chan *request, 512),
		primaryq:  make(chan *request, 512),
		fallbackq: make(chan *request, 512),
		fwGroup:   &sync.WaitGroup{},
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

	if len(opts.TLSCertFile) > 0 && len(opts.TLSPrivateKey) > 0 {
		cert, err := tls.LoadX509KeyPair(opts.TLSCertFile, opts.TLSPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("dns: error loading certificate: " + err.Error())
		}

		srv.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				cert,
			},
			InsecureSkipVerify: opts.TLSAllowInsecure, //nolint:gosec
		}
	}

	srv.errListener = make(chan error, 1)
	srv.caches = newCaches(opts.PruneDelay, opts.PruneThreshold)

	return srv, nil
}

//
// isResponseValid check if request name, type, and class match with response.
// It will return true if both matched, otherwise it will return false.
//
func isResponseValid(req *request, res *Message) bool {
	if !bytes.Equal(req.message.Question.Name, res.Question.Name) {
		log.Printf("dns: unmatched response name, got %s want %s\n",
			req.message.Question.Name, res.Question.Name)
		return false
	}
	if req.message.Question.Type != res.Question.Type {
		log.Printf("dns: unmatched response type, got %s want %s\n",
			req.message.Question.String(), res.Question.String())
		return false
	}
	if req.message.Question.Class != res.Question.Class {
		log.Printf("dns: unmatched response class, got %s want %s\n",
			req.message.Question.String(), res.Question.String())
		return false
	}

	return true
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
// LoadMasterFile populate caches with DNS record from master (zone) formatted
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
// RestartForwarders stop and start new forwarders with new nameserver address
// and protocol.
// Empty nameservers means server will run without forwarding request.
//
func (srv *Server) RestartForwarders(nameServers, fallbackNS []string) {
	fmt.Printf("dns: RestartForwarders: %s %s\n", nameServers, fallbackNS)

	srv.opts.NameServers = nameServers
	srv.opts.FallbackNS = fallbackNS

	srv.opts.parseNameServers()

	srv.stopForwarders()
	srv.runForwarders()
}

//
// Start the server, listening and serve query from clients.
//
func (srv *Server) Start() {
	srv.runForwarders()

	go srv.processRequest()

	go srv.serveDoT()
	go srv.serveDoH()
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
func (srv *Server) Wait() (err error) {
	err = <-srv.errListener
	if err != nil && err != io.EOF {
		log.Println(err)
	}

	srv.Stop()

	return err
}

//
// serveDoH listen for request over HTTPS using certificate and key
// file in parameter.  The path to request is static "/dns-query".
//
func (srv *Server) serveDoH() {
	var err error

	addr := srv.opts.getHTTPAddress().String()

	srv.doh = &http.Server{
		Addr:        addr,
		IdleTimeout: srv.opts.HTTPIdleTimeout,
	}

	http.Handle("/dns-query", srv)

	if srv.tlsConfig != nil && !srv.opts.DoHBehindProxy {
		log.Println("dns.Server: listening for DNS over HTTPS at", addr)
		srv.doh.TLSConfig = srv.tlsConfig
		err = srv.doh.ListenAndServeTLS("", "")
	} else {
		log.Println("dns.Server: listening for DNS over HTTP at", addr)
		err = srv.doh.ListenAndServe()
	}
	if err != io.EOF {
		err = fmt.Errorf("dns: error on DoH: " + err.Error())
	}

	srv.errListener <- err
}

func (srv *Server) serveDoT() {
	var (
		err error
	)

	if srv.tlsConfig == nil {
		return
	}

	dotAddr := srv.opts.getDoTAddress()

	for {
		srv.dot, err = tls.Listen("tcp", dotAddr.String(), srv.tlsConfig)
		if err != nil {
			log.Println("dns: Server.serveDoT: " + err.Error())
			time.Sleep(3 * time.Second)
			continue
		}

		log.Println("dns.Server: listening for DNS over TLS at", dotAddr.String())

		for {
			conn, err := srv.dot.Accept()
			if err != nil {
				if err != io.EOF {
					err = fmt.Errorf("dns: error on accepting DoT connection: " + err.Error())
				}
				srv.errListener <- err
				break
			}

			cl := &TCPClient{
				timeout: clientTimeout,
				conn:    conn,
			}

			go srv.serveTCPClient(cl, connTypeDoT)
		}
	}
}

//
// serveTCP serve DNS request from TCP connection.
//
func (srv *Server) serveTCP() {
	log.Println("dns.Server: listening for DNS over TCP at", srv.tcp.Addr())

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
			timeout: clientTimeout,
			conn:    conn,
		}

		go srv.serveTCPClient(cl, connTypeTCP)
	}
}

//
// serveUDP serve DNS request from UDP connection.
//
func (srv *Server) serveUDP() {
	log.Println("dns.Server: listening for DNS over UDP at", srv.udp.LocalAddr())

	for {
		req := newRequest()

		n, raddr, err := srv.udp.ReadFromUDP(req.message.Packet)
		if err != nil {
			if err != io.EOF {
				err = fmt.Errorf("dns: error when reading from UDP: " + err.Error())
			}
			srv.errListener <- err
			return
		}

		req.kind = connTypeUDP
		req.message.Packet = req.message.Packet[:n]
		req.writer = &UDPClient{
			timeout: clientTimeout,
			conn:    srv.udp,
			addr:    raddr,
		}

		err = req.message.UnpackHeaderQuestion()
		if err != nil {
			log.Println(err)
			req.error(RCodeErrServer)
			continue
		}

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
	req := newRequest()

	req.kind = connTypeDoH
	cl := &DoHClient{
		w:         w,
		responded: make(chan bool, 1),
	}

	req.writer = cl
	req.message.Packet = append(req.message.Packet[:0], raw...)

	err := req.message.UnpackHeaderQuestion()
	if err != nil {
		log.Println(err)
		req.error(RCodeErrServer)
		return
	}

	srv.requestq <- req

	cl.waitResponse()
}

//
// hasForwarders will return true if server run at least one forwarder,
// otherwise it will return false.
//
func (srv *Server) hasForwarders() (ok bool) {
	srv.fwLocker.Lock()
	ok = (srv.fwn > 0)
	srv.fwLocker.Unlock()
	return
}

func (srv *Server) decForwarder() {
	srv.fwLocker.Lock()
	srv.fwn--
	if srv.fwn <= 0 {
		srv.fwn = 0
	}
	srv.fwLocker.Unlock()
}

func (srv *Server) incForwarder() {
	srv.fwLocker.Lock()
	srv.fwn++
	srv.fwLocker.Unlock()
}

func (srv *Server) serveTCPClient(cl *TCPClient, kind connType) {
	for {
		req := newRequest()

		n, err := cl.recv(req.message)
		if err != nil {
			if err == io.EOF {
				goto out
			}
			continue
		}
		if n == 0 || len(req.message.Packet) == 0 {
			goto out
		}

		req.kind = kind
		req.writer = cl

		err = req.message.UnpackHeaderQuestion()
		if err != nil {
			log.Println(err)
			req.error(RCodeErrServer)
			continue
		}

		srv.requestq <- req
	}
out:
	err := cl.conn.Close()
	if err != nil {
		log.Println("serveTCPClient: conn.Close:", err)
	}
}

func (srv *Server) isImplemented(msg *Message) bool {
	switch msg.Question.Class {
	case QueryClassCS, QueryClassCH, QueryClassHS:
		log.Printf("dns: class %d is not implemented\n", msg.Question.Class)
		return false
	}

	if msg.Question.Type >= QueryTypeA && msg.Question.Type <= QueryTypeTXT {
		return true
	}
	switch msg.Question.Type {
	case QueryTypeAAAA, QueryTypeSRV, QueryTypeOPT, QueryTypeAXFR,
		QueryTypeMAILB, QueryTypeMAILA:
		return true
	}

	log.Printf("dns: type %d is not implemented\n",
		msg.Question.Type)

	return false
}

//
// processRequest from client.
//
func (srv *Server) processRequest() {
	for req := range srv.requestq {
		if !srv.isImplemented(req.message) {
			req.error(RCodeNotImplemented)
			continue
		}

		if debug.Value >= 1 {
			fmt.Printf("dns: < %s %d:%s\n",
				connTypeNames[req.kind],
				req.message.Header.ID,
				req.message.Question.String())
		}

		ans, an := srv.caches.get(string(req.message.Question.Name),
			req.message.Question.Type,
			req.message.Question.Class)

		if ans == nil || an == nil {
			if srv.hasForwarders() {
				srv.primaryq <- req
			} else {
				if debug.Value >= 1 {
					fmt.Printf("dns: * %s %d:%s\n",
						connTypeNames[req.kind],
						req.message.Header.ID,
						req.message.Question.String())
				}
				req.error(RCodeErrServer)
			}
			continue
		}

		if an.msg.IsExpired() {
			if srv.hasForwarders() {
				if debug.Value >= 1 {
					fmt.Printf("dns: ~ %s %d:%s\n",
						connTypeNames[req.kind],
						req.message.Header.ID,
						req.message.Question.String())
				}
				srv.primaryq <- req
			} else {
				if debug.Value >= 1 {
					fmt.Printf("dns: * %s %d:%s\n",
						connTypeNames[req.kind],
						req.message.Header.ID,
						req.message.Question.String())
				}
				req.error(RCodeErrServer)
			}
			continue
		}

		an.msg.SetID(req.message.Header.ID)
		an.updateTTL()
		res := an.msg

		if debug.Value >= 1 {
			fmt.Printf("dns: > %s %d:%s\n",
				connTypeNames[req.kind],
				res.Header.ID, res.Question.String())
		}

		_, err := req.writer.Write(res.Packet)
		if err != nil {
			log.Println("dns: processRequest: ", err.Error())
		}
	}
}

func (srv *Server) processResponse(req *request, res *Message) {
	if !isResponseValid(req, res) {
		req.error(RCodeErrServer)
		return
	}

	_, err := req.writer.Write(res.Packet)
	if err != nil {
		log.Println("dns: processResponse: ", err.Error())
		return
	}

	if res.Header.RCode != 0 {
		log.Printf("dns: ! %s %s %d:%s\n",
			connTypeNames[req.kind], rcodeNames[res.Header.RCode],
			res.Header.ID, res.Question.String())
		return
	}

	an := newAnswer(res, false)
	inserted := srv.caches.upsert(an)

	if debug.Value >= 1 {
		if inserted {
			fmt.Printf("dns: + %s %d:%s\n",
				connTypeNames[req.kind],
				res.Header.ID, res.Question.String())
		} else {
			fmt.Printf("dns: # %s %d:%s\n",
				connTypeNames[req.kind],
				res.Header.ID, res.Question.String())
		}
	}
}

func (srv *Server) runForwarders() {
	srv.fwStoppers = nil

	nforwarders := 0
	for x := 0; x < len(srv.opts.primaryUDP); x++ {
		tag := fmt.Sprintf("UDP-%d-%s", nforwarders, asPrimary)
		go srv.runUDPForwarder(srv.opts.primaryUDP[x].String(), srv.primaryq, srv.fallbackq, tag)
		nforwarders++
	}
	for x := 0; x < len(srv.opts.primaryTCP); x++ {
		tag := fmt.Sprintf("TCP-%d-%s", nforwarders, asPrimary)
		go srv.runTCPForwarder(srv.opts.primaryTCP[x].String(), srv.primaryq, srv.fallbackq, tag)
		nforwarders++
	}
	for x := 0; x < len(srv.opts.primaryDoh); x++ {
		tag := fmt.Sprintf("DoH-%d-%s", nforwarders, asPrimary)
		go srv.runDohForwarder(srv.opts.primaryDoh[x], srv.primaryq, srv.fallbackq, tag)
		nforwarders++
	}
	for x := 0; x < len(srv.opts.primaryDot); x++ {
		tag := fmt.Sprintf("DoT-%d-%s", nforwarders, asPrimary)
		go srv.runTLSForwarder(srv.opts.primaryDot[x], srv.primaryq, srv.fallbackq, tag)
		nforwarders++
	}

	nforwarders = 0

	for x := 0; x < len(srv.opts.fallbackUDP); x++ {
		tag := fmt.Sprintf("UDP-%d-%s", nforwarders, asFallback)
		go srv.runUDPForwarder(srv.opts.fallbackUDP[x].String(), srv.fallbackq, nil, tag)
		nforwarders++
	}
	for x := 0; x < len(srv.opts.fallbackTCP); x++ {
		tag := fmt.Sprintf("TCP-%d-%s", nforwarders, asFallback)
		go srv.runTCPForwarder(srv.opts.fallbackTCP[x].String(), srv.fallbackq, nil, tag)
		nforwarders++
	}
	for x := 0; x < len(srv.opts.fallbackDoh); x++ {
		tag := fmt.Sprintf("DoH-%d-%s", nforwarders, asFallback)
		go srv.runDohForwarder(srv.opts.fallbackDoh[x], srv.fallbackq, nil, tag)
		nforwarders++
	}
	for x := 0; x < len(srv.opts.fallbackDot); x++ {
		tag := fmt.Sprintf("DoT-%d-%s", nforwarders, asFallback)
		go srv.runTLSForwarder(srv.opts.fallbackDot[x], srv.fallbackq, nil, tag)
		nforwarders++
	}
}

func (srv *Server) runDohForwarder(nameserver string, primaryq, fallbackq chan *request, tag string) {
	var (
		res       *Message
		isRunning = true
		stopper   = srv.newStopper()
	)

	for isRunning {
		forwarder, err := NewDoHClient(nameserver, false)
		if err != nil {
			log.Printf("dns: failed to create forwarder %s: %s\n",
				tag, err.Error())

			select {
			case <-stopper:
				isRunning = false
				goto out
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf("dns: forwarder %s for %s has been connected ...\n", tag, nameserver)

		if fallbackq != nil {
			srv.incForwarder()
		}

		for err == nil {
			select {
			case req, ok := <-primaryq:
				if !ok {
					log.Println("dns: primary queue has been closed")
					isRunning = false
					goto out
				}
				if debug.Value >= 1 {
					fmt.Printf("dns: ^ %s %s %d:%s\n",
						tag, nameserver,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf("dns: %s forward failed: %s\n", tag, err.Error())
					if fallbackq != nil {
						fallbackq <- req
					}
				} else {
					srv.processResponse(req, res)
				}
			case <-stopper:
				isRunning = false
				goto out
			}
		}

		log.Printf("dns: reconnect forwarder %s for %s\n", tag, nameserver)
	out:
		if forwarder != nil {
			forwarder.Close()
		}
		if fallbackq != nil {
			srv.decForwarder()
		}
	}
	srv.fwGroup.Done()
	log.Printf("dns: forwarder %s for %s has been stopped\n", tag, nameserver)
}

func (srv *Server) runTLSForwarder(nameserver string, primaryq, fallbackq chan *request, tag string) {
	var (
		res       *Message
		isRunning = true
		stopper   = srv.newStopper()
	)

	for isRunning {
		forwarder, err := NewDoTClient(nameserver, srv.opts.TLSAllowInsecure)
		if err != nil {
			log.Printf("dns: failed to create forwarder %s: %s\n",
				tag, err.Error())

			select {
			case <-stopper:
				isRunning = false
				goto out
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf("dns: forwarder %s for %s has been connected ...\n", tag, nameserver)

		if fallbackq != nil {
			srv.incForwarder()
		}

		for err == nil {
			select {
			case req, ok := <-primaryq:
				if !ok {
					log.Println("dns: primary queue has been closed")
					isRunning = false
					goto out
				}
				if debug.Value >= 1 {
					fmt.Printf("dns: ^ %s %s %d:%s\n",
						tag, nameserver,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf("dns: %s forward failed: %s\n", tag, err.Error())
					if fallbackq != nil {
						fallbackq <- req
					}
				} else {
					srv.processResponse(req, res)
				}
			case <-stopper:
				isRunning = false
				goto out
			}
		}

		log.Printf("dns: reconnect forwarder %s for %s\n", tag, nameserver)
	out:
		if forwarder != nil {
			forwarder.Close()
		}
		if fallbackq != nil {
			srv.decForwarder()
		}
	}
	srv.fwGroup.Done()
	log.Printf("dns: forwarder %s for %s has been stopped\n", tag, nameserver)
}

func (srv *Server) runTCPForwarder(remoteAddr string, primaryq, fallbackq chan *request, tag string) {
	stopper := srv.newStopper()

	log.Printf("dns: starting forwarder %s for %s\n", tag, remoteAddr)

	if fallbackq != nil {
		srv.incForwarder()
	}

	for {
		select {
		case req, ok := <-primaryq:
			if !ok {
				log.Println("dns: primary queue has been closed")
				goto out
			}
			if debug.Value >= 1 {
				fmt.Printf("dns: ^ %s %s %d:%s\n",
					tag, remoteAddr,
					req.message.Header.ID,
					req.message.Question.String())
			}

			cl, err := NewTCPClient(remoteAddr)
			if err != nil {
				log.Printf("dns: failed to create forwarder %s: %s\n", tag, err.Error())
				err = nil
				continue
			}

			res, err := cl.Query(req.message)
			cl.Close()
			if err != nil {
				log.Printf("dns: %s forward failed: %s\n", tag, err.Error())
				if fallbackq != nil {
					fallbackq <- req
				}
				continue
			}

			srv.processResponse(req, res)
		case <-stopper:
			goto out
		}
	}
out:
	if fallbackq != nil {
		srv.decForwarder()
	}
	srv.fwGroup.Done()
	log.Printf("dns: forwarder %s for %s has been stopped\n", tag, remoteAddr)
}

//
// runUDPForwarder create a UDP client that consume request from forward queue
// and forward it to parent server at "remoteAddr".
//
func (srv *Server) runUDPForwarder(remoteAddr string, primaryq, fallbackq chan *request, tag string) {
	var (
		res       *Message
		isRunning = true
		stopper   = srv.newStopper()
	)

	// The first loop handle broken connection.
	for isRunning {
		forwarder, err := NewUDPClient(remoteAddr)
		if err != nil {
			log.Printf("dns: failed to create forwarder %s: %s\n",
				tag, err.Error())

			select {
			case <-stopper:
				isRunning = false
				goto out
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf("dns: forwarder %s for %s has been connected ...\n", tag, remoteAddr)

		if fallbackq != nil {
			srv.incForwarder()
		}

		// The second loop consume the forward queue.
		for err == nil {
			select {
			case req, ok := <-primaryq:
				if !ok {
					log.Println("dns: primary queue has been closed")
					isRunning = false
					break
				}
				if debug.Value >= 1 {
					fmt.Printf("dns: ^ %s %s %d:%s\n",
						tag, remoteAddr,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf("dns: %s forward failed: %s\n", tag, err.Error())
					if fallbackq != nil {
						fallbackq <- req
					}
				} else {
					srv.processResponse(req, res)
				}
			case <-stopper:
				isRunning = false
				goto out
			}
		}

		log.Printf("dns: reconnect forwarder %s for %s\n", tag, remoteAddr)
	out:
		if forwarder != nil {
			forwarder.Close()
		}
		if fallbackq != nil {
			srv.decForwarder()
		}
	}
	srv.fwGroup.Done()
	log.Printf("dns: forwarder %s for %s has been stopped\n", tag, remoteAddr)
}

//
// stopForwarders stop all forwarder connections.
//
func (srv *Server) stopForwarders() {
	for _, stopper := range srv.fwStoppers {
		stopper <- true
	}
	srv.fwGroup.Wait()

	srv.fwLocker.Lock()
	srv.fwn = 0
	srv.fwLocker.Unlock()

	fmt.Println("dns: all forwarders has been stopped")
}

func (srv *Server) newStopper() (stopper chan bool) {
	srv.fwLocker.Lock()

	stopper = make(chan bool, 1)
	srv.fwStoppers = append(srv.fwStoppers, stopper)
	srv.fwGroup.Add(1)

	srv.fwLocker.Unlock()

	return
}
