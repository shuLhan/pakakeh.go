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

	requestq        chan *request
	primaryq        chan *request
	fallbackq       chan *request
	forwardStoppers []chan bool

	// fwGroup maintain reference counting for all forwarders.
	fwGroup *sync.WaitGroup

	hasForwarders bool
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
			req.message.Question, res.Question)
		return false
	}
	if req.message.Question.Class != res.Question.Class {
		log.Printf("dns: unmatched response class, got %s want %s\n",
			req.message.Question, res.Question)
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

	if !srv.hasForwarders {
		log.Println("dns: no valid forward nameservers")
	}
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
	var err error

	addr := srv.opts.getHTTPAddress().String()

	srv.doh = &http.Server{
		Addr:        addr,
		IdleTimeout: srv.opts.HTTPIdleTimeout,
		TLSConfig:   srv.tlsConfig,
	}

	http.Handle("/dns-query", srv)

	if srv.tlsConfig != nil {
		log.Println("dns.Server: listening for DNS over HTTPS at", addr)
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

			go srv.serveTCPClient(cl)
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

		go srv.serveTCPClient(cl)
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

		req.message.UnpackHeaderQuestion()
		req.writer = &UDPClient{
			timeout: clientTimeout,
			conn:    srv.udp,
			addr:    raddr,
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
	req.message.UnpackHeaderQuestion()

	srv.requestq <- req

	cl.waitResponse()
}

func (srv *Server) serveTCPClient(cl *TCPClient) {
	var (
		n   int
		err error
	)
	for {
		req := newRequest()
		for {
			n, err = cl.recv(req.message)
			if err == nil {
				break
			}
			if err == io.EOF {
				break
			}
			if n != 0 {
				log.Println("serveTCPClient:", err)
				req.message.Reset()
			}
			continue
		}
		if err == io.EOF {
			break
		}

		req.kind = connTypeTCP
		req.message.UnpackHeaderQuestion()
		req.writer = cl

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
		res *Message
	)

	for req := range srv.requestq {
		if debug.Value >= 1 {
			fmt.Printf("dns: < %s %d:%s\n",
				connTypeNames[req.kind],
				req.message.Header.ID,
				req.message.Question)
		}

		ans, an := srv.caches.get(string(req.message.Question.Name),
			req.message.Question.Type,
			req.message.Question.Class)

		if ans == nil {
			if srv.hasForwarders {
				srv.primaryq <- req
				continue
			}

			req.message.SetResponseCode(RCodeErrName)
		}

		if an == nil {
			if srv.hasForwarders {
				srv.primaryq <- req
				continue
			}

			req.message.SetQuery(false)
			req.message.SetAuthorativeAnswer(true)
			res = req.message

			if debug.Value >= 1 {
				fmt.Printf("dns: ! %s %d:%s\n",
					connTypeNames[req.kind],
					res.Header.ID, res.Question)
			}
		} else {
			if an.msg.IsExpired() && srv.hasForwarders {
				if debug.Value >= 1 {
					fmt.Printf("dns: ~ %s %d:%s\n",
						connTypeNames[req.kind],
						req.message.Header.ID,
						req.message.Question)
				}
				srv.primaryq <- req
				continue
			}

			an.msg.SetID(req.message.Header.ID)
			an.updateTTL()
			res = an.msg

			if debug.Value >= 1 {
				fmt.Printf("dns: > %s %d:%s\n",
					connTypeNames[req.kind],
					res.Header.ID, res.Question)
			}
		}

		_, err := req.writer.Write(res.Packet)
		if err != nil {
			log.Println("dns: processRequest: ", err.Error())
		}
	}
}

func (srv *Server) processResponse(req *request, res *Message, fallbackq chan *request) {
	if !isResponseValid(req, res) {
		srv.requestq <- req
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
			res.Header.ID, res.Question)
		if fallbackq != nil {
			fallbackq <- req
		}
		return
	}

	an := newAnswer(res, false)
	inserted := srv.caches.upsert(an)

	if debug.Value >= 1 {
		if inserted {
			fmt.Printf("dns: + %s %d:%s\n",
				connTypeNames[req.kind],
				res.Header.ID, res.Question)
		} else {
			fmt.Printf("dns: # %s %d:%s\n",
				connTypeNames[req.kind],
				res.Header.ID, res.Question)
		}
	}
}

func (srv *Server) runForwarders() {
	srv.hasForwarders = false
	srv.forwardStoppers = nil

	nforwarders := 0
	for x := 0; x < len(srv.opts.primaryUDP); x++ {
		go srv.runUDPForwarder(srv.opts.primaryUDP[x].String(), srv.primaryq, srv.fallbackq)
		nforwarders++
	}

	for x := 0; x < len(srv.opts.primaryTCP); x++ {
		go srv.runTCPForwarder(srv.opts.primaryTCP[x].String(), srv.primaryq, srv.fallbackq)
		nforwarders++
	}

	for x := 0; x < len(srv.opts.primaryDoh); x++ {
		go srv.runDohForwarder(srv.opts.primaryDoh[x], srv.primaryq, srv.fallbackq)
		nforwarders++
	}

	if nforwarders > 0 {
		srv.hasForwarders = true
	}

	nforwarders = 0
	for x := 0; x < len(srv.opts.fallbackUDP); x++ {
		go srv.runUDPForwarder(srv.opts.fallbackUDP[x].String(), srv.fallbackq, nil)
		nforwarders++
	}

	for x := 0; x < len(srv.opts.fallbackTCP); x++ {
		go srv.runTCPForwarder(srv.opts.fallbackTCP[x].String(), srv.fallbackq, nil)
		nforwarders++
	}

	for x := 0; x < len(srv.opts.fallbackDoh); x++ {
		go srv.runDohForwarder(srv.opts.fallbackDoh[x], srv.fallbackq, nil)
		nforwarders++
	}
}

func (srv *Server) runDohForwarder(nameserver string, primaryq, fallbackq chan *request) {
	var asWhat = asPrimary

	if fallbackq == nil {
		asWhat = asFallback
	}

	stop := make(chan bool)
	srv.forwardStoppers = append(srv.forwardStoppers, stop)
	srv.fwGroup.Add(1)

	fmt.Printf("dns: starting %s DoH forwarder at %s\n", asWhat, nameserver)

	for {
		forwarder, err := NewDoHClient(nameserver, false)
		if err != nil {
			log.Fatal("dns: failed to create DoH forwarder: " + err.Error())
		}

		for err == nil {
			select {
			case req, ok := <-primaryq:
				if !ok {
					break
				}
				if debug.Value >= 1 {
					fmt.Printf("dns: ^ DoH %s %d:%s\n",
						nameserver,
						req.message.Header.ID, req.message.Question)
				}

				res, err := forwarder.Query(req.message)
				if err != nil {
					log.Println("dns: failed to query DoH: " + err.Error())
					if fallbackq != nil {
						fallbackq <- req
					}
				} else {
					srv.processResponse(req, res, fallbackq)
				}
			case <-stop:
				goto out
			}
		}

		forwarder.Close()

		log.Println("dns: restarting DoH forwarder for " + nameserver)
	}
out:
	srv.fwGroup.Done()
	fmt.Printf("dns: DoH %s forwarder %s has been stopped\n", asWhat, nameserver)
}

func (srv *Server) runTCPForwarder(remoteAddr string, primaryq, fallbackq chan *request) {
	var asWhat = asPrimary

	if fallbackq == nil {
		asWhat = asFallback
	}

	stop := make(chan bool)
	srv.forwardStoppers = append(srv.forwardStoppers, stop)
	srv.fwGroup.Add(1)

	fmt.Printf("dns: starting %s TCP forwarder at %s\n", asWhat, remoteAddr)

	for {
		select {
		case req, ok := <-primaryq:
			if !ok {
				break
			}
			if debug.Value >= 1 {
				fmt.Printf("dns: ^ TCP %s %d:%s\n",
					remoteAddr,
					req.message.Header.ID, req.message.Question)
			}

			cl, err := NewTCPClient(remoteAddr)
			if err != nil {
				log.Println("dns: failed to create TCP client: " + err.Error())
				err = nil
				continue
			}

			res, err := cl.Query(req.message)
			cl.Close()
			if err != nil {
				log.Println("dns: failed to query TCP: " + err.Error())
				if fallbackq != nil {
					fallbackq <- req
				}
				continue
			}

			srv.processResponse(req, res, fallbackq)
		case <-stop:
			goto out
		}
	}
out:
	srv.fwGroup.Done()
	fmt.Printf("dns: TCP %s forwarder %s has been stopped\n", asWhat, remoteAddr)
}

//
// runUDPForwarder create a UDP client that consume request from forward queue
// and forward it to parent server at "remoteAddr".
//
func (srv *Server) runUDPForwarder(remoteAddr string, primaryq, fallbackq chan *request) {
	var asWhat = asPrimary

	if fallbackq == nil {
		asWhat = asFallback
	}

	stop := make(chan bool)
	srv.forwardStoppers = append(srv.forwardStoppers, stop)
	srv.fwGroup.Add(1)

	fmt.Printf("dns: starting %s UDP forwarder at %s\n", asWhat, remoteAddr)

	// The first loop handle broken connection.
	for {
		forwarder, err := NewUDPClient(remoteAddr)
		if err != nil {
			log.Fatal("dns: failed to create UDP forwarder: " + err.Error())
		}

		// The second loop consume the forward queue.
		for err == nil {
			select {
			case req, ok := <-primaryq:
				if !ok {
					fmt.Printf("dns: UDP break %s\n", remoteAddr)
					break
				}
				if debug.Value >= 1 {
					fmt.Printf("dns: ^ UDP %s %d:%s\n",
						remoteAddr,
						req.message.Header.ID, req.message.Question)
				}

				res, err := forwarder.Query(req.message)
				if err != nil {
					log.Println("dns: failed to query UDP: " + err.Error())
					if fallbackq != nil {
						fallbackq <- req
					}
				} else {
					srv.processResponse(req, res, fallbackq)
				}
			case <-stop:
				goto out
			}
		}

		forwarder.Close()
		log.Println("dns: restarting UDP forwarder for " + remoteAddr)
	}
out:
	srv.fwGroup.Done()
	fmt.Printf("dns: UDP %s forwarder %s has been stopped\n", asWhat, remoteAddr)
}

//
// stopForwarders stop all forwarder connections.
//
func (srv *Server) stopForwarders() {
	srv.hasForwarders = false
	for _, forwardStop := range srv.forwardStoppers {
		forwardStop <- true
	}
	srv.fwGroup.Wait()
	fmt.Println("dns: all forwarders has been stopped")
}
