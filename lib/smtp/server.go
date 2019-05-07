// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/shuLhan/share/lib/debug"
)

const (
	localPostmaster = "postmaster"
)

//
// Server defines parameters for running an SMTP server.
//
type Server struct {
	// address to listen for incoming connections.
	// This field is optional and exist only for the purpose of testing.
	// If its empty, it will set default to ":25".
	address string

	// tlsAddress define an address when listening with security layer.
	// This field is optional and exist only for the purpose of testing.
	// If its empty, it will set default to ":465".
	tlsAddress string

	// TLSCert the server certificate for TLS or nil if no certificate.
	// This field is optional, if its non nil, the server will also listen
	// on address defined in TLSAddress.
	TLSCert *tls.Certificate

	// Env contains server environment.
	Env *Environment

	//
	// Exts define list of custom extensions that the server will provide.
	//
	Exts []Extension

	//
	// Handler define an interface that will process the bouncing email,
	// incoming email, EXPN command, and VRFY command.
	// This field is optional, if not set, it will default to
	// LocalHandler.
	//
	Handler Handler

	// listener is a socket that listen for new connection from client.
	listener net.Listener

	// tlsListener is a socket that listen for new connection from client
	// on secure layer on port 465.
	tlsListener net.Listener

	// mailTxQueue hold mail objects before being relayed or stored.
	mailTxQueue chan *MailTx

	// bounceQueue hold mail objects with invalid recipient where it
	// will be notified to sender.
	bounceQueue chan *MailTx

	// relayQueue hold mail objects that need to be relayed to other MTA.
	relayQueue chan *MailTx

	wg      sync.WaitGroup
	running bool
}

//
// LoadCertificate load TLS certificate and its private key from file.
//
func (srv *Server) LoadCertificate(certFile, keyFile string) (err error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("smtp: LoadCertificate: " + err.Error())
	}

	srv.TLSCert = &cert

	return nil
}

//
// Start listening for SMTP connections.
// Each client connection will be handled in a single routine.
//
func (srv *Server) Start() (err error) {
	err = srv.initialize()
	if err != nil {
		return
	}

	srv.running = true

	srv.wg.Add(1)
	go srv.processRelayQueue()

	srv.wg.Add(1)
	go srv.processBounceQueue()

	srv.wg.Add(1)
	go srv.processMailTxQueue()

	srv.wg.Add(1)
	go srv.serveIncoming()

	srv.wg.Add(1)
	go srv.serveTLS()

	srv.wg.Wait()
	srv.running = false

	return nil
}

//
// Stop the server.
//
func (srv *Server) Stop() {
	err := srv.tlsListener.Close()
	if err != nil {
		log.Printf("smtp: tlsListener.Close: %s", err)
	}

	err = srv.listener.Close()
	if err != nil {
		log.Printf("smtp: listener.Close: %s", err)
	}

	close(srv.mailTxQueue)
	close(srv.bounceQueue)
	close(srv.relayQueue)
}

//
// serveIncoming serve incoming message from other mail transfer agent on port
// 25.
//
func (srv *Server) serveIncoming() {
	for {
		if debug.Value >= 2 {
			fmt.Println("smtp: server: MTA ready ...")
		}

		conn, err := srv.listener.Accept()
		if err != nil {
			log.Printf("smtp: listener.Accept: %s", err)
			srv.wg.Done()
			return
		}

		recv := newReceiver(conn, receiverModeServer)

		go srv.handle(recv)
	}
}

func (srv *Server) serveTLS() {
	for {
		if debug.Value >= 2 {
			fmt.Println("smtp: server: MSA ready ...")
		}

		conn, err := srv.tlsListener.Accept()
		if err != nil {
			log.Printf("smtp: tlsListener.Accept: %s", err)
			srv.wg.Done()
			return
		}

		recv := newReceiver(conn, receiverModeClient)

		go srv.handle(recv)
	}
}

