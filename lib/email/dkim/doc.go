// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package dkim provide a library to parse and create DKIM-Signature header
// field value, as defined in RFC 6376, DomainKeys Identified Mail (DKIM)
// Signatures.
//
// The process to signing and verying a message is handled by parent package
// "lib/email", not by this package.
//
package dkim
