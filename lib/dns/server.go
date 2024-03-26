// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
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
// If [ServerOptions.Debug] is set to value DebugLevelCache,
// server will print each processed request, forward, and response.
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
	Caches      Caches
	opts        *ServerOptions
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
		return nil, fmt.Errorf(`dns: error listening on UDP '%v': %w`, udpAddr, err)
	}

	tcpAddr = opts.getTCPAddress()
	srv.tcp, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, fmt.Errorf(`dns: error listening on TCP '%v': %w`, tcpAddr, err)
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
			InsecureSkipVerify: opts.TLSAllowInsecure, //nolint:gosec
		}
	}

	srv.errListener = make(chan error, 1)
	srv.Caches.init(opts.PruneDelay, opts.PruneThreshold, opts.Debug)

	return srv, nil
}

// isResponseValid check if request name, type, and class match with response.
// It will return true if both matched, otherwise it will return false.
func isResponseValid(req *request, res *Message) bool {
	if req.message.Question.Name != res.Question.Name {
		log.Printf(`dns: unmatched response name, got %s want %s`,
			req.message.Question.Name, res.Question.Name)
		return false
	}
	if req.message.Question.Type != res.Question.Type {
		log.Printf(`dns: unmatched response type, got %s want %s`,
			req.message.Question.String(), res.Question.String())
		return false
	}
	if req.message.Question.Class != res.Question.Class {
		log.Printf(`dns: unmatched response class, got %s want %s`,
			req.message.Question.String(), res.Question.String())
		return false
	}

	return true
}

// RestartForwarders stop and start new forwarders with new nameserver address
// and protocol.
// Empty nameservers means server will run without forwarding request.
func (srv *Server) RestartForwarders(nameServers []string) {
	log.Printf(`dns: RestartForwarders: %s`, nameServers)

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
		srv.dot = nil
	}
	if srv.doh != nil {
		var ctx = context.Background()
		err = srv.doh.Shutdown(ctx)
		if err != nil {
			log.Println("dns: error when closing DoH: " + err.Error())
		}
		srv.doh = nil
	}
}

// serveDoH listen for request over HTTPS using certificate and key
// file in parameter.  The path to request is static "/dns-query".
func (srv *Server) serveDoH() {
	var (
		logp = `serveDoH`
		addr = srv.opts.getHTTPAddress().String()

		err error
	)

	srv.doh = &http.Server{
		Addr:              addr,
		IdleTimeout:       srv.opts.HTTPIdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}

	http.Handle("/dns-query", srv)

	if srv.tlsConfig != nil && !srv.opts.DoHBehindProxy {
		log.Printf(`%s: listening at %s`, logp, addr)
		srv.doh.TLSConfig = srv.tlsConfig
		err = srv.doh.ListenAndServeTLS("", "")
	} else {
		log.Printf(`%s: listening behind proxy at %s`, logp, addr)
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
		logp    = `serveDoT`
		dotAddr = srv.opts.getDoTAddress()

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
			log.Printf(`%s: failed to listen at %s: %s`,
				logp, dotAddr.String(), err)
			time.Sleep(3 * time.Second)
			continue
		}

		log.Printf(`%s: listening at %s`, logp, dotAddr.String())

		for {
			conn, err = srv.dot.Accept()
			if err != nil {
				if errors.Is(err, io.EOF) {
					err = nil
				} else {
					err = fmt.Errorf(`%s: accept: %w`, logp, err)
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
		logp = `serveTCP`
		cl   *TCPClient
		conn net.Conn
		err  error
	)

	log.Printf(`%s: listening at %s`, logp, srv.tcp.Addr())

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
		logp   = `serveUDP`
		n      int
		packet = make([]byte, maxUDPPacketSize)
		raddr  *net.UDPAddr
		req    *request
		err    error
	)

	log.Printf(`%s: listening at %s`, logp, srv.udp.LocalAddr())
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
			log.Printf(`%s: %s`, logp, err)
			req.error(RCodeErrServer)
			continue
		}

		srv.requestq <- req
	}
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		hdr = w.Header()

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
		q         = r.URL.Query()
		msgBase64 = q.Get("dns")

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
		logp = `handleDoHRequest`
		req  = newRequest()
		cl   = &DoHClient{
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
		log.Printf(`%s: %s`, logp, err)
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
		logp = `serveTCPClient`

		req *request
		err error
	)
	for {
		req = newRequest()

		req.message.packet, err = cl.recv()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf(`%s %s: %s`, logp, connTypeNames[kind], err)
			}
			break
		}

		req.kind = kind
		req.writer = cl

		err = req.message.UnpackHeaderQuestion()
		if err != nil {
			log.Printf(`%s %s: %s`, logp, connTypeNames[kind], err)
			req.error(RCodeErrServer)
			continue
		}

		srv.requestq <- req
	}

	err = cl.conn.Close()
	if err != nil {
		log.Printf(`%s %s: connection closed: %s`, logp, connTypeNames[kind], err)
	}
}