//
// handle receiver connection.
//
func (srv *Server) handle(recv *receiver) {
	err := recv.sendReply(StatusReady, srv.Env.PrimaryDomain.Name, nil)
	if err != nil {
		log.Println("receiver.sendReply: ", err.Error())
		recv.close()
		return
	}

	for srv.running {
		cmd, err := recv.readCommand()
		if err != nil {
			log.Println(err)
			_ = recv.sendError(err)
			break
		}
		if cmd == nil {
			continue
		}

		for _, ext := range srv.Exts {
			err = ext.ValidateCommand(cmd)
			if err != nil {
				break
			}
		}
		if err != nil {
			_ = recv.sendError(err)
			continue
		}

		err = srv.handleCommand(recv, cmd)
		if err != nil {
			log.Println("Server.handleCommand: ", err.Error())
			break
		}

		switch recv.state {
		case CommandDATA:
			srv.processMailTx(recv.mail)
			// TODO: add return error and check it.
			//			if err != nil {
			//				log.Println("server.processMailTx: ", err.Error())
			//				err = recv.sendError(errInProcessing)
			//				if err != nil {
			//					goto out
			//				}
			//				continue
			//			}

			err = recv.sendReply(StatusOK, "OK", nil)
			if err != nil {
				goto out
			}

			recv.reset()

		case CommandQUIT:
			goto out
		}
	}
out:
	recv.close()
}

//
// handleCommand from client.
func (srv *Server) handleCommand(recv *receiver, cmd *Command) (err error) { // nolint: gocyclo
	if debug.Value > 0 {
		log.Printf("handleCommand: %v\n", cmd)
	}

	switch cmd.Kind {
	case CommandAUTH:
		err = srv.handleAUTH(recv, cmd)

	case CommandDATA:
		err = srv.handleDATA(recv)

	case CommandEHLO:
		err = srv.handleEHLO(recv, cmd)

	case CommandHELO:
		recv.clientDomain = cmd.Arg

		err = recv.sendReply(StatusOK, srv.Env.PrimaryDomain.Name, nil)
		if err != nil {
			return err
		}
		recv.state = cmd.Kind

	case CommandMAIL:
		err = srv.handleMAIL(recv, cmd)

	case CommandRCPT:
		err = srv.handleRCPT(recv, cmd)

	case CommandRSET:
		recv.reset()

		err = recv.sendReply(StatusOK, "OK", nil)
		if err != nil {
			return err
		}

	case CommandVRFY:
		err = srv.handleVRFY(recv, cmd)

	case CommandEXPN:
		err = srv.handleEXPN(recv, cmd)

	case CommandHELP:
		if !recv.isAuthenticated() {
			err = recv.sendError(errNotAuthenticated)
			if err != nil {
				return err
			}
		}

		err = srv.handleHELP(recv)
		if err != nil {
			return err
		}

	case CommandNOOP:
		err = recv.sendReply(StatusOK, "OK", nil)

	case CommandQUIT:
		_ = recv.sendReply(StatusClosing,
			"Service closing transmission channel", nil)
		recv.state = CommandQUIT
	}

	return err
}

//
// handleAUTH process the AUTH command from client.
//
func (srv *Server) handleAUTH(recv *receiver, cmd *Command) (err error) {
	if recv.mode == receiverModeServer {
		return recv.sendError(errCmdUnknown)
	}

	if recv.isAuthenticated() {
		return recv.sendError(errBadSequence)
	}

	switch recv.state {
	case CommandMAIL, CommandRCPT, CommandDATA:
		return recv.sendError(errBadSequence)
	}

	var username, password string

	switch cmd.Arg {
	case "PLAIN":
		// AUTH PLAIN with two steps handshake.
		if len(cmd.Param) == 0 {
			err = recv.sendReply(StatusAuthReady, "", nil)
			if err != nil {
				return err
			}

			err = recv.readAuthData(cmd)
			if err != nil {
				return err
			}

			if cmd.Param == "*" {
				err = recv.sendReply(StatusCmdSyntaxError,
					"Authentication cancelled", nil)
				return err
			}
		}

		param, err := base64.StdEncoding.DecodeString(cmd.Param)
		if err != nil {
			_ = recv.sendError(errCmdSyntaxError)
			return err
		}

		args := bytes.Split(param, []byte{'\x00'})
		if len(args) != 3 {
			return recv.sendError(errCmdSyntaxError)
		}

		username = string(args[1])
		password = string(args[2])

	default:
		return recv.sendError(errAuthMechanism)
	}

	res, err := srv.Handler.ServeAuth(username, password)
	if err != nil {
		return recv.sendError(err)
	}

	err = recv.sendReply(res.Code, res.Message, res.Body)
	if err != nil {
		return err
	}

	recv.authenticated = true
	recv.state = CommandAUTH

	return nil
}

