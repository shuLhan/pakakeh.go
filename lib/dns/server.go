// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
	"github.com/shuLhan/share/lib/debug"
)

const (
	aliveInterval = 10 * time.Second
)

// Server defines DNS server.
//
// # Services
//
// The server will listening for DNS over TLS only if certificates file is
// exist and valid.
//
// # Caches
//
// There are two type of answer: internal and external.
// Internal answers is DNS records that are loaded from hosts or zone files.
// Internal answers never get pruned.
// External answers is DNS records that are received from parent name
// servers.
//
// Server caches the DNS answers in two storages: map and list.List.
// The map caches store internal and external answers, using domain name as a
// key and list of answers as value,
//
//	domain-name -> [{A,IN,...},{AAAA,IN,...}]
//
// The list.List store external answers, ordered by accessed time,
// it is used to prune least frequently accessed answers.
//
// # Debugging
//
// If debug.Value is set to value greater than 1, server will print each
// processed request, forward, and response.
// The debug information prefixed with single character to differentiate
// single action,
//
//	> : incoming request from client
//	< : the answer is sent to client
//	! : no answer found on cache and the query is not recursive, or
//	    response contains error code
//	^ : request is forwarded to parent name server
//	* : request is dropped from queue
//	~ : answer exist on cache but its expired
//	- : answer is pruned from caches
//	+ : new answer is added to caches
//	# : the expired answer is renewed and updated on caches
//
// Following the prefix is connection type, parent name server address,
// message ID, and question.
type Server struct {
	HostsFiles  map[string]*HostsFile
	opts        *ServerOptions
	caches      *caches
	tlsConfig   *tls.Config
	udp         *net.UDPConn
	tcp         *net.TCPListener
	doh         *http.Server
	dot         net.Listener
	requestq    chan *request
	primaryq    chan *request
	tcpq        chan *request
	errListener chan error
	fwStoppers  []chan bool
	fwn         int
	fwLocker    sync.Mutex
}

// NewServer create and initialize DNS server.
func NewServer(opts *ServerOptions) (srv *Server, err error) {
	err = opts.init()
	if err != nil {
		return nil, err
	}

	srv = &Server{
		opts:     opts,
		requestq: make(chan *request, 512),
		primaryq: make(chan *request, 512),
		tcpq:     make(chan *request, 512),
	}

	var (
		udpAddr *net.UDPAddr
		tcpAddr *net.TCPAddr
	)

	udpAddr = opts.getUDPAddress()
	srv.udp, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("dns: error listening on UDP '%v': %s", udpAddr, err)
	}

	tcpAddr = opts.getTCPAddress()
	srv.tcp, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("dns: error listening on TCP '%v': %s", tcpAddr, err)
	}

	if len(opts.TLSCertFile) > 0 && len(opts.TLSPrivateKey) > 0 {
		var (
			cert tls.Certificate
		)

		cert, err = tls.LoadX509KeyPair(opts.TLSCertFile, opts.TLSPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("dns: error loading certificate: %w", err)
		}

		srv.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{
				cert,
			},
			InsecureSkipVerify: opts.TLSAllowInsecure,
		}
	}

	srv.errListener = make(chan error, 1)
	srv.caches = newCaches(opts.PruneDelay, opts.PruneThreshold)

	return srv, nil
}

// isResponseValid check if request name, type, and class match with response.
// It will return true if both matched, otherwise it will return false.
func isResponseValid(req *request, res *Message) bool {
	if req.message.Question.Name != res.Question.Name {
		log.Printf("dns: unmatched response name, got %s want %s",
			req.message.Question.Name, res.Question.Name)
		return false
	}
	if req.message.Question.Type != res.Question.Type {
		log.Printf("dns: unmatched response type, got %s want %s",
			req.message.Question.String(), res.Question.String())
		return false
	}
	if req.message.Question.Class != res.Question.Class {
		log.Printf("dns: unmatched response class, got %s want %s",
			req.message.Question.String(), res.Question.String())
		return false
	}

	return true
}

