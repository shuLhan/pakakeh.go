// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/shuLhan/share/lib/debug"
)

const (
	localPostmaster = "postmaster"
)

//
// Server defines parameters for running an SMTP server.
//
type Server struct {
	// Addr to listen for incoming connections.
	Addr string

	//
	// Env define the environment of SMTP server.  Default environment is
	// EnvironmentIni, which read configuration through ini formated file.
	//
	Env Environment

	//
	// Exts define list of custom extensions that the server will provide.
	//
	Exts []Extension

	//
	// Handler define an interface that will process the bouncing email,
	// incoming email, EXPN command, and VRFY command.
	//
	Handler Handler

	//
	// Storage define the storage that will be used to load and store
	// email.  Default Storage is StorageFile, where incoming email will
	// be stored on file system.
	//
	Storage Storage

	// listener is a socket that listen for new connection from client.
	listener net.Listener

	// mailTxQueue hold mail objects before being relayed or stored.
	mailTxQueue chan *MailTx

	// bounceQueue hold mail objects with invalid recipient where it
	// will be notified to sender.
	bounceQueue chan *MailTx

	// relayQueue hold mail objects that need to be relayed to other MTA.
	relayQueue chan *MailTx
}

//
// ListenAndServe start listening the SMTP request.
// Each client connection will be handled in a single routine.
//
func (srv *Server) ListenAndServe() (err error) {
	err = srv.init()
	if err != nil {
		return
	}

	go srv.processRelayQueue()
	go srv.processBounceQueue()
	go srv.processMailTxQueue()

	for {
		fmt.Println("ListenAndServe: waiting for client ...")
		conn, err := srv.listener.Accept()
		if err != nil {
			log.Printf("ListenAndServe.Accept: %s", err)
			break
		}

		recv := newReceiver(conn)

		go srv.handle(recv)
	}

	err = srv.listener.Close()
	if err != nil {
		log.Printf("ListenAndServe.Close: %s", err)
	}

	return
}