func (srv *Server) isImplemented(msg *Message) bool {
	switch msg.Question.Class {
	case RecordClassCH, RecordClassHS:
		if srv.opts.Debug&DebugLevelDNS != 0 {
			log.Printf(`dns: class %d is not implemented`, msg.Question.Class)
		}
		return false
	}

	if msg.Question.Type >= RecordTypeA && msg.Question.Type <= RecordTypeTXT {
		return true
	}
	switch msg.Question.Type {
	case RecordTypeAAAA, RecordTypeSRV, RecordTypeOPT, RecordTypeAXFR,
		RecordTypeMAILB, RecordTypeMAILA,
		RecordTypeSVCB, RecordTypeHTTPS:
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
		err error
	)

	for req = range srv.requestq {
		if !srv.isImplemented(req.message) {
			req.error(RCodeNotImplemented)
			continue
		}

		if srv.opts.Debug&DebugLevelCache != 0 {
			log.Printf(`dns: > %s %d:%s`,
				connTypeNames[req.kind],
				req.message.Header.ID,
				req.message.Question.String())
		}

		an = srv.Caches.query(req.message)
		if an == nil {
			switch {
			case srv.hasForwarders():
				if req.kind == connTypeTCP {
					srv.tcpq <- req
				} else {
					srv.primaryq <- req
				}
			default:
				if srv.opts.Debug&DebugLevelCache != 0 {
					log.Printf(`dns: * %s %d:%s`,
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
				if srv.opts.Debug&DebugLevelCache != 0 {
					log.Printf(`dns: ~ %s %d:%s`,
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
				if srv.opts.Debug&DebugLevelCache != 0 {
					log.Printf(`dns: * %s %d:%s`,
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

		if srv.opts.Debug&DebugLevelCache != 0 {
			log.Printf(`dns: < %s %d:%s`, connTypeNames[req.kind], res.Header.ID, res.Question.String())
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
		if srv.opts.Debug&DebugLevelDNS != 0 {
			log.Printf(`dns: ! %s %s %d:%s`,
				connTypeNames[req.kind], rcodeNames[res.Header.RCode],
				res.Header.ID, res.Question.String())
		}
		return
	}
	if res.Header.IsTC {
		if srv.opts.Debug&DebugLevelDNS != 0 {
			log.Printf(`dns: ! %s TRUNCATED %s`, connTypeNames[req.kind], res.Question.String())
		}
		return
	}
	if res.Header.ANCount == 0 {
		// Ignore empty answers.
		// The use case if one use and switch between two different
		// networks with internal zone, frequently.
		// For example, if on network Y they have domain MY.Y and
		// current connection is X, request to MY.Y will return an
		// empty answers.
		// Once they connect to Y again, any request to MY.Y will not
		// be possible because rescached caches contains empty answer
		// for MY.Y.
		if srv.opts.Debug&DebugLevelDNS != 0 {
			log.Printf(`dns: ! %s EMPTY: %s`, connTypeNames[req.kind], res.Question.String())
		}
		return
	}

	an = newAnswer(res, false)
	inserted = srv.Caches.upsert(an)

	if srv.opts.Debug&DebugLevelCache != 0 {
		if inserted {
			log.Printf(`dns: + %s %d:%s`,
				connTypeNames[req.kind],
				res.Header.ID, res.Question.String())
		} else {
			log.Printf(`dns: # %s %d:%s`,
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
		go srv.udpForwarder(tag, nameserver)
	}
	for x = 0; x < len(srv.opts.primaryTCP); x++ {
		tag = fmt.Sprintf("TCP-%d-%s", x, asPrimary)
		nameserver = srv.opts.primaryTCP[x].String()
		go srv.tcpForwarder(tag, nameserver)
	}
	for x = 0; x < len(srv.opts.primaryDoh); x++ {
		tag = fmt.Sprintf("DoH-%d-%s", x, asPrimary)
		nameserver = srv.opts.primaryDoh[x]
		go srv.dohForwarder(tag, nameserver)
	}
	for x = 0; x < len(srv.opts.primaryDot); x++ {
		tag = fmt.Sprintf("DoT-%d-%s", x, asPrimary)
		nameserver = srv.opts.primaryDot[x]
		go srv.tlsForwarder(tag, nameserver)
	}
}

func (srv *Server) dohForwarder(tag, nameserver string) {
	var (
		logp    = `dohForwarder`
		stopper = srv.newStopper()

		forwarder *DoHClient
		ticker    *time.Ticker
		req       *request
		res       *Message
		err       error
		isRunning bool
		ok        bool
	)

	defer func() {
		log.Printf(`%s %s: forwarder for %s has been stopped`, logp, tag, nameserver)
	}()

	for {
		forwarder, err = NewDoHClient(nameserver, false)
		if err != nil {
			log.Printf(`%s %s: failed to connect to %s: %s`, logp, tag, nameserver, err)

			select {
			case <-stopper:
				srv.stopForwarder(nil)
				return
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf(`%s %s: connected to namesever %s`, logp, tag, nameserver)

		srv.incForwarder()

		isRunning = true
		ticker = time.NewTicker(aliveInterval)
		for isRunning {
			select {
			case req, ok = <-srv.primaryq:
				if !ok {
					log.Printf(`%s %s: primary queue has been closed`,
						logp, tag)
					srv.stopForwarder(forwarder)
					return
				}
				if srv.opts.Debug&DebugLevelCache != 0 {
					log.Printf(`dns: ^ %s %s %d:%s`,
						tag, nameserver,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf(`%s %s: forward failed for %q: %s`,
						logp, tag, req.message.Question.Name, err)
					if !errors.Is(err, errUnpack) {
						isRunning = false
					}
					continue
				}
				srv.processResponse(req, res)
			case <-ticker.C:
				if srv.opts.Debug&DebugLevelConnPacket != 0 {
					log.Printf(`%s %s: alive`, logp, tag)
				}
			case <-stopper:
				srv.stopForwarder(forwarder)
				return
			}
		}

		log.Printf(`%s %s: reconnect to nameserver %s`, logp, tag, nameserver)
		srv.stopForwarder(forwarder)
	}
}

func (srv *Server) tlsForwarder(tag, nameserver string) {
	var (
		logp    = `tlsForwarder`
		stopper = srv.newStopper()

		forwarder *DoTClient
		ticker    *time.Ticker
		req       *request
		res       *Message
		err       error
		isRunning bool
		ok        bool
	)

	defer func() {
		log.Printf(`%s %s: forwarder for %s has been stopped`, logp, tag, nameserver)
	}()

	for {
		forwarder, err = NewDoTClient(nameserver, srv.opts.TLSAllowInsecure)
		if err != nil {
			log.Printf(`%s %s: failed to connect to %s: %s`, logp, tag, nameserver, err)

			select {
			case <-stopper:
				srv.stopForwarder(nil)
				return
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf(`%s %s: connected to nameserver %s`, logp, tag, nameserver)

		srv.incForwarder()

		isRunning = true
		ticker = time.NewTicker(aliveInterval)
		for isRunning {
			select {
			case req, ok = <-srv.primaryq:
				if !ok {
					log.Printf(`%s %s: primary queue has been closed`, logp, tag)
					srv.stopForwarder(forwarder)
					return
				}
				if srv.opts.Debug&DebugLevelCache != 0 {
					log.Printf(`dns: ^ %s %s %d:%s`,
						tag, nameserver,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf(`%s %s: forward failed for %s: %s`,
						logp, tag, req.message.Question.Name, err)
					if !errors.Is(err, errUnpack) {
						isRunning = false
					}
					continue
				}

				srv.processResponse(req, res)
			case <-ticker.C:
				if srv.opts.Debug&DebugLevelConnPacket != 0 {
					log.Printf(`%s %s: alive`, logp, tag)
				}
			case <-stopper:
				srv.stopForwarder(forwarder)
				return
			}
		}

		log.Printf(`%s %s: reconnect to nameserver %s`, logp, tag, nameserver)
		srv.stopForwarder(forwarder)
	}
}

func (srv *Server) tcpForwarder(tag, nameserver string) {
	var (
		logp    = `tcpForwarder`
		stopper = srv.newStopper()

		ticker *time.Ticker
		cl     *TCPClient
		req    *request
		res    *Message
		err    error
		ok     bool
	)

	log.Printf(`%s %s: starting forwarder for %s`, logp, tag, nameserver)

	srv.incForwarder()

	defer func() {
		srv.decForwarder()
		log.Printf(`%s %s: forwarder for %s has been stopped`, logp, tag, nameserver)
	}()

	ticker = time.NewTicker(aliveInterval)
	for {
		select {
		case req, ok = <-srv.tcpq:
			if !ok {
				log.Printf(`%s %s: primary queue has been closed`,
					logp, tag)
				return
			}
			if srv.opts.Debug&DebugLevelCache != 0 {
				log.Printf(`dns: ^ %s %s %d:%s`, tag, nameserver,
					req.message.Header.ID,
					req.message.Question.String())
			}

			cl, err = NewTCPClient(nameserver)
			if err != nil {
				log.Printf(`%s %s: failed to connect to %s: %s`,
					logp, tag, nameserver, err)
				continue
			}

			res, err = cl.Query(req.message)
			cl.Close()
			if err != nil {
				log.Printf(`%s %s: forward failed for %s: %s`,
					logp, tag, req.message.Question.Name, err)
				continue
			}

			srv.processResponse(req, res)
		case <-ticker.C:
			if srv.opts.Debug&DebugLevelConnPacket != 0 {
				log.Printf(`%s %s: alive`, logp, tag)
			}
		case <-stopper:
			return
		}
	}
}

// udpForwarder create a UDP client that consume request from queue
// and forward it to parent name server.
func (srv *Server) udpForwarder(tag, nameserver string) {
	var (
		logp    = `udpForwarder`
		stopper = srv.newStopper()

		forwarder *UDPClient
		ticker    *time.Ticker
		req       *request
		res       *Message
		err       error
		isRunning bool
		ok        bool
	)

	defer func() {
		log.Printf(`%s %s: forwarder for %s has been stopped`,
			logp, tag, nameserver)
	}()

	// The first loop handle broken connection.
	for {
		forwarder, err = NewUDPClient(nameserver)
		if err != nil {
			log.Printf(`%s %s: failed to connect to %s: %s`,
				logp, tag, nameserver, err)

			select {
			case <-stopper:
				srv.stopForwarder(nil)
				return
			default:
				time.Sleep(3 * time.Second)
			}
			continue
		}

		log.Printf(`%s %s: connected to %s`, logp, tag, nameserver)

		srv.incForwarder()

		// The second loop consume the forward queue.
		isRunning = true
		ticker = time.NewTicker(aliveInterval)
		for isRunning {
			select {
			case req, ok = <-srv.primaryq:
				if !ok {
					log.Printf(`%s %s: primary queue has been closed`,
						logp, tag)
					srv.stopForwarder(forwarder)
					return
				}
				if srv.opts.Debug&DebugLevelCache != 0 {
					log.Printf(`dns: ^ %s %s %d:%s`,
						tag, nameserver,
						req.message.Header.ID,
						req.message.Question.String())
				}

				res, err = forwarder.Query(req.message)
				if err != nil {
					log.Printf(`%s %s: forward failed for %s: %s`,
						logp, tag,
						req.message.Question.Name, err)
					if !errors.Is(err, errUnpack) {
						isRunning = false
					}
					continue
				}
				srv.processResponse(req, res)
			case <-ticker.C:
				if srv.opts.Debug&DebugLevelConnPacket != 0 {
					log.Printf(`%s %s: alive`, logp, tag)
				}
			case <-stopper:
				srv.stopForwarder(forwarder)
				return
			}
		}

		log.Printf(`%s %s: reconnect forwarder for %s`, logp, tag, nameserver)
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

	log.Println(`dns: all forwarders has been stopped`)
}

func (srv *Server) newStopper() <-chan bool {
	srv.fwLocker.Lock()

	var stopper = make(chan bool, 1)
	srv.fwStoppers = append(srv.fwStoppers, stopper)

	srv.fwLocker.Unlock()

	return stopper
}