func (srv *Server) handleDATA(recv *receiver) (err error) {
	if !recv.isAuthenticated() {
		err = recv.sendError(errNotAuthenticated)
		if err != nil {
			return err
		}
	}
	if recv.state != CommandRCPT {
		err = recv.sendReply(StatusCmdBadSequence,
			"Bad sequences of commands", nil)
		if err != nil {
			return err
		}
		recv.reset()
		return nil
	}

	err = recv.sendReply(StatusDataReady, "Start mail input.", nil)
	if err != nil {
		return err
	}

	err = recv.readDATA()
	if err != nil {
		return err
	}

	recv.state = CommandDATA

	return nil
}

func (srv *Server) handleEHLO(recv *receiver, cmd *Command) (err error) {
	recv.clientDomain = cmd.Arg

	body := make([]string, len(srv.Exts))
	for x, ext := range srv.Exts {
		body[x] = ext.Name()
		body[x] += " " + ext.Params()
	}

	if !recv.isAuthenticated() {
		body = append(body, "AUTH PLAIN")
	}

	err = recv.sendReply(StatusOK, srv.Env.PrimaryDomain.Name, body)
	if err != nil {
		return err
	}

	recv.state = cmd.Kind

	return nil
}

func (srv *Server) handleEXPN(recv *receiver, cmd *Command) (err error) {
	if recv.mode == receiverModeServer {
		return recv.sendError(errCmdUnknown)
	}

	if !recv.isAuthenticated() {
		err = recv.sendError(errNotAuthenticated)
		if err != nil {
			return err
		}
	}

	res, err := srv.Handler.ServeExpand(cmd.Arg)
	if err != nil {
		return err
	}

	err = recv.sendReply(res.Code, res.Message, res.Body)

	return err
}

func (srv *Server) handleMAIL(recv *receiver, cmd *Command) (err error) {
	if !recv.isAuthenticated() {
		return recv.sendError(errNotAuthenticated)
	}

	recv.mail.From = cmd.Arg

	err = recv.sendReply(StatusOK, "OK", nil)
	if err != nil {
		return err
	}

	recv.state = CommandMAIL

	return nil
}

func (srv *Server) handleHELP(recv *receiver) (err error) {
	return recv.sendReply(StatusHelp, "Everything will be alright", nil)
}

func (srv *Server) handleRCPT(recv *receiver, cmd *Command) (err error) {
	if !recv.isAuthenticated() {
		err = recv.sendError(errNotAuthenticated)
		if err != nil {
			return err
		}
	}

	recv.mail.Recipients = append(recv.mail.Recipients, cmd.Arg)

	// RFC 5321, 4.5.3.1.8.  Recipients Buffer
	if len(recv.mail.Recipients) > 100 {
		err = recv.sendReply(StatusNoStorage,
			"Too many recipients", nil)
	} else {
		err = recv.sendReply(StatusOK, "OK", nil)
	}
	if err != nil {
		return err
	}

	recv.state = CommandRCPT

	return nil
}

func (srv *Server) handleVRFY(recv *receiver, cmd *Command) (err error) {
	if recv.mode == receiverModeServer {
		return recv.sendError(errCmdUnknown)
	}

	if !recv.isAuthenticated() {
		err = recv.sendError(errNotAuthenticated)
		if err != nil {
			return err
		}
	}

	res, err := srv.Handler.ServeVerify(cmd.Arg)
	if err != nil {
		return err
	}

	err = recv.sendReply(res.Code, res.Message, res.Body)

	return err
}

