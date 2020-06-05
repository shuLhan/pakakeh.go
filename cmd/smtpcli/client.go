// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/shuLhan/share/lib/smtp"
)

var (
	stdin  = os.Stdin
	stdout = os.Stdout
	noop   = []byte("NOOP\r\n")
)

type client struct {
	con       *smtp.Client
	input     []byte
	remoteURL string
	mailTx    *smtp.MailTx
}

func newClient(remoteURL string) (cli *client, err error) {
	cli = &client{
		remoteURL: remoteURL,
		input:     make([]byte, 0, 128),
	}
	if len(remoteURL) == 0 {
		return
	}

	cli.con, err = smtp.NewClient("", remoteURL, false)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func (cli *client) handleInput() (isQuit bool) {
	input := bytes.ToLower(bytes.TrimSpace(cli.input))

	var (
		res *smtp.Response
		err error
	)

	ins := bytes.Split(input, []byte{' '})
	cmd := string(ins[0])
	switch cmd {
	case "":
		res, err = cli.con.SendCommand(noop)

	case "starttls":
		res, err = cli.con.StartTLS()

	case "from":
		if len(ins) < 2 {
			usageFrom()
			return false
		}
		from := smtp.ParseMailbox(ins[1])
		if len(from) == 0 {
			log.Printf("Invalid mailbox: %s\n", ins[1])
			return false
		}
		cli.mailTx = &smtp.MailTx{
			From: string(from),
		}
		stdout.WriteString("OK\n")
		return false

	case "to":
		if len(ins) < 2 {
			usageTo()
			return false
		}
		to := smtp.ParseMailbox(ins[1])
		if len(to) == 0 {
			log.Printf("Invalid mailbox: %s\n", ins[1])
			return false
		}
		if cli.mailTx == nil {
			log.Println("Invalid sequence of command, missing 'from'")
			return false
		}
		cli.mailTx.Recipients = append(cli.mailTx.Recipients, string(to))
		stdout.WriteString("OK\n")
		return false

	case "data":
		if cli.mailTx == nil {
			log.Println("Invalid sequence of command, missing 'from' and 'to'")
			return false
		}
		if len(ins) < 2 {
			usageData()
			return false
		}
		err = cli.readData(string(ins[1]))
		if err != nil {
			break
		}
		stdout.WriteString("OK\n")

		res, err = cli.send()

	case "quit":
		res, err = cli.con.Quit()
		isQuit = true

	default:
		input = append(input, "\r\n"...)
		res, err = cli.con.SendCommand(input)
	}
	if err != nil {
		log.Println(err)
	}
	if res != nil {
		fmt.Printf("Response < %+v\n", res)
	}

	return isQuit
}

func (cli *client) readData(fin string) (err error) {
	cli.mailTx.Data, err = ioutil.ReadFile(fin)
	if err != nil {
		return err
	}

	return nil
}

func (cli *client) readInput() (err error) {
	cli.input = cli.input[:0]
	c := make([]byte, 1)
	for {
		_, err = stdin.Read(c)
		if err != nil {
			return
		}
		if c[0] == '\n' {
			break
		}
		cli.input = append(cli.input, c[0])
	}
	return nil
}

func (cli *client) send() (res *smtp.Response, err error) {
	fmt.Fprintf(stdout, "From: %s\n", cli.mailTx.From)
	fmt.Fprintf(stdout, "Recipients: %s\n", cli.mailTx.Recipients)
	fmt.Fprintf(stdout, "Data:\n%s\n", cli.mailTx.Data)

	res, err = cli.con.MailTx(cli.mailTx)

	cli.mailTx = nil

	return res, err
}

func (cli *client) writePrompt() {
	if cli.con == nil {
		stdout.WriteString("(disconnected)> ")
	} else {
		stdout.WriteString(cli.remoteURL + "> ")
	}
}

func usageData() {
	stdout.WriteString("usage: 'data <path-to-file>'\n")
}

func usageFrom() {
	stdout.WriteString("usage: 'from mailbox'\n")
}

func usageTo() {
	stdout.WriteString("usage: 'to mailbox'\n")
}
