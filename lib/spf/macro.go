// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package spf

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
	libdns "git.sr.ht/~shulhan/pakakeh.go/lib/dns"
	libnet "git.sr.ht/~shulhan/pakakeh.go/lib/net"
)

const (
	macroSender       byte = 's'
	macroSenderLocal  byte = 'l'
	macroSenderDomain byte = 'o'
	macroDomain       byte = 'd'
	macroIP           byte = 'i'
	macroPtr          byte = 'p'
	macroInAddr       byte = 'v'
	macroEhloDomain   byte = 'h'

	// The following macro letters are allowed only in "exp".
	macroExpClientIP byte = 'c'
	macroExpDomain   byte = 'r'
	macroExpCurTime  byte = 't'
)

type macro struct {
	ref *Result

	// mode is either mechanism "include", modifier "redirect", or
	// modifier "exp".
	mode string

	out  []byte
	dels []byte

	nright int

	letter byte

	isReversed bool
}

func macroExpand(ref *Result, mode string, data []byte) (out []byte, err error) {
	m := &macro{
		ref:  ref,
		mode: mode,
	}

	err = m.parse(data)
	if err != nil {
		return nil, err
	}

	return m.out, nil
}

// isDelimiter will return true if value of character c is one of delimiter
// in,
//
//	delimiter = "." / "-" / "+" / "," / "/" / "_" / "="
func isDelimiter(c byte) bool {
	if c == '.' || c == '-' || c == '+' || c == ',' || c == '/' || c == '_' || c == '=' {
		return true
	}
	return false
}

// isMacroLetter will return true if value of character c is one of the valid
// letter for macro:
//
//	macro-letter     = "s" / "l" / "o" / "d" / "i" / "p" / "h" / "v" /
//	                   "c" / "r" / "t"
//
// Letter "c", "r", and "t" only valid if modifier is "exp".
func isMacroLetter(mode string, c byte) bool {
	if c == macroSender || c == macroSenderLocal ||
		c == macroSenderDomain || c == macroDomain || c == macroIP ||
		c == macroPtr || c == macroEhloDomain || c == macroInAddr {
		return true
	}
	if mode == modifierExp {
		if c == macroExpClientIP || c == macroExpDomain || c == macroExpCurTime {
			return true
		}
	}
	return false
}

func (m *macro) reset() {
	m.letter = 0
	m.nright = 0
	m.dels = m.dels[:0]
	m.isReversed = false
}

func (m *macro) parse(data []byte) (err error) {
	var (
		state byte
	)

	for x := 0; x < len(data); x++ {
		switch state {
		case 0:
			switch data[x] {
			case '%':
				state = 1
			default:
				if data[x] < 0x21 || data[x] > 0x7E {
					return fmt.Errorf("invalid macro literal '%c' at position %d", data[x], x)
				}
				m.out = append(m.out, data[x])
			}
		case 1:
			switch data[x] {
			case '{':
				state = 2
			case '%':
				m.out = append(m.out, '%')
				state = 0
			case '_':
				m.out = append(m.out, ' ')
				state = 0
			case '-':
				m.out = append(m.out, []byte("%20")...)
				state = 0
			default:
				return fmt.Errorf("syntax error '%c' at position %d", data[x], x)
			}

		// macro-letter
		case 2:
			if !isMacroLetter(m.mode, data[x]) {
				return fmt.Errorf("unknown macro letter '%c' at position %d", data[x], x)
			}
			m.letter = data[x]
			state = 3

		// *DIGIT [ "r" ] *delimiter
		case 3:
			if ascii.IsDigit(data[x]) {
				digits := make([]byte, 0, 3)

				digits = append(digits, data[x])
				x++
				for ; x < len(data); x++ {
					if !ascii.IsDigit(data[x]) {
						break
					}
					digits = append(digits, data[x])
				}

				m.nright, err = strconv.Atoi(string(digits))
				if err != nil {
					return fmt.Errorf("failed to convert digits %q: %s", digits, err)
				}

				if x == len(data) {
					return fmt.Errorf("missing closing '}'")
				}
			}
			if data[x] == 'r' {
				m.isReversed = true
				x++
				if x == len(data) {
					return fmt.Errorf("missing closing '}'")
				}
			}
			if data[x] != '}' {
				for ; x < len(data); x++ {
					if isDelimiter(data[x]) {
						m.dels = append(m.dels, data[x])
						continue
					}
					break
				}
				if x == len(data) {
					return fmt.Errorf("missing closing '}'")
				}
				if data[x] != '}' {
					return fmt.Errorf("missing closing '}', got '%c'", data[x])
				}
			}

			m.expand()
			state = 0
		}
	}

	return nil
}

func (m *macro) expand() {
	var (
		values [][]byte
	)

	value := m.expandLetter()

	if len(m.dels) > 0 {
		values = bytes.Split(value, m.dels)
	} else {
		values = bytes.Split(value, []byte{'.'})
	}

	if m.isReversed {
		values = reverseValues(values)
	}

	if m.nright > 0 {
		values = chopRight(values, m.nright)
	}

	value = bytes.Join(values, []byte{'.'})

	m.out = append(m.out, value...)

	m.reset()
}

func (m *macro) expandLetter() (value []byte) {
	switch m.letter {
	case macroSender:
		value = m.ref.Sender
	case macroSenderLocal:
		value = m.ref.senderLocal
	case macroSenderDomain:
		value = m.ref.senderDomain
	case macroDomain:
		value = m.ref.Domain
	case macroIP:
		value = toDotIP(m.ref.IP)
	case macroPtr:
		ptrDomain, err := libdns.LookupPTR(dnsClient, m.ref.IP)

		// If there are no validated domain names or if a DNS error
		// occurs, the string "unknown" is used.
		// RFC 7208 Section 7.3.
		if err != nil || len(ptrDomain) == 0 {
			ptrDomain = "unknown"
		}
		value = []byte(ptrDomain)

	case macroInAddr:
		if libnet.IsIPv4(m.ref.IP) {
			value = []byte("in-addr")
		}
		if libnet.IsIPv6(m.ref.IP) {
			value = []byte("ip6")
		}
	case macroEhloDomain:
		value = m.ref.Domain
	case macroExpClientIP:
		value = toDotIP(m.ref.IP)
	case macroExpDomain:
		value = m.ref.Hostname
	case macroExpCurTime:
		now := time.Now()
		strNow := fmt.Sprintf("%d", now.Unix())
		value = []byte(strNow)
	}
	return value
}

// reverseValues reverse the items in slices.  For example, "[a b]" would
// become "[b a]".
func reverseValues(in [][]byte) (out [][]byte) {
	out = make([][]byte, 0, len(in))
	for x := len(in) - 1; x >= 0; x-- {
		out = append(out, in[x])
	}
	return
}

// chopRight take n items from input slice.  If input length less than n,
// return all of them.
func chopRight(in [][]byte, n int) (out [][]byte) {
	if len(in) < n {
		return in
	}

	out = make([][]byte, 0, n)

	for x := len(in) - n; x < len(in); x++ {
		out = append(out, in[x])
	}
	return
}

// toDotIP convert the IP address into dotted format.  For IPv4, it will
// return the usual IP address, while for IPv6 it will split each hexa numbers
// into dot.
func toDotIP(ip net.IP) []byte {
	if libnet.IsIPv4(ip) {
		return []byte(ip.String())
	}
	if libnet.IsIPv6(ip) {
		return libnet.ToDotIPv6(ip)
	}
	return []byte(ip.String())
}