//
// initialize handler, storage, extensions, and listeners.
//
func (srv *Server) initialize() (err error) {
	if srv.Env == nil {
		return fmt.Errorf("smtp: server environment is not defined")
	}
	if srv.Env.PrimaryDomain == nil {
		return fmt.Errorf("smtp: server primary domain is not defined")
	}

	if srv.Handler == nil {
		srv.Handler = NewLocalHandler(srv.Env)
	}

	if srv.Exts == nil {
		srv.Exts = defaultExts
	} else {
		srv.Exts = append(srv.Exts, defaultExts...)
	}

	err = srv.initListener()
	if err != nil {
		return err
	}

	srv.mailTxQueue = make(chan *MailTx, 512)
	srv.bounceQueue = make(chan *MailTx, 512)
	srv.relayQueue = make(chan *MailTx, 512)

	return nil
}

func (srv *Server) initListener() (err error) {
	if len(srv.address) == 0 {
		srv.address = ":25"
	}

	addr, err := net.ResolveTCPAddr("tcp", srv.address)
	if err != nil {
		return err
	}

	srv.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	if srv.TLSCert == nil {
		return fmt.Errorf("smtp: server certificate is not defined")
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{
			*srv.TLSCert,
		},
		MinVersion: tls.VersionTLS11,
	}

	if len(srv.tlsAddress) == 0 {
		srv.tlsAddress = ":465"
	}

	srv.tlsListener, err = tls.Listen("tcp", srv.tlsAddress, tlsCfg)

	return err
}

func (srv *Server) isOurDomain(d string) bool {
	if d == srv.Env.PrimaryDomain.Name {
		return true
	}
	for _, domain := range srv.Env.VirtualDomains {
		if d == domain.Name {
			return true
		}
	}
	return false
}

//
// processMailTxQueue process incoming mail transactions.
//
// There are three possibilities for incoming mail:
//
// (1) when the recipient domain is managed by server, the mail will be
// forwarded to handler ServeMailTx; the email transaction is terminated.
//
// (2) when the recipient domain is not managed by server, the mail will be
// relayed to another server based on recipient's domain address.
//
// (3) when recipient address is unknown or invalid, the mail transaction will
// be bounced back to sender.
//
func (srv *Server) processMailTxQueue() {
	for mail, ok := <-srv.mailTxQueue; ok; {
		if mail.isPostponed() {
			continue
		}

		// At this point, only one recipient exist in mail object.
		rcpt := mail.Recipients[0]
		addr := strings.Split(rcpt, "@")

		var err error

		switch len(addr) {
		case 2:
			if srv.isOurDomain(addr[1]) {
				// This is the first case.
				_, err = srv.Handler.ServeMailTx(mail)
			} else {
				// This is the second case.
				srv.relayQueue <- mail
			}
		case 1:
			// The first case, where recipient domain is assumed
			// to be primary domain.
			if addr[0] == localPostmaster {
				_, err = srv.Handler.ServeMailTx(mail)
			} else {
				srv.bounceQueue <- mail
			}
		default:
			srv.bounceQueue <- mail
		}

		if err != nil {
			if mail.Retry < 5 {
				mail.postpone()
			} else {
				srv.bounceQueue <- mail
			}
		}
	}
	srv.wg.Done()
}

//
// processBounceQueue send the mail back to reverse-path (sender).
//
// If sender domain is one of ours, call the handler; otherwise send them
// using SMTP through relay queue.
//
func (srv *Server) processBounceQueue() {
	var err error

	for mail, ok := <-srv.bounceQueue; ok; {
		_, err = srv.Handler.ServeBounce(mail)
		if err != nil {
			continue
		}
	}
	srv.wg.Done()
}

//
// processRelayQueue send mail to other MTA or final destination.
// A mail transaction will be relayed on the following conditions: the
// domain's name in MAIL FROM is managed by server and the recipient domain's
// address is not managed by server.
//
func (srv *Server) processRelayQueue() {
	for _, ok := <-srv.relayQueue; ok; {
		// TODO:
	}
	srv.wg.Done()
}

//
// processMailTx process mail transaction by breaking down recipients into one
// mail object, storing it into storage, and push it to the queue for further
// processing.
//
func (srv *Server) processMailTx(mail *MailTx) {
	mails := make([]*MailTx, len(mail.Recipients))
	for x, rcpt := range mail.Recipients {
		mails[x] = NewMailTx(mail.From, []string{rcpt}, mail.Data)
	}
	for _, mail := range mails {
		srv.mailTxQueue <- mail
	}
}
