// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry

import (
	"fmt"
	"net"
	"time"
)

// QuestdbForwarder forward the metrics to [QuestDB] using TCP.
//
// [QuestDB]: https://questdb.io
type QuestdbForwarder struct {
	conn net.Conn
	opts QuestdbOptions
}

// NewQuestdbForwarder create new forwarder for QuestDB.
func NewQuestdbForwarder(opts QuestdbOptions) (fwd *QuestdbForwarder, err error) {
	var logp = `NewQuestdbForwarder`

	err = opts.init()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	fwd = &QuestdbForwarder{
		opts: opts,
	}

	fwd.conn, err = net.DialTimeout(opts.proto, opts.address, opts.Timeout)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return fwd, nil
}

// Close the connection to the questdb server.
func (fwd *QuestdbForwarder) Close() (err error) {
	if fwd.conn == nil {
		return nil
	}

	var logp = `QuestdbForwarder.Close`

	err = fwd.conn.Close()
	if err != nil {
		fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

// Formatter return the Formatter used by questdb.
func (fwd *QuestdbForwarder) Formatter() Formatter {
	return fwd.opts.Fmt
}

// Write forward the formatted Metric into the questdb server.
func (fwd *QuestdbForwarder) Write(b []byte) (n int, err error) {
	var (
		logp = `QuestdbForwarder.Write`
		now  = time.Now()
	)

	err = fwd.conn.SetWriteDeadline(now.Add(5 * time.Second))
	if err != nil {
		return 0, fmt.Errorf(`%s: SetWriteDeadline: %s`, logp, err)
	}

	_, err = fwd.conn.Write(b)
	if err != nil {
		return 0, fmt.Errorf(`%s: %s`, logp, err)
	}

	return n, nil
}
