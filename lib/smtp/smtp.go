// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package smtp

import (
	"bytes"
	"errors"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

// List of SMTP status codes.
const (
	//
	// 2yz  Positive Completion reply
	//
	// The requested action has been successfully completed.  A new
	// request may be initiated.
	//
	StatusSystem        = 211
	StatusHelp          = 214
	StatusReady         = 220
	StatusClosing       = 221
	StatusAuthenticated = 235 // RFC 4954
	StatusOK            = 250
	StatusAddressChange = 251 // RFC 5321, section 3.4.
	StatusVerifyFailed  = 252 // RFC 5321, section 3.5.3.

	//
	// 3xx Positive Intermediate reply.
	//
	// The command has been accepted, but the requested action is being
	// held in abeyance, pending receipt of further information.  The
	// SMTP client should send another command specifying this
	// information.  This reply is used in command DATA.
	//
	StatusAuthReady = 334
	StatusDataReady = 354

	//
	// 4xx Transient Negative Completion reply
	//
	// The command was not accepted, and the requested action did not
	// occur.  However, the error condition is temporary, and the action
	// may be requested again.  The sender should return to the beginning
	// of the command sequence (if any).  It is difficult to assign a
	// meaning to "transient" when two different sites (receiver- and
	// sender-SMTP agents) must agree on the interpretation.  Each reply
	// in this category might have a different time value, but the SMTP
	// client SHOULD try again.  A rule of thumb to determine whether a
	// reply fits into the 4yz or the 5yz category (see below) is that
	// replies are 4yz if they can be successful if repeated without any
	// change in command form or in properties of the sender or receiver
	// (that is, the command is repeated identically and the receiver
	// does not put up a new implementation).
	//
	StatusShuttingDown             = 421
	StatusPasswordTransitionNeeded = 432 // RFC 4954 section 4.7.12.
	StatusMailboxUnavailable       = 450
	StatusLocalError               = 451
	StatusNoStorage                = 452
	StatusTemporaryAuthFailure     = 454 // RFC 4954 section 4.7.0.
	StatusParameterUnprocessable   = 455

	//
	// 5xx indicate permanent failure.
	//
	// The command was not accepted and the requested action did not
	// occur.  The SMTP client SHOULD NOT repeat the exact request (in
	// the same sequence).  Even some "permanent" error conditions can be
	// corrected, so the human user may want to direct the SMTP client to
	// reinitiate the command sequence by direct action at some point in
	// the future (e.g., after the spelling has been changed, or the user
	// has altered the account status).
	//
	StatusCmdUnknown           = 500 // RFC 5321 section 4.2.4.
	StatusCmdTooLong           = 500 // RFC 5321 section 4.3.2, RFC 4954 section 5.5.6.
	StatusCmdSyntaxError       = 501
	StatusCmdNotImplemented    = 502 // RFC 5321 section 4.2.4.
	StatusCmdBadSequence       = 503
	StatusParamUnimplemented   = 504
	StatusNotAuthenticated     = 530
	StatusAuthMechanismTooWeak = 534 // RFC 4954 section 5.7.9.
	StatusInvalidCredential    = 535 // RFC 4954 section 5.7.8.
	StatusMailboxNotFound      = 550
	StatusAddressChangeAborted = 551 // RFC 5321 section 3.4.
	StatusMailNoStorage        = 552
	StatusMailboxIncorrect     = 553
	StatusTransactionFailed    = 554
	StatusMailRcptParamUnknown = 555
)

// ParsePath parse the Reverse-path or Forward-path as in argument of MAIL and
// RCPT commands.
// This function ignore the source route and only return the mailbox.
// Empty mailbox without an error is equal to Null Reverse-Path "<>".
func ParsePath(path []byte) (mailbox []byte, err error) {
	if len(path) == 0 {
		return nil, errors.New("ParsePath: empty path")
	}
	if path[0] != '<' {
		return nil, errors.New("ParsePath: missing opening '<'")
	}
	if path[len(path)-1] != '>' {
		return nil, errors.New("ParsePath: missing closing '>'")
	}
	if len(path) == 2 {
		return nil, nil
	}

	// Skip the source-route
	x := 1
	if path[x] == '@' {
		for ; x < len(path); x++ {
			if path[x] == ':' {
				x++
				break
			}
		}
	}

	mailbox = ParseMailbox(path[x : len(path)-1])
	if mailbox == nil {
		return nil, errors.New("ParsePath: invalid mailbox format")
	}

	return mailbox, nil
}

// ParseMailbox parse the mailbox, remove comment or any escaped characters
// insided quoted-string.
func ParseMailbox(data []byte) (mailbox []byte) {
	if len(data) == 0 {
		return nil
	}

	allowOnLocal := []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-',
		'/', '=', '?', '^', '_', '`', '{', '|', '}', '~'}

	at := bytes.LastIndex(data, []byte{'@'})
	if at < 0 {
		return parseLocalDomain(data, allowOnLocal)
	}

	local := data[:at]
	domain := data[at+1:]

	if local[0] == '"' {
		if len(local) == 1 {
			return nil
		}
		if local[len(local)-1] == '"' {
			local = parseQuotedMailbox(local[1 : len(local)-1])
		}
	} else {
		local = parseLocalDomain(local, allowOnLocal)
	}
	if local == nil {
		return nil
	}

	domain = parseDomainAddress(domain)
	if domain == nil {
		return nil
	}

	mailbox = append(mailbox, local...)
	mailbox = append(mailbox, '@')
	mailbox = append(mailbox, domain...)

	return mailbox
}

