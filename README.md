<!--
SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

SPDX-License-Identifier: BSD-3-Clause
-->

# `import "git.sr.ht/~shulhan/pakakeh.go"`

[Go documentation](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go).

"pakakeh.go" is a collection of tools, public HTTP APIs, and libraries
written and for working with Go programming language.

This Go module usually released every month, at the first week of the month.

## Public APIs

[**slack**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/api/slack)::
Package slack provide a simple API for sending message to Slack using
only standard packages.

[**telegram/bot**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/api/telegram/bot)::
Package bot implement the
[Telegram Bot API](https://core.telegram.org/bots/api).


## Command Line Interface

[**ansua**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/ansua)::
The ansua command run a timer on defined duration and optionally run a
command when timer finished.

[**bcrypt**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/bcrypt)::
CLI to compare or generate hash using bcrypt.

[**emaildecode**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/emaildecode)::
Program emaildecode convert the email body from quoted-printable to plain
text.

[**epoch**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/epoch)::
Program epoch print the current date and time (Unix seconds, milliseconds,
nanoseconds, local time, and UTC time) or the date and time based on the
epoch on first parameter.

[**gofmtcomment**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/gofmtcomment)::
Program to convert multi lines "/**/" comments into single line "//" format.

[**httpdfs**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/httpdfs)::
Program httpdfs implement [libhttp.Server] with [memfs.MemFS].

[**ini**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/ini)::
Program ini provide a command line interface to get and set values in the
[INI file format](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ini).

[**sendemail**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/sendemail)::
Program sendemail is command line interface that use lib/email and
lib/smtp to send email.

[**smtpcli**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/smtpcli)::
Command line interface SMTP client protocol.
This is an example of implementation Client from
[lib/smtp](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/smtp).

[**totp**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/totp)::
Program to generate Time-based One-time Password using secret key.
This is just an example of implementation of
[lib/totp](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/totp).
See
<https://kilabit.info/project/gotp/> for a complete implementation that
support encryption.

[**xtrk**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/xtrk)::
Program xtrk is command line interface to uncompress and/or unarchive a
file.
Supported format: bzip2, gzip, tar, zip, tar.bz2, tar.gz.


## Libraries

[**ascii**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ascii)::
A library for working with ASCII characters.

[**binary**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/binary)::
Package binary complement the standard [binary] package.
Currently it implement append-only binary that encode the data using
binary.Writer.
We call them "Apo" for short.

[**bytes**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/bytes)::
A library for working with slice of bytes.

[**clise**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/clise)::
Package clise implements circular slice.


[**contact**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/contact)::
A library to import contact from Google, Microsoft, or Yahoo.

[**contact/google**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/contact/google)::
Package "contact/google" implement Google's contact API v3.

[**contact/microsoft**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/contact/microsoft)::
Package "contact/microsoft" implement Microsoft's Live contact API v1.0.

[**contact/vcard**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/contact/vcard)::
Package "contact/vcard" implement RFC6350 for encoding and decoding VCard
formatted data.

[**contact/yahoo**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/contact/yahoo)::
Package yahoo implement user's contacts import using Yahoo API.


[**crypto**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/crypto)::
Package crypto provide a wrapper to simplify working with standard crypto
package.

[**debug**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/debug)::
Package debug provide global debug variable, initialized through environment
variable "DEBUG" or directly.

[**dns**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/dns)::
A library for working with Domain Name System (DNS) protocol.

[**dsv**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/dsv)::
A library for working with delimited separated value (DSV).


[**email**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/email)::
A library for working with Internet Message Format, as defined in RFC 5322.

[**email/dkim**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/email/dkim)::
A library to parse and create DKIM-Signature header field value, as
defined in RFC 6376.

[**email/maildir**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/email/maildir)::
A library to manage email using maildir format.


[**errors**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/errors)::
Package errors provide an error type with Code, Message, and Name.

[**git**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/git)::
A wrapper for git command line interface.

[**hexdump**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/hexdump)::
Package hexdump implements reading and writing bytes from and into
hexadecimal number.
It support parsing output from hexdump(1) tool.

[**html**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/html)::
Package html extends the golang.org/x/net/html by providing simplified
methods for working with Node.


[**http**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/http)::
Package http extends the standard http package with simplified routing handler
and builtin memory file system.

[**http/sseclient**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/http/sseclient)::
Package sseclient implement HTTP client for Server-Sent Events (SSE).


[**ini**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ini)::
A library for reading and writing INI configuration as defined by Git
configuration file syntax.

[**json**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/json)::
Package json extends the capabilities of standard json package.


[**math**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/math)::
Package math provide generic functions working with math.

[**math/big**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/math/big)::
Package big extends the capabilities of standard "math/big" package by
adding custom global precision to Float, Int, and Rat, global rounding
mode, and custom bits precision to Float.


[**memfs**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/memfs)::
A library for mapping file system into memory and to generate an embedded Go
file from it.


[**mining**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining)::
A library for data mining.

[**mining/classifiers/cart**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier/cart)::
An implementation of the Classification and Regression Tree by Breiman, et al.

[**mining/classifier/crf**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining/classififer/crf)::
An implementation of the Cascaded Random Forest (CRF) algorithm, by Baumann,
Florian, et al.

[**mining/classifier/rf**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier/rf)::
An implementation of ensemble of classifiers using random forest algorithm by
Breiman and Cutler.

[**mining/gain/gini**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining/gain/gini)::
A library to calculate Gini gain.

[**mining/knn**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining/knn)::
An implementation of the K Nearest Neighbor (KNN) using Euclidian to
compute the distance between samples.

[**mining/resampling/lnsmote**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining/resampling/lnsmote)::
An implementation of the Local-Neighborhood algorithm from the paper of
Maciejewski, Tomasz, and Jerzy Stefanowski.

[**mining/resampling/smote**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining/resampling/smote)::
An implementation of the Synthetic Minority Oversampling TEchnique (SMOTE).

[**mining/tree/binary**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mining/tree/binary)::
An implementation of binary tree.


[**mlog**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/mlog)::
Package mlog implement buffered multi writers of log.

[**net**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/net)::
Constants and library for networking.

[**numbers**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/numbers)::
A library for working with integer, float, slice of integer, and slice of
floats.


[**os**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/os)::
Package os extend the standard os package to provide additional
functionalities.

[**os/exec**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/os/exec)::
Package exec wrap the standar package "os/exec" to simplify calling Run
with stdout and stderr.


[**paseto**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/paseto)::
A simple, ready to use, implementation of Platform-Agnostic SEcurity TOkens
(PASETO).

[**path**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/path)::
Package path implements utility routines for manipulating slash-separated
paths.

[**play**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/play)::
Package play provides callable APIs and HTTP handlers to format,
run, and test Go code, similar to Go playground but using HTTP instead of
WebSocket.

[**reflect**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/reflect)::
Package reflect extends the standard reflect package.

[**runes**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/runes)::
A library for working with slice of rune.

[**slices**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/slices)::
Package slices complement the standard slices package for working with
slices with comparable and [cmp.Ordered] types.

[**smtp**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/smtp)::
A library for building SMTP server or client. This package is working in
progress.

[**spf**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/spf)::
Package spf implement Sender Policy Framework (SPF) per RFC 7208.

[**sql**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/sql)::
Package sql extends the standard library "database/sql.DB" that provide common
functionality across DBMS.

[**ssh**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ssh)::
Package ssh provide a wrapper for golang.org/x/crypto/ssh and a parser for SSH
client configuration specification ssh_config(5).

[**ssh/sftp**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ssh/sftp)::
Package sftp implement native SSH File Transport Protocol v3.

[**sshconfig**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/sshconfig)::
Package config provide the ssh_config(5) parser and getter.

[**strings**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/strings)::
A library for working with slice of string.

[**tabula**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/tabula)::
A library for working with rows, columns, or matrix (table), or in another
terms working with data set.

[**telemetry**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/telemetry)::
Package telemetry is a library for collecting various [Metric], for example
from standard runtime/metrics, and send or write it to one or more
[Forwarder].


[**test**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/test)::
A library for helping with testing.

[**test/httptest**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/test/httptest)::
Package httptest implement testing HTTP package.

[**test/mock**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/test/mock)::
Package mock provide a mocking for standard output and standard error.


[**text**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/text)::
A library for working with text.

[**text/diff**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/text/diff)::
Package diff implement text comparison.


[**time**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/time)::
A library for working with time.

[**totp**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/totp)::
Package totp implement Time-Based One-Time Password Algorithm based on RFC
6238.


[**watchfs**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/watchfs)::
Package watchfs implement naive file and directory watcher.
This package is deprecated, we keep it here for historical only.
The new implementation should use "watchfs/v2".

[**watchfs/v2**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/watchfs/v2)::
Package watchfs implement naive file watcher.
The version 2 simplify watching single file and directory.
For directory watcher, it watch only one file instead of all included files.


[**websocket**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/websocket)::
The WebSocket library for server and client. This WebSocket library has
been tested with autobahn testsuite with 100% success rate.
[the status report](https://git.sr.ht/~shulhan/pakakeh.go/blob/main/lib/websocket/AUTOBAHN.adoc).

[**xmlrpc**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/xmlrpc)::
Package xmlrpc provide an implementation of
[XML-RPC specification](http://xmlrpc.com/spec.md).


## Simplified RFCs

### MIME

* [RFC 2045: MIME Part One - Format of Internet Message Bodies](RFC_2045__MIME_I_FORMAT.html).
* [RFC 2046: MIME Part Two - Media Types](RFC_2046__MIME_II_MEDIA_TYPES.html).
* [RFC 2049: MIME Part Five: Conformance Criteria and Examples](RFC_2049__MIME_V_CONFORMANCE.html).
* [RFC 5322: Internet Message Format](RFC_5322__IMF.html).


### SASL

* [RFC 4422: Simple Authentication and Security Layer](RFC_4422__SASL.html).
* [RFC 4616: The PLAIN Simple Authentication and Security Layer (SASL) Mechanism](RFC_4616__SASL_PLAIN.html).


### DKIM

* [RFC 4685: Analysis of Threats Motivating DKIM](RFC_4865__DKIM_THREATS.html).
* [RFC 5585: DomainKeys Identified Mail Service Overview](RFC_5585__DKIM_OVERVIEW.html).
* [RFC 5863: DKIM Development, Deployment, and Operations](RFC_5863__DKIM_DEVOPS.html).
* [RFC 6376: DKIM Signatures](RFC_6376__DKIM_SIGNATURES.html).


### SMTP

* [RFC 3207: SMTP Service Extension for Secure SMTP over Transport Layer Security](RFC_3207__ESMTP_TLS.html).
* [RFC 3461-3464: Delivery Status Notification](RFC_3461-3464__ESMTP_DSN.html).
* [RFC 4954: SMTP Service Extension for Authentication](RFC_4954__ESMTP_AUTH.html).
* [RFC 5321: Simple Mail Transfer Protocol](RFC_5321__SMTP.html).


### SPF

* [RFC 7208: Sender Policy Framework version 1](RFC_7808__SPFv1.html).


## DNS

* [RFC 6891: Extension Mechanisms for DNS (EDNS(0)](RFC_6891_EDNS0.html).

* [RFC 9460: Service Binding and Parameter Specification via the DNS
  (SVCB and HTTPS Resource Records)](RFC_9460__SVCB_and_HTTP_RR.html).


## Changelog

This library is released every month, usually at the first week of month.

[Latest changelog](CHANGELOG.html).

[Changelog in 2024](CHANGELOG_2024.html).
Changelog for `pakakeh.go` module since v0.52.0 until v0.58.1.

[Changelog in 2023](CHANGELOG_2023.html).
Changelog for `pakakeh.go` module since v0.43.0 until v0.51.0.

[Changelog in 2022](CHANGELOG_2022.html).
Changelog for `pakakeh.go` module since v0.33.0 until v0.42.0.

[Changelog in 2021](CHANGELOG_2021.html).
Changelog for `pakakeh.go` module since v0.22.0 until v0.32.0.

[Changelog in 2020](CHANGELOG_2020.html).
Changelog for `pakakeh.go` module since v0.12.0 until v0.21.0.

[Changelog from 2018 to 2019](CHANGELOG_2018-2019.html).
Changelog for `pakakeh.go` v0.1.0 until v0.11.0.


## Credits

[Autobahn testsuite](https://github.com/crossbario/autobahn-testsuite) for
testing WebSocket library.


##  Development

<https://git.sr.ht/~shulhan/pakakeh.go>::
Link to the source code.

<https://todo.sr.ht/~shulhan/pakakeh.go>::
List of open issues.

<https://lists.sr.ht/~shulhan/pakakeh.go>::
Link to submit the patches.


## License

Copyright (c) 2018 M. Shulhan &lt;ms@kilabit.info&gt;

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
   this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its
   contributors may be used to endorse or promote products derived from this
  software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED.
IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

----
That's it! Happy hacking!
