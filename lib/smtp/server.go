// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
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

	// l a socket that listen for new connection from client.
	l *net.TCPListener

	// mailTxQueue hold mail objects before being relayed or stored.
	mailTxQueue chan *MailTx

	// bounceQueue hold mail objects with invalid recipient where it
	// will be notified to sender.
	bounceQueue chan *MailTx

	// relayQueue hold mail objects that need to be relayed to other MTA.
	relayQueue chan *MailTx
}

//
// ListenAndServe start listening the SMTP request on port 25.
// Each client connection will be handled in a single routine.
//
func (s *Server) ListenAndServe() (err error) {
	err = s.init()
	if err != nil {
		return
	}

	go s.processRelayQueue()
	go s.processBounceQueue()
	go s.processMailTxQueue()

	for {
		conn, err := s.l.AcceptTCP()
		if err != nil {
			log.Printf("ListenAndServe.AcceptTCP: %s", err)
			break
		}

		recv := newReceiver(conn)

		go s.handle(recv)
	}

	eClose := s.l.Close()
	if eClose != nil {
		log.Printf("ListenAndServe.Close: %s", eClose)
	}

	return
}

//
// handle receiver connection.
//
func (s *Server) handle(recv *receiver) {
	err := recv.sendReply(StatusReady, s.Env.Hostname(), nil)
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

		for _, ext := range s.Exts {
			err = ext.ValidateCommand(cmd)
			if err != nil {
				break
			}
		}
		if err != nil {
			_ = recv.sendError(err)
			continue
		}

		err = s.handleCommand(recv, cmd)
		if err != nil {
			log.Println("Server.handleCommand: ", err.Error())
			break
		}

		switch recv.state {
		case CommandDATA:
			err = s.processMailTx(recv.mail)
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
func (s *Server) handleCommand(recv *receiver, cmd *Command) (err error) {
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

		body := make([]string, len(s.Exts))
		for x, ext := range s.Exts {
			body[x] = ext.Name()
		}

		err = recv.sendReply(StatusOK, s.Env.Hostname(), body)
		if err != nil {
			return err
		}
		recv.state = cmd.Kind

	case CommandHELO:
		recv.clientDomain = cmd.Arg

		err = recv.sendReply(StatusOK, s.Env.Hostname(), nil)
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
		res, err := s.Handler.ServeVerify(cmd.Arg)
		if err != nil {
			return err
		}
		err = recv.sendReply(res.Code, res.Message, res.Body)
		if err != nil {
			return err
		}

	case CommandEXPN:
		res, err := s.Handler.ServeExpand(cmd.Arg)
		if err != nil {
			return err
		}
		err = recv.sendReply(res.Code, res.Message, res.Body)
		if err != nil {
			return err
		}

	case CommandHELP:
		err = s.handleHELP(recv, cmd.Arg)
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

func (s *Server) handleHELP(recv *receiver, arg string) (err error) {
	return recv.sendReply(StatusHelp, "Everything will be alright", nil)
}

//
// init initiliazer environment, handler, extensions, and connection listener.
//
func (s *Server) init() (err error) {
	if len(s.Addr) == 0 {
		s.Addr = ":25"
	}

	if s.Env == nil {
		s.Env, err = NewEnvironmentIni("")
		if err != nil {
			return
		}
	}

	if s.Handler == nil {
		s.Handler = &HandlerPosix{}
	}

	if s.Storage == nil {
		s.Storage, err = NewStorageFile("")
		if err != nil {
			return
		}
	}

	if s.Exts == nil {
		s.Exts = defaultExts
	} else {
		s.Exts = append(s.Exts, defaultExts...)
	}

	if s.l == nil {
		addr, err := net.ResolveTCPAddr("tcp", s.Addr)
		if err != nil {
			return err
		}

		s.l, err = net.ListenTCP("tcp", addr)
		if err != nil {
			return err
		}
	}

	s.mailTxQueue = make(chan *MailTx, 512)
	s.bounceQueue = make(chan *MailTx, 512)
	s.relayQueue = make(chan *MailTx, 512)

	return nil
}

func (s *Server) isLocalDomain(d string) bool {
	for _, domain := range s.Env.Domains() {
		if d == domain {
			return true
		}
	}
	return false
}

func (s *Server) processMailTxQueue() {
	for mail := range s.mailTxQueue {
		if mail.isPostponed() {
			continue
		}

		// At this point, only one recipient exist in mail object.
		rcpt := mail.Recipients[0]
		addr := strings.Split(rcpt, "@")

		var err error

		switch len(addr) {
		case 2:
			if s.isLocalDomain(addr[1]) {
				_, err = s.Handler.ServeMailTx(mail)
			} else {
				s.relayQueue <- mail
			}
		case 1:
			if addr[0] == localPostmaster {
				_, err = s.Handler.ServeMailTx(mail)
			} else {
				s.bounceQueue <- mail
			}
		default:
			s.bounceQueue <- mail
		}

		if err != nil {
			if mail.Retry < 5 {
				mail.postpone()
			} else {
				s.bounceQueue <- mail
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
func (s *Server) processBounceQueue() {
	for mail := range s.bounceQueue {
		err := s.Storage.Bounce(mail.ID)
		if err != nil {
			continue
		}
	}
}

//
// processRelayQueue send mail to other MTA.
//
func (s *Server) processRelayQueue() {
	for range s.relayQueue {
		// TODO:
	}
}

//
// processMailTx process mail transaction by breaking down recipients into one
// mail object, storing it into storage, and push it to the queue for further
// processing.
//
func (s *Server) processMailTx(mail *MailTx) (err error) {
	mails := make([]*MailTx, len(mail.Recipients))
	for x, rcpt := range mail.Recipients {
		mails[x] = NewMailTx(mail.From, []string{rcpt}, mail.Data)

		err = s.Storage.Store(mails[x])
		if err != nil {
			return
		}
	}
	for _, mail := range mails {
		s.mailTxQueue <- mail
	}

	return nil
}
