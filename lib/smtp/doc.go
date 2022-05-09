// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package smtp provide a library for building SMTP server and client.
//
// # Server
//
// By default, server will listen on port 25 and 465.
//
// Port 25 is only used to receive message relay from other mail server.  Any
// command that require authentication will be rejected by this port.
//
// Port 465 is used to receive message submission from SMTP accounts with
// authentication.
//
// # Server Environment
//
// The server require one primary domain with one primary account called
// "postmaster".  Domain can have two or more accounts.  Domain can have
// their own DKIM certificate.
//
// # Limitations
//
// The server favor implicit TLS over STARTTLS (RFC 8314) on port 465 for
// message submission.
package smtp