// CachesLoad load the gob encoded external answers from r.
func (srv *Server) CachesLoad(r io.Reader) (answers []*Answer, err error) {
	var (
		logp   = "CachesLoad"
		answer *Answer
	)

	answers, err = srv.caches.read(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	for _, answer = range answers {
		_ = srv.caches.upsert(answer)
	}
	return answers, nil
}

// CachesLRU return list of external caches ordered by the least recently
// used.
func (srv *Server) CachesLRU() []*Answer {
	return srv.caches.list()
}

// CachesSave write the external answers into w, encoded with gob.
func (srv *Server) CachesSave(w io.Writer) (n int, err error) {
	n, err = srv.caches.write(w)
	if err != nil {
		return 0, fmt.Errorf("CachesSave: %w", err)
	}
	return n, nil
}

// SearchCaches search caches by query (domain) name that match with the
// regular expresion.
func (srv *Server) SearchCaches(re *regexp.Regexp) []*Message {
	return srv.caches.search(re)
}

// PopulateCaches add list of message to caches.
func (srv *Server) PopulateCaches(msgs []*Message, from string) {
	var (
		msg      *Message
		an       *Answer
		n        int
		inserted bool
		isLocal  = true
	)

	for _, msg = range msgs {
		an = newAnswer(msg, isLocal)
		inserted = srv.caches.upsert(an)
		if inserted {
			n++
		}
	}

	if debug.Value >= 1 {
		fmt.Printf("dns: %d out of %d records cached from %q\n", n, len(msgs), from)
	}
}

// PopulateCachesByRR update or insert new ResourceRecord into caches.
func (srv *Server) PopulateCachesByRR(listRR []*ResourceRecord, from string) (err error) {
	var (
		rr *ResourceRecord
		n  int
	)

	for _, rr = range listRR {
		err = srv.caches.upsertInternalRR(rr)
		if err != nil {
			return err
		}
		n++
	}
	if debug.Value >= 1 {
		fmt.Printf("dns: %d out of %d records cached from %q\n", n, len(listRR), from)
	}
	return nil
}

// CachesClear remove all external answers.
func (srv *Server) CachesClear() (listAnswer []*Answer) {
	listAnswer = srv.caches.prune(math.MaxInt64)
	return listAnswer
}

// RemoveCachesByNames remove the caches by domain names.
func (srv *Server) RemoveCachesByNames(names []string) (listAnswer []*Answer) {
	var (
		answers []*Answer
		name    string
	)
	for _, name = range names {
		answers = srv.caches.remove(name)
		if len(answers) > 0 {
			listAnswer = append(listAnswer, answers...)
			if debug.Value >= 1 {
				fmt.Println("dns: - ", name)
			}
		}
	}
	return listAnswer
}

// RemoveCachesByRR remove the answer from caches by ResourceRecord name,
// type, class, and value.
func (srv *Server) RemoveCachesByRR(rr *ResourceRecord) (rrOut *ResourceRecord, err error) {
	rrOut, err = srv.caches.removeInternalByRR(rr)
	return rrOut, err
}

// RemoveLocalCachesByNames remove internal caches by domain names.
func (srv *Server) RemoveLocalCachesByNames(names []string) {
	var (
		x int
	)
	srv.caches.Lock()
	for ; x < len(names); x++ {
		delete(srv.caches.internal, names[x])
		if debug.Value >= 1 {
			fmt.Println("dns: - ", names[x])
		}
	}
	srv.caches.Unlock()
}

// RestartForwarders stop and start new forwarders with new nameserver address
// and protocol.
// Empty nameservers means server will run without forwarding request.
func (srv *Server) RestartForwarders(nameServers []string) {
	fmt.Printf("dns: RestartForwarders: %s\n", nameServers)

	srv.opts.NameServers = nameServers

	srv.opts.initNameServers()

	srv.stopAllForwarders()
	srv.startAllForwarders()
}

// ListenAndServe start listening and serve queries from clients.
func (srv *Server) ListenAndServe() (err error) {
	srv.startAllForwarders()

	go srv.processRequest()
	if srv.opts.TLSPort > 0 {
		go srv.serveDoT()
	}
	if srv.opts.HTTPPort > 0 {
		go srv.serveDoH()
	}
	go srv.serveTCP()
	go srv.serveUDP()

	return <-srv.errListener
}

// Stop the forwarders and close all listeners.
func (srv *Server) Stop() {
	var (
		err error
	)

	srv.stopAllForwarders()

	err = srv.udp.Close()
	if err != nil {
		log.Println("dns: error when closing UDP: " + err.Error())
	}
	err = srv.tcp.Close()
	if err != nil {
		log.Println("dns: error when closing TCP: " + err.Error())
	}
	if srv.dot != nil {
		err = srv.dot.Close()
		if err != nil {
			log.Println("dns: error when closing DoT: " + err.Error())
		}
	}
	if srv.doh != nil {
		err = srv.doh.Close()
		if err != nil {
			log.Println("dns: error when closing DoH: " + err.Error())
		}
	}
}

// serveDoH listen for request over HTTPS using certificate and key
// file in parameter.  The path to request is static "/dns-query".
func (srv *Server) serveDoH() {
	var (
		addr string = srv.opts.getHTTPAddress().String()

		err error
	)

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
	if errors.Is(err, io.EOF) {
		err = nil
	} else {
		err = fmt.Errorf("dns: error on DoH: %w", err)
	}

	srv.errListener <- err
}

func (srv *Server) serveDoT() {
	var (
		dotAddr *net.TCPAddr = srv.opts.getDoTAddress()

		cl   *TCPClient
		conn net.Conn
		err  error
	)

	for {
		if srv.opts.DoHBehindProxy || srv.tlsConfig == nil {
			srv.dot, err = net.ListenTCP("tcp", dotAddr)
		} else {
			srv.dot, err = tls.Listen("tcp", dotAddr.String(), srv.tlsConfig)
		}
		if err != nil {
			log.Println("dns: Server.serveDoT: " + err.Error())
			time.Sleep(3 * time.Second)
			continue
		}

		log.Println("dns.Server: listening for DNS over TLS at", dotAddr.String())

		for {
			conn, err = srv.dot.Accept()
			if err != nil {
				if errors.Is(err, io.EOF) {
					err = nil
				} else {
					err = fmt.Errorf("dns: error on accepting DoT connection: %w", err)
				}
				srv.errListener <- err
				break
			}

			cl = &TCPClient{
				writeTimeout: clientTimeout,
				conn:         conn,
			}

			go srv.serveTCPClient(cl, connTypeDoT)
		}
	}
}

// serveTCP serve DNS request from TCP connection.
func (srv *Server) serveTCP() {
	var (
		cl   *TCPClient
		conn net.Conn
		err  error
	)

	log.Println("dns.Server: listening for DNS over TCP at", srv.tcp.Addr())

	for {
		conn, err = srv.tcp.AcceptTCP()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			} else {
				err = fmt.Errorf("dns: error on accepting TCP connection: %w", err)
			}
			srv.errListener <- err
			return
		}

		cl = &TCPClient{
			writeTimeout: clientTimeout,
			conn:         conn,
		}

		go srv.serveTCPClient(cl, connTypeTCP)
	}
}

