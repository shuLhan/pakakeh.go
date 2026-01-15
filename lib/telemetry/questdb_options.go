// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import (
	"fmt"
	"net"
	"net/url"
	"time"

	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

const (
	defQuestdbPort    = 9009
	defQuestdbTimeout = 10 * time.Second
)

// QuestdbOptions options for QuestdbForwarder.
type QuestdbOptions struct {
	// Fmt the Formatter to use to convert the Metric.
	// Usually set to IlpFormatter.
	Fmt Formatter

	// ServerURL define the QuestDB server URL.
	// Currently, it only support the TCP scheme using the following
	// format "tcp://<host>:<port>".
	ServerURL string

	proto   string
	address string

	// Timeout define default timeout for Write.
	// Default to 10 seconds.
	Timeout time.Duration
}

func (opts *QuestdbOptions) init() (err error) {
	var surl *url.URL

	surl, err = url.Parse(opts.ServerURL)
	if err != nil {
		return err
	}
	if len(surl.Scheme) == 0 {
		surl.Scheme = `tcp`
	}
	opts.proto = surl.Scheme

	var (
		ip   net.IP
		port uint16
	)

	opts.address, ip, port = libnet.ParseIPPort(surl.Host, defQuestdbPort)
	if len(opts.address) == 0 {
		opts.address = fmt.Sprintf(`%s:%d`, ip, port)
	} else {
		opts.address = fmt.Sprintf(`%s:%d`, opts.address, port)
	}

	if opts.Timeout <= 0 {
		opts.Timeout = defQuestdbTimeout
	}
	return nil
}
