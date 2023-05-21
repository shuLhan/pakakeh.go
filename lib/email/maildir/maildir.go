// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package maildir provide a library to manage message (email), and its
// folder, using maildir format.
//
// This library use the following file name format for tmp message,
//
//	epoch ".M" usec "_P" pid "_Q" counter "." hostname
//
// and the following format for new message,
//
//	epoch ".M" usec "_P" pid "_V" dev "_I" inode "_Q" counter "." hostname ",S=" size ":2"
//
// The epoch is Unix timestamp--number of elapsed seconds,
// usec is the 6 digits of micro seconds,
// pid is the process ID of the program,
// dev is the file device number,
// inode is the file inode number,
// counter is a auto increment number start from 0,
// hostname is the system host name, and
// size is the message size.
//
// References,
//
//   - [Courier Maildir]
//   - [Courier Maildir++]
//   - [Dovecot Maildir]
//   - [Maildir]
//   - [Qmail Maildir manual]
//
// [Courier Maildir]: https://www.courier-mta.org/maildir.html
// [Courier Maildir++]: http://www.courier-mta.org/imap/README.maildirquota.html
// [Dovecot Maildir]: https://doc.dovecot.org/admin_manual/mailbox_formats/maildir/
// [Maildir]: https://cr.yp.to/proto/maildir.html
// [Qmail Maildir manual]: http://qmail.org/qmail-manual-html/man5/maildir.html
package maildir

import (
	"os"
	"syscall"
	"time"
)

// osGetpid variable for mocking os.Getpid during testing.
var osGetpid = os.Getpid

// osHostname variable for mocking os.Hostname during testing.
var osHostname = os.Hostname

// syscallStat define a variable that can be mocked in testing to provide
// predictable syscall.Stat value.
var syscallStat = syscall.Stat

// timeNow define a variable that can be mocked in testing to provide
// predictable time.
var timeNow = time.Now