// serveUDP serve DNS request from UDP connection.
func (srv *Server) serveUDP() {
	var (
		n      int
		packet = make([]byte, maxUdpPacketSize)
		raddr  *net.UDPAddr
		req    *request
		err    error
	)

	log.Println("dns.Server: listening for DNS over UDP at", srv.udp.LocalAddr())
	for {
		n, raddr, err = srv.udp.ReadFromUDP(packet)
		if err != nil {
			if n == 0 || errors.Is(err, io.EOF) {
				err = nil
			} else {
				err = fmt.Errorf("dns: error when reading from UDP: %w", err)
			}
			srv.errListener <- err
			return
		}

		req = newRequest()
		req.message.packet = libbytes.Copy(packet[:n])

		req.kind = connTypeUDP
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
	var (
		hdr http.Header = w.Header()

		hdrAcceptValue string
	)

	hdr.Set(dohHeaderKeyContentType, dohHeaderValDNSMessage)

	hdrAcceptValue = r.Header.Get(dohHeaderKeyAccept)
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
	var (
		q         url.Values = r.URL.Query()
		msgBase64 string     = q.Get("dns")

		raw []byte
		err error
	)

	if len(msgBase64) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	raw, err = base64.RawURLEncoding.DecodeString(msgBase64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srv.handleDoHRequest(raw, w)
}

func (srv *Server) handleDoHPost(w http.ResponseWriter, r *http.Request) {
	var (
		raw []byte
		err error
	)

	raw, err = io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	srv.handleDoHRequest(raw, w)
}

func (srv *Server) handleDoHRequest(raw []byte, w http.ResponseWriter) {
	var (
		req *request = newRequest()
		cl           = &DoHClient{
			w:         w,
			responded: make(chan bool, 1),
		}

		err error
	)

	req.kind = connTypeDoH
	req.writer = cl
	req.message.packet = append(req.message.packet[:0], raw...)

	err = req.message.UnpackHeaderQuestion()
	if err != nil {
		log.Println(err)
		req.error(RCodeErrServer)
		return
	}

	srv.requestq <- req

	cl.waitResponse()
}

// hasForwarders will return true if server run at least one forwarder,
// otherwise it will return false.
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
	var (
		req *request
		err error
	)
	for {
		req = newRequest()

		req.message, err = cl.recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("serveTCPClient: %s: %s", connTypeNames[kind], err)
			}
			break
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

	err = cl.conn.Close()
	if err != nil {
		log.Printf("serveTCPClient: conn.Close: %s: %s", connTypeNames[kind], err)
	}
}

func (srv *Server) isImplemented(msg *Message) bool {
	switch msg.Question.Class {
	case RecordClassCH, RecordClassHS:
		log.Printf("dns: class %d is not implemented", msg.Question.Class)
		return false
	}

	if msg.Question.Type >= RecordTypeA && msg.Question.Type <= RecordTypeTXT {
		return true
	}
	switch msg.Question.Type {
	case RecordTypeAAAA, RecordTypeSRV, RecordTypeOPT, RecordTypeAXFR,
		RecordTypeMAILB, RecordTypeMAILA:
		return true
	}

	log.Printf("dns: type %d is not implemented", msg.Question.Type)

	return false
}

// processRequest from client.
func (srv *Server) processRequest() {
	var (
		an  *Answer
		res *Message
		req *request
		ans *answers
		err error
	)

	for req = range srv.requestq {
		if !srv.isImplemented(req.message) {
			req.error(RCodeNotImplemented)
			continue
		}

		if debug.Value >= 1 {
			fmt.Printf("dns: > %s %d:%s\n",
				connTypeNames[req.kind],
				req.message.Header.ID,
				req.message.Question.String())
		}

		ans, an = srv.caches.get(req.message.Question.Name,
			req.message.Question.Type,
			req.message.Question.Class)

		if ans == nil || an == nil {
			switch {
			case srv.hasForwarders():
				if req.kind == connTypeTCP {
					srv.tcpq <- req
				} else {
					srv.primaryq <- req
				}
			default:
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
			switch {
			case srv.hasForwarders():
				if debug.Value >= 1 {
					fmt.Printf("dns: ~ %s %d:%s\n",
						connTypeNames[req.kind],
						req.message.Header.ID,
						req.message.Question.String())
				}
				if req.kind == connTypeTCP {
					srv.tcpq <- req
				} else {
					srv.primaryq <- req
				}

			default:
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
		res = an.msg

		if debug.Value >= 1 {
			fmt.Printf("dns: < %s %d:%s\n", connTypeNames[req.kind], res.Header.ID, res.Question.String())
		}

		_, err = req.writer.Write(res.packet)
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

	var (
		an       *Answer
		err      error
		inserted bool
	)

	_, err = req.writer.Write(res.packet)
	if err != nil {
		log.Println("dns: processResponse: ", err.Error())
		return
	}

	if res.Header.RCode != 0 {
		log.Printf("dns: ! %s %s %d:%s",
			connTypeNames[req.kind], rcodeNames[res.Header.RCode],
			res.Header.ID, res.Question.String())
		return
	}
	if res.Header.IsTC {
		log.Printf("dns: ! %s TRUNCATED %s", connTypeNames[req.kind], res.Question.String())
		return
	}

	an = newAnswer(res, false)
	inserted = srv.caches.upsert(an)

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

func (srv *Server) startAllForwarders() {
	srv.fwStoppers = nil

	var (
		asPrimary = "primary"

		tag        string
		nameserver string
		x          int
	)

	for x = 0; x < len(srv.opts.primaryUDP); x++ {
		tag = fmt.Sprintf("UDP-%d-%s", x, asPrimary)
		nameserver = srv.opts.primaryUDP[x].String()
		go srv.runUDPForwarder(tag, nameserver)
	}
	for x = 0; x < len(srv.opts.primaryTCP); x++ {
		tag = fmt.Sprintf("TCP-%d-%s", x, asPrimary)
		nameserver = srv.opts.primaryTCP[x].String()
		go srv.runTCPForwarder(tag, nameserver)
	}
	for x = 0; x < len(srv.opts.primaryDoh); x++ {
		tag = fmt.Sprintf("DoH-%d-%s", x, asPrimary)
		nameserver = srv.opts.primaryDoh[x]
		go srv.runDohForwarder(tag, nameserver)
	}
	for x = 0; x < len(srv.opts.primaryDot); x++ {
		tag = fmt.Sprintf("DoT-%d-%s", x, asPrimary)
		nameserver = srv.opts.primaryDot[x]
		go srv.runTLSForwarder(tag, nameserver)
	}
}

func (srv *Server) runDohForwarder(tag, nameserver string) {
	var (
		stopper <-chan bool = srv.newStopper()

		forwarder *DoHClient
		ticker    *time.Ticker
		req       *request
		res       *Message
		err       error
		isRunning bool
		ok        bool
	)

	defer func() {
		log.Printf("dns: forwarder %s for %s has been stopped", tag, nameserver)
	}()

	for {
		forwarder, err = NewDoHClient(nameserver, false)
		if err != nil {
			log.Printf("dns: failed to create forwarder %s: %s", tag, err)

			select {
			case <-stopper:
				srv.stopForwarder(nil)
				return
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf("dns: forwarder %s for %s has been connected ...", tag, nameserver)

		srv.incForwarder()

		isRunning = true
		ticker = time.NewTicker(aliveInterval)
		for isRunning {
			select {
			case req, ok = <-srv.primaryq:
				if !ok {
					log.Println("dns: primary queue has been closed")
					srv.stopForwarder(forwarder)
					return
				}
				if debug.Value >= 1 {
					fmt.Printf("dns: ^ %s %s %d:%s\n",
						tag, nameserver,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf("dns: %s forward failed: %s: %s", tag, req.message.Question.Name, err)
					isRunning = false
					continue
				}
				srv.processResponse(req, res)
			case <-ticker.C:
				if debug.Value >= 2 {
					log.Printf("dns: %s alive", tag)
				}
			case <-stopper:
				srv.stopForwarder(forwarder)
				return
			}
		}

		log.Printf("dns: reconnect forwarder %s for %s", tag, nameserver)
		srv.stopForwarder(forwarder)
	}
}

func (srv *Server) runTLSForwarder(tag, nameserver string) {
	var (
		stopper <-chan bool = srv.newStopper()

		forwarder *DoTClient
		ticker    *time.Ticker
		req       *request
		res       *Message
		err       error
		isRunning bool
		ok        bool
	)

	defer func() {
		log.Printf("dns: forwarder %s for %s has been stopped", tag, nameserver)
	}()

	for {
		forwarder, err = NewDoTClient(nameserver, srv.opts.TLSAllowInsecure)
		if err != nil {
			log.Printf("dns: failed to create forwarder %s: %s", tag, err)

			select {
			case <-stopper:
				srv.stopForwarder(nil)
				return
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf("dns: forwarder %s for %s has been connected ...", tag, nameserver)

		srv.incForwarder()

		isRunning = true
		ticker = time.NewTicker(aliveInterval)
		for isRunning {
			select {
			case req, ok = <-srv.primaryq:
				if !ok {
					log.Println("dns: primary queue has been closed")
					srv.stopForwarder(forwarder)
					return
				}
				if debug.Value >= 1 {
					fmt.Printf("dns: ^ %s %s %d:%s\n",
						tag, nameserver,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf("dns: %s forward failed: %s: %s", tag, req.message.Question.Name, err)
					isRunning = false
					continue
				}

				srv.processResponse(req, res)
			case <-ticker.C:
				if debug.Value >= 2 {
					log.Printf("dns: %s alive", tag)
				}
			case <-stopper:
				srv.stopForwarder(forwarder)
				return
			}
		}

		log.Printf("dns: reconnect forwarder %s for %s", tag, nameserver)
		srv.stopForwarder(forwarder)
	}
}

func (srv *Server) runTCPForwarder(tag, nameserver string) {
	var (
		stopper <-chan bool = srv.newStopper()

		ticker *time.Ticker
		cl     *TCPClient
		req    *request
		res    *Message
		err    error
		ok     bool
	)

	log.Printf("dns: starting forwarder %s for %s", tag, nameserver)

	srv.incForwarder()

	defer func() {
		srv.decForwarder()
		log.Printf("dns: forwarder %s for %s has been stopped", tag, nameserver)
	}()

	ticker = time.NewTicker(aliveInterval)
	for {
		select {
		case req, ok = <-srv.tcpq:
			if !ok {
				log.Println("dns: primary queue has been closed")
				return
			}
			if debug.Value >= 1 {
				fmt.Printf("dns: ^ %s %s %d:%s\n", tag, nameserver,
					req.message.Header.ID,
					req.message.Question.String())
			}

			cl, err = NewTCPClient(nameserver)
			if err != nil {
				log.Printf("dns: failed to create forwarder %s: %s", tag, err)
				continue
			}

			res, err = cl.Query(req.message)
			cl.Close()
			if err != nil {
				log.Printf("dns: %s forward failed: %s: %s", tag, req.message.Question.Name, err)
				continue
			}

			srv.processResponse(req, res)
		case <-ticker.C:
			if debug.Value >= 2 {
				log.Printf("dns: %s alive", tag)
			}
		case <-stopper:
			return
		}
	}
}

// runUDPForwarder create a UDP client that consume request from queue
// and forward it to parent name server.
func (srv *Server) runUDPForwarder(tag, nameserver string) {
	var (
		stopper <-chan bool = srv.newStopper()

		forwarder *UDPClient
		ticker    *time.Ticker
		req       *request
		res       *Message
		err       error
		isRunning bool
		ok        bool
	)

	defer func() {
		log.Printf("dns: forwarder %s for %s has been stopped", tag, nameserver)
	}()

	// The first loop handle broken connection.
	for {
		forwarder, err = NewUDPClient(nameserver)
		if err != nil {
			log.Printf("dns: failed to create forwarder %s: %s", tag, err)

			select {
			case <-stopper:
				srv.stopForwarder(nil)
				return
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf("dns: forwarder %s for %s has been connected ...", tag, nameserver)

		srv.incForwarder()

		// The second loop consume the forward queue.
		isRunning = true
		ticker = time.NewTicker(aliveInterval)
		for isRunning {
			select {
			case req, ok = <-srv.primaryq:
				if !ok {
					log.Println("dns: primary queue has been closed")
					srv.stopForwarder(forwarder)
					return
				}
				if debug.Value >= 1 {
					fmt.Printf("dns: ^ %s %s %d:%s\n",
						tag, nameserver,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf("dns: %s forward failed: %s: %s", tag, req.message.Question.Name, err)
					isRunning = false
					continue
				}
				srv.processResponse(req, res)
			case <-ticker.C:
				if debug.Value >= 2 {
					log.Printf("dns: %s alive", tag)
				}
			case <-stopper:
				srv.stopForwarder(forwarder)
				return
			}
		}

		log.Printf("dns: reconnect forwarder %s for %s", tag, nameserver)
		srv.stopForwarder(forwarder)
	}
}

func (srv *Server) stopForwarder(fw Client) {
	if fw != nil {
		fw.Close()
	}
	srv.decForwarder()
}

// stopAllForwarders stop all forwarder connections.
func (srv *Server) stopAllForwarders() {
	var (
		x int
	)
	for x = 0; x < len(srv.fwStoppers); x++ {
		srv.fwStoppers[x] <- true
	}
	for x = 0; x < len(srv.fwStoppers); x++ {
		close(srv.fwStoppers[x])
	}

	fmt.Println("dns: all forwarders has been stopped")
}

func (srv *Server) newStopper() <-chan bool {
	srv.fwLocker.Lock()

	var stopper = make(chan bool, 1)
	srv.fwStoppers = append(srv.fwStoppers, stopper)

	srv.fwLocker.Unlock()

	return stopper
}
