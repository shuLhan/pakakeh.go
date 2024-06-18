# `import "git.sr.ht/~shulhan/pakakeh.go"`

[Go documentation](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go).

"pakakeh.go" is a collection of tools, public HTTP APIs, and libraries
written and for working with Go programming language.

This library is released every month, usually at the first week of month.

## Public APIs

[**Telegram bot**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/api/telegram/bot)::
Package bot implement the
[Telegram Bot API](https://core.telegram.org/bots/api).


## Command Line Interface

[**bcrypt**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/bcrypt)::
CLI to compare or generate hash using bcrypt.

[**epoch**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/epoch)::
Program epoch print the current date and time (Unix seconds, milliseconds,
nanoseconds, local time, and UTC time) or the date and time based on the
epoch on first parameter.

[**gofmtcomment**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/gofmtcomment)::
Program to convert multi lines "/**/" comments into single line "//" format.

[**ini**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/ini)::
Program ini provide a command line interface to get and set values in the
[INI file format](https://godocs.io/git.sr.ht/~shulhan/pakakeh.go/lib/ini).

[**smtpcli**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/smtpcli)::
Command line interface SMTP client protocol.
This is an example of implementation Client from
[lib/smtp](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/smtp).

[**xtrk**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/xtrk)::
Program xtrk is command line interface to uncompress and/or unarchive a
file.
Supported format: bzip2, gzip, tar, zip, tar.bz2, tar.gz.

[**totp**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/cmd/totp)::
Program to generate Time-based One-time Password using secret key.
This is just an example of implementation of
[lib/totp](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/totp).
See
<https://sr.ht/~shulhan/gotp/> for a complete implementation that support
encryption.

## Libraries

[**ascii**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ascii)::
A library for working with ASCII characters.

[**bytes**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/bytes)::
A library for working with slice of bytes.

[**clise**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/clise)::
Package clise implements circular slice.

[**contact**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/contact)::
A library to import contact from Google, Microsoft, or Yahoo.

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

[**floats64**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/floats64)::
A library for working with slice of float64.

[**git**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/git)::
A wrapper for git command line interface.

[**http**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/http)::
Package http extends the standard http package with simplified routing handler
and builtin memory file system.

[**html**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/html)::
Package html extends the golang.org/x/net/html by providing simplified
methods for working with Node.

[**hunspell**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/hunspell)::
[WORK IN PROGRESS].
A library to parse the Hunspell file format.

[**ini**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ini)::
A library for reading and writing INI configuration as defined by Git
configuration file syntax.

[**ints**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ints)::
A library for working with slice of integer.

[**ints64**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/ints64)::
A library for working with slice of int64.

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

[**reflect**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/reflect)::
Package reflect extends the standard reflect package.

[**runes**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/runes)::
A library for working with slice of rune.

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

[**websocket**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/websocket)::
The WebSocket library for server and client. This WebSocket library has
been tested with autobahn testsuite with 100% success rate.
[the status report](https://git.sr.ht/~shulhan/pakakeh.go/blob/main/lib/websocket/AUTOBAHN.adoc).

[**xmlrpc**](https://pkg.go.dev/git.sr.ht/~shulhan/pakakeh.go/lib/xmlrpc)::
Package xmlrpc provide an implementation of
[XML-RPC specification](http://xmlrpc.com/spec.md).


## Changelog

Latest
[CHANGELOG](https://git.sr.ht/~shulhan/pakakeh.go/blob/main/CHANGELOG.adoc).


## Credits

[Autobahn testsuite](https://github.com/crossbario/autobahn-testsuite) for
testing WebSocket library.

That's it! Happy hacking!
