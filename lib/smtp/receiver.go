// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	liberrors "github.com/shuLhan/share/lib/errors"
)

type receiver struct {
	conn            net.Conn
	data            []byte
	buff            bytes.Buffer
	state           CommandKind
	clientDomain    string
	clientAddress   string
	localAddress    string
	mail            *MailTx
	isAuthenticated bool
}

func newReceiver(conn net.Conn) (recv *receiver) {
	recv = &receiver{
		conn: conn,
		data: make([]byte, 4096),
		mail: &MailTx{},
	}

	recv.clientAddress = conn.RemoteAddr().String()
	recv.localAddress = conn.LocalAddr().String()

	return recv
}

//
// close the receiving line.
//
func (recv *receiver) close() {
	err := recv.conn.Close()
	if err != nil {
		log.Printf("receiver.close: %s\n", err)
	}
}

//
// readAuthData read AUTH initial response from client into Command Param.
//
func (recv *receiver) readAuthData(cmd *Command) (err error) {
	recv.buff.Reset()

	for {
		recv.data = recv.data[0:]
		n, err := recv.conn.Read(recv.data)
		if n > 0 {
			_, _ = recv.buff.Write(recv.data[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if n == cap(recv.data) {
			continue
		}
		break
	}

	cmd.Param = strings.TrimSpace(recv.buff.String())

	return nil
}

//
// readCommand from client.
//
// Any error from command line (for example, unknown command, or syntax error)
// will be handled directly by this function by replying to client.
//
// An error returned from this function, MUST be considered error on system
// which should stop the receiver for further processing.
//
func (recv *receiver) readCommand() (cmd *Command, err error) {
	recv.buff.Reset()

	cmd = newCommand()

	for {
		recv.data = recv.data[0:]
		n, err := recv.conn.Read(recv.data)
		if n > 0 {
			_, _ = recv.buff.Write(recv.data[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if n == cap(recv.data) {
			continue
		}
		break
	}

	err = cmd.unpack(recv.buff.Bytes())
	if err != nil {
		return nil, err
	}

	return cmd, nil
}

//
// readDATA start mail input.
//
func (recv *receiver) readDATA() (err error) {
	for {
		recv.data = recv.data[0:]
		n, err := recv.conn.Read(recv.data)
		if err != nil {
			return err
		}

		recv.mail.Data = append(recv.mail.Data, recv.data[:n]...)

		if recv.mail.isTerminated() {
			break
		}
	}

	l := len(recv.mail.Data)

	// Remove the end-of-mail data indicator.
	recv.mail.Data = recv.mail.Data[:l-5]

	recv.mail.seal(recv.clientDomain, recv.clientAddress, recv.localAddress)

	return nil
}

func (recv *receiver) reset() {
	recv.state = CommandZERO
	recv.mail.Reset()
}

func (recv *receiver) sendError(errRes error) (err error) {
	reply, ok := errRes.(*liberrors.E)
	if !ok {
		reply.Code = StatusLocalError
		reply.Message = errRes.Error()
	} else if reply.Code == 0 {
		reply.Code = StatusLocalError
	}

	_, err = fmt.Fprintf(recv.conn, "%d %s\r\n", reply.Code, reply.Message)
	if err != nil {
		log.Println("sendError: ", err.Error())
		return err
	}

	recv.reset()

	return nil
}

//
// sendReply send single or multiple lines reply to client.
//
// An error returned from this function, MUST be considered error on system
// which should stop the receiver for further processing.
//
func (recv *receiver) sendReply(code int, msg string, body []string) (err error) {
	recv.buff.Reset()
	if len(body) == 0 {
		_, err = fmt.Fprintf(&recv.buff, "%d %s\r\n", code, msg)
	} else {
		_, err = fmt.Fprintf(&recv.buff, "%d-%s\r\n", code, msg)
	}
	if err != nil {
		return
	}

	for x, line := range body {
		if x == len(body)-1 {
			_, err = fmt.Fprintf(&recv.buff, "%d %s\r\n", code, line)
		} else {
			_, err = fmt.Fprintf(&recv.buff, "%d-%s\r\n", code, line)
		}
		if err != nil {
			return
		}
	}

	_, err = recv.conn.Write(recv.buff.Bytes())

	return err
}