//
// handle receiver connection.
//
func (srv *Server) handle(recv *receiver) {
	err := recv.sendReply(StatusReady, srv.Env.Hostname(), nil)
	if err != nil {
		log.Println("receiver.sendReply: ", err.Error())
		recv.close()
		return
	}

	for {
		cmd, err := recv.readCommand()
		if err != nil {
			log.Println("receiver.readCommand: ", err)
			recv.sendError(err)
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
			err = srv.processMailTx(recv.mail)
			if err != nil {
				log.Println("server.processMailTx: ", err.Error())
				err = recv.sendError(errInProcessing)
				if err != nil {
					goto out
				}
				continue
			}

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
//
func (srv *Server) handleCommand(recv *receiver, cmd *Command) (err error) {
	if debug.Value > 0 {
		log.Printf("handleCommand: %v\n", cmd)
	}

	switch cmd.Kind {
	case CommandDATA:
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

	case CommandEHLO:
		recv.clientDomain = cmd.Arg

		body := make([]string, len(srv.Exts))
		for x, ext := range srv.Exts {
			body[x] = ext.Name()
		}

		err = recv.sendReply(StatusOK, srv.Env.Hostname(), body)
		if err != nil {
			return err
		}
		recv.state = cmd.Kind

	case CommandHELO:
		recv.clientDomain = cmd.Arg

		err = recv.sendReply(StatusOK, srv.Env.Hostname(), nil)
		if err != nil {
			return err
		}
		recv.state = cmd.Kind

	case CommandMAIL:
		recv.mail.From = cmd.Arg
		err = recv.sendReply(StatusOK, "OK", nil)
		if err != nil {
			return err
		}
		recv.state = CommandMAIL

	case CommandRCPT:
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

	case CommandRSET:
		recv.reset()

		err = recv.sendReply(StatusOK, "OK", nil)
		if err != nil {
			return err
		}

	case CommandVRFY:
		res, err := srv.Handler.ServeVerify(cmd.Arg)
		if err != nil {
			return err
		}
		err = recv.sendReply(res.Code, res.Message, res.Body)
		if err != nil {
			return err
		}

	case CommandEXPN:
		res, err := srv.Handler.ServeExpand(cmd.Arg)
		if err != nil {
			return err
		}
		err = recv.sendReply(res.Code, res.Message, res.Body)
		if err != nil {
			return err
		}

	case CommandHELP:
		err = srv.handleHELP(recv, cmd.Arg)
		if err != nil {
			return err
		}

	case CommandNOOP:
		err = recv.sendReply(StatusOK, "OK", nil)
		if err != nil {
			return err
		}

	case CommandQUIT:
		_ = recv.sendReply(StatusClosing,
			"Service closing transmission channel", nil)
		recv.state = CommandQUIT
	}

	return nil
}

func (srv *Server) handleHELP(recv *receiver, arg string) (err error) {
	return recv.sendReply(StatusHelp, "Everything will be alright", nil)
}

//
// init initiliaze environment, handler, extensions, and connection listener.
//
func (srv *Server) init() (err error) {
	if srv.Env == nil {
		srv.Env, err = NewEnvironmentIni("")
		if err != nil {
			return
		}
	}

	if srv.Handler == nil {
		srv.Handler = &HandlerPosix{}
	}

	if srv.Storage == nil {
		srv.Storage, err = NewStorageFile("")
		if err != nil {
			return
		}
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
	cert := srv.Env.Certificate()
	if cert == nil {
		if len(srv.Addr) == 0 {
			srv.Addr = ":25"
		}
	} else {
		if len(srv.Addr) == 0 {
			srv.Addr = ":465"
		}
	}

	addr, err := net.ResolveTCPAddr("tcp", srv.Addr)
	if err != nil {
		return err
	}

	if cert == nil {
		srv.listener, err = net.ListenTCP("tcp", addr)
	} else {
		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{
				*cert,
			},
			MinVersion: tls.VersionTLS11,
		}
		srv.listener, err = tls.Listen("tcp", srv.Addr, tlsCfg)
	}
	if err != nil {
		return err
	}

	return nil
}

func (srv *Server) isLocalDomain(d string) bool {
	for _, domain := range srv.Env.Domains() {
		if d == domain {
			return true
		}
	}
	return false
}

func (srv *Server) processMailTxQueue() {
	for mail := range srv.mailTxQueue {
		if mail.isPostponed() {
			continue
		}

		// At this point, only one recipient exist in mail object.
		rcpt := mail.Recipients[0]
		addr := strings.Split(rcpt, "@")

		var err error

		switch len(addr) {
		case 2:
			if srv.isLocalDomain(addr[1]) {
				_, err = srv.Handler.ServeMailTx(mail)
			} else {
				srv.relayQueue <- mail
			}
		case 1:
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
}

//
// processBounceQueue send the mail back to reverse-path (sender).
//
// If sender domain is one of ours, call the handler; otherwise send them
// using SMTP through relay queue.
//
func (srv *Server) processBounceQueue() {
	for mail := range srv.bounceQueue {
		err := srv.Storage.Bounce(mail.ID)
		if err != nil {
			continue
		}
	}
}

//
// processRelayQueue send mail to other MTA.
//
func (srv *Server) processRelayQueue() {
	for range srv.relayQueue {
		// TODO:
	}
}

//
// processMailTx process mail transaction by breaking down recipients into one
// mail object, storing it into storage, and push it to the queue for further
// processing.
//
func (srv *Server) processMailTx(mail *MailTx) (err error) {
	mails := make([]*MailTx, len(mail.Recipients))
	for x, rcpt := range mail.Recipients {
		mails[x] = NewMailTx(mail.From, []string{rcpt}, mail.Data)

		err = srv.Storage.Store(mails[x])
		if err != nil {
			return
		}
	}
	for _, mail := range mails {
		srv.mailTxQueue <- mail
	}

	return nil
}
