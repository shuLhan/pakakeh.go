// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package smtp provide a library for building SMTP server and client.
//
// Limitations
//
// The server favor implicit TLS over STARTTLS (RFC 8314).
// When server's environment is configured with certificate, server will
// listen on port 465.
//
package smtp