// parseLocalDomain parse local-part or domain-name of mailbox.
// Rules,
// * dot is not allowed at beginning or end.
// * local part or domain can contains comment "(any)"
// * dot must not appear consecutively
func parseLocalDomain(data []byte, allow []byte) (out []byte) {
	if data[0] == '.' || data[len(data)-1] == '.' {
		return nil
	}
	var (
		found bool
		isDot bool
	)
	var x int
	for ; x < len(data); x++ {
		if data[x] == '(' {
			x = skipComment(data, x)
			if x == len(data) {
				return nil
			}
			continue
		}
		if isDot {
			if data[x] == '.' {
				return nil
			}
			out = append(out, '.')
			isDot = false
		}
		if ascii.IsAlnum(data[x]) {
			out = append(out, data[x])
			continue
		}
		found = false
		for _, c := range allow {
			if c == data[x] {
				out = append(out, data[x])
				found = true
				break
			}
		}
		if found {
			continue
		}
		if data[x] == '.' {
			isDot = true
			continue
		}
		return nil
	}
	return out
}

func parseDomainAddress(data []byte) (out []byte) {
	if data[0] == '[' && data[len(data)-1] == ']' {
		return data
	}
	allowOnDomain := []byte{'-', '_'}
	return parseLocalDomain(data, allowOnDomain)
}

func skipComment(data []byte, x int) int {
	for ; x < len(data); x++ {
		if data[x] == ')' {
			return x
		}
	}
	return x
}

// parseQuotedMailbox parse the mailbox in quoted format.
//
// The following ASCII characters are accepted: %d32-33 / %d35-91 / %d93-126.
// Character code %d34 is '"'.
// Quoted-pair character is character %d92 ("\" or backslash) followed by
// %d32-126.
func parseQuotedMailbox(data []byte) (out []byte) {
	out = append(out, '"')
	var x int
	for ; x < len(data); x++ {
		if data[x] < 32 || data[x] == 34 || data[x] > 126 {
			return nil
		}
		if data[x] == 92 {
			x++
			if x == len(data) {
				return nil
			}
			if data[x] < 32 || data[x] > 126 {
				return nil
			}
			out = append(out, data[x])
			continue
		}
		out = append(out, data[x])
	}
	out = append(out, '"')
	return out
}
