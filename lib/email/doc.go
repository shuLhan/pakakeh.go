// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package email provide a library for working with Internet Message Format as
// defined in [RFC 5322].
//
// The building block of a [Message].
//
//	+---------------+
//	| Message       |
//	+--------+------+
//	| Header | Body |
//	+--------+------+
//
// A [Body] contains one or more [MIME],
//
//	+------------------+
//	| MIME             |
//	+--------+---------+
//	| Header | Content |
//	+--------+---------+
//
// A [Header] contains one or more [Field],
//
//	+---------------------+
//	| Field               |
//	+------+-------+------+
//	| Name | Value | Type |
//	+------+-------+------+
//
// [Field] is parsed line that contains Name and Value separated by colon ':'.
//
// A [ContentType] is special Field where Name is "Content-Type", and its
// Value is parsed from string "top/sub; <param>; ...".
// A ContentType can contains zero or more [Param], each separated by ";".
//
//	+-------------------+
//	| ContentType       |
//	+-----+-----+-------+
//	| Top | Sub | Param |
//	+-----+-----+-------+
//
// A [Param] is parsed string of key and value separated by "=", where value
// can be quoted, for example `key=value` or `key="quoted value"`.
//
//	+----------------------+
//	| Param                |
//	+-----+-------+--------+
//	| Key | Value | Quoted |
//	+-----+-------+--------+
//
// # Notes
//
// In the comment and/or methods of some type, you will see the word "simple"
// or "relaxed".
// This method only used for message signing and verification using DKIM
// (see [Message.DKIMSign] and [Message.DKIMVerify] implementation).
// In short, "simple" return the formatted header or body as is, while
// "relaxed" return the formatted header or body by trimming the space.
// See [RFC 6376 Section 3.4] or [our summary].
//
// [our summary]: https://git.sr.ht/~shulhan/pakakeh.go/blob/master/_doc/RFC_6376__DKIM_SIGNATURES.adoc#canonicalization
// [RFC 5322]: https://datatracker.ietf.org/doc/html/rfc5322
// [RFC 6376 Section 3.4]: https://datatracker.ietf.org/doc/html/rfc6376#section-3.4
package email
