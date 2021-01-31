# `import "github/shuLhan/share"`
Shulhan <ms@kilabit.info>

[![PkgGoDev](https://pkg.go.dev/badge/github.com/shuLhan/share)](https://pkg.go.dev/github.com/shuLhan/share)

A collection of tools, public APIs, and libraries written and for working with
Go programming language.

This library is released every month, usually at the first week of month.

## Public APIs

* [**Telegram bot**](https://pkg.go.dev/github.com/shuLhan/share/api/telegram/bot):
  Package bot implement the Telegram Bot API https://core.telegram.org/bots/api.


## Command Line Interface

* [**gofmtcomment**](https://pkg.go.dev/github.com/shuLhan/share/cmd/gofmtcomment):
  Program to convert multi lines "/**/" comments into single line "//"
  format.

* [**smtpcli**](https://pkg.go.dev/github.com/shuLhan/share/cmd/smtpcli):
  Command line interface to SMTP client protocol.

* [**totp**](https://pkg.go.dev/github.com/shuLhan/share/cmd/totp):
  Program to generate Time-based One-time Password using secret key.

## Libraries

* [**ascii**](https://pkg.go.dev/github.com/shuLhan/share/lib/ascii): A
  library for working with ASCII characters.

* [**bytes**](https://pkg.go.dev/github.com/shuLhan/share/lib/bytes): A
  library for working with slice of bytes.

* [**clise**](https://pkg.go.dev/github.com/shuLhan/share/lib/clise): Package
  clise implements circular slice.

* [**contact**](https://pkg.go.dev/github.com/shuLhan/share/lib/contact): A
  library to import contact from Google, Microsoft, or Yahoo.

* [**crypto**](https://pkg.go.dev/github.com/shuLhan/share/lib/crypto):
  Package crypto provide a wrapper to simplify working with standard crypto
  package.

* [**debug**](https://pkg.go.dev/github.com/shuLhan/share/lib/debug): Package
  debug provide global debug variable, initialized through environment
  variable "DEBUG" or directly.

* [**dns**](https://pkg.go.dev/github.com/shuLhan/share/lib/dns): A library
  for working with Domain Name System (DNS) protocol.

* [**dsv**](https://pkg.go.dev/github.com/shuLhan/share/lib/dsv): A library
  for working with delimited separated value (DSV).

* [**email**](https://pkg.go.dev/github.com/shuLhan/share/lib/email): A
  library for working with Internet Message Format, as defined in RFC 5322.

 * [**dkim**](https://pkg.go.dev/github.com/shuLhan/share/lib/email/dkim):
   A library to parse and create DKIM-Signature header field value, as
   defined in RFC 6376.

 * [**maildir**](https://pkg.go.dev/github.com/shuLhan/share/lib/email/maildir):
   A library to manage email using maildir format.

* [**errors**](https://pkg.go.dev/github.com/shuLhan/share/lib/errors):
  Package errors provide an error type with Code, Message, and Name.

* [**floats64**](https://pkg.go.dev/github.com/shuLhan/share/lib/floats64): A
  library for working with slice of float64.

* [**git**](https://pkg.go.dev/github.com/shuLhan/share/lib/git): A wrapper
  for git command line interface.

* [**http**](https://pkg.go.dev/github.com/shuLhan/share/lib/http): Package
  http extends the standard http package with simplified routing handler and
  builtin memory file system.

* [**hunspell**](https://pkg.go.dev/github.com/shuLhan/share/lib/hunspell):
  A library to parse the Hunspell file format.

* [**ini**](https://pkg.go.dev/github.com/shuLhan/share/lib/ini): A library
  for reading and writing INI configuration as defined by Git configuration
  file syntax.

* [**ints**](https://pkg.go.dev/github.com/shuLhan/share/lib/ints): A library
  for working with slice of integer.

* [**ints64**](https://pkg.go.dev/github.com/shuLhan/share/lib/ints64): A
  library for working with slice of int64.

* [**io**](https://pkg.go.dev/github.com/shuLhan/share/lib/io): A library for
  simplify reading and watching files.

* [**json**](https://pkg.go.dev/github.com/shuLhan/share/lib/json): Package
  json extends the capabilities of standard json package.

* [**math**](https://pkg.go.dev/github.com/shuLhan/share/lib/math): Package
  math provide generic functions working with math.

 * [**big**](https://pkg.go.dev/github.com/shuLhan/share/lib/math/big):
  Package big extends the capabilities of standard "math/big" package by
  adding custom global precision to Float and Rat, global rounding mode, and
  custom bits precision to Float.

* [**memfs**](https://pkg.go.dev/github.com/shuLhan/share/lib/memfs): A
  library for mapping file system into memory and to generate an embedded Go
  file from it.

* [**mining**](https://pkg.go.dev/github.com/shuLhan/share/lib/mining): A
  mini library for data mining.

 * [**classifier/cart**](https://pkg.go.dev/github.com/shuLhan/share/lib/mining/classifier/cart):
  An implementation of the Classification and Regression Tree by
  Breiman, et al.

 * [**classififer/crf**](https://pkg.go.dev/github.com/shuLhan/share/lib/mining/classififer/crf):
  An implementation of the Cascaded Random Forest (CRF) algorithm, by
  Baumann, Florian, et al.

 * [**classifier/rf**](https://pkg.go.dev/github.com/shuLhan/share/lib/mining/classifier/rf):
  An implementation of ensemble of classifiers using random forest algorithm by
  Breiman and Cutler.

 * [**gain/gini**](https://pkg.go.dev/github.com/shuLhan/share/lib/gain/gini):
  A library to calculate Gini gain.

 * [**knn**](https://pkg.go.dev/github.com/shuLhan/share/lib/mining/knn):
  An implementation of the K Nearest Neighbor (KNN) using Euclidian to
  compute the distance between samples.

 * [**resampling/lnsmote**](https://pkg.go.dev/github.com/shuLhan/share/lib/mining/resampling/lnsmote):
  An implementation of the Local-Neighborhood algorithm from the paper of
  Maciejewski, Tomasz, and Jerzy Stefanowski.

 * [**resampling/smote**](https://pkg.go.dev/github.com/shuLhan/share/lib/mining/resampling/smote):
  An implementation of the Synthetic Minority Oversampling TEchnique
  (SMOTE).

 * [**tree/binary**](https://pkg.go.dev/github.com/shuLhan/share/lib/mining/tree/binary):
  An implementation of binary tree.

* [**net**](https://pkg.go.dev/github.com/shuLhan/share/lib/net): Constants
  and library for networking.

 * [**html**](https://pkg.go.dev/github.com/shuLhan/share/lib/net/html):
  Package html extends the golang.org/x/net/html by providing simplified
  methods for working with Node.

* [**numbers**](https://pkg.go.dev/github.com/shuLhan/share/lib/numbers): A
  library for working with integer, float, slice of integer, and slice of
  floats.

* [**os/exec**](https://pkg.go.dev/github.com/shuLhan/share/lib/os/exec):
  Package exec wrap the standar package "os/exec" to simplify calling Run
  with stdout and stderr.

* [**parser**](https://pkg.go.dev/github.com/shuLhan/share/lib/parser):
  Package parser provide a common text parser, using delimiters.

* [**paseto**](https://pkg.go.dev/github.com/shuLhan/share/lib/paseto): A
  simple, ready to use, implementation of Platform-Agnostic SEcurity TOkens
  (PASETO).

* [**reflect**](https://pkg.go.dev/github.com/shuLhan/share/lib/reflect):
  Package reflect extends the standard reflect package.

* [**runes**](https://pkg.go.dev/github.com/shuLhan/share/lib/runes): A
  library for working with slice of rune.

* [**sanitize**](https://pkg.go.dev/github.com/shuLhan/share/lib/sanitize): A
 library to sanitize markup document into plain text.

* [**smtp**](https://pkg.go.dev/github.com/shuLhan/share/lib/smtp): A library
 for building SMTP server or client. This package is working in progress.

* [**spf**](https://pkg.go.dev/github.com/shuLhan/share/lib/spf): Package spf
  implement Sender Policy Framework (SPF) per RFC 7208.

* [**sql**](https://pkg.go.dev/github.com/shuLhan/share/lib/sql): Package sql
  extends the standard library "database/sql.DB" that provide common
  functionality across DBMS.

* [**ssh**](https://pkg.go.dev/github.com/shuLhan/share/lib/ssh): Package ssh
  provide a wrapper for golang.org/x/crypto/ssh and a parser for SSH client
  configuration specification ssh_config(5).

* [**strings**](https://pkg.go.dev/github.com/shuLhan/share/lib/strings): A
  library for working with slice of string.

* [**tabula**](https://pkg.go.dev/github.com/shuLhan/share/lib/tabula): A
  library for working with rows, columns, or matrix (table), or in another
  terms working with data set.

* [**test**](https://pkg.go.dev/github.com/shuLhan/share/lib/test): A library
  for helping with testing.

* [**text**](https://pkg.go.dev/github.com/shuLhan/share/lib/text): A library
  for working with text.

 * [**text/diff**](https://pkg.go.dev/github.com/shuLhan/share/lib/text/diff):
  Package diff implement text comparison.

* [**time**](https://pkg.go.dev/github.com/shuLhan/share/lib/time): A library
  for working with time.

* [**totp**](https://pkg.go.dev/github.com/shuLhan/share/lib/totp): Package
  totp implement Time-Based One-Time Password Algorithm based on RFC 6238.

* [**websocket**](https://pkg.go.dev/github.com/shuLhan/share/lib/websocket):
  The WebSocket library for server and client. This websocket library has
  been tested with autobahn testsuite with 100% success rate.
  See [the status reports](https://github.com/shuLhan/share/blob/master/lib/websocket/AUTOBAHN.adoc).

* [**xmlrpc**](https://pkg.go.dev/github.com/shuLhan/share/lib/xmlrpc):
  Package xmlrpc provide an implementation of [XML-RPC specification](http://xmlrpc.com/spec.md).


## Changelog

Latest and full
[CHANGELOG](https://github.com/shuLhan/share/blob/master/CHANGELOG.adoc).


## Credits

* [Autobahn testsuite](https://github.com/crossbario/autobahn-testsuite)

That's it! Happy hacking!
