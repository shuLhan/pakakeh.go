// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

// Package dkim provide a library to parse and create DKIM-Signature header
// field value, as defined in RFC 6376, DomainKeys Identified Mail (DKIM)
// Signatures.
//
// The process to signing and verying a message is handled by parent package
// "lib/email", not by this package.
package dkim
