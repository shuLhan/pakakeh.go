// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/shuLhan/share/lib/ascii"
	libbytes "github.com/shuLhan/share/lib/bytes"
)

const (
	envLocaldomain = "LOCALDOMAIN"
)

var (
	// lambda to test os.Hostname.
	getHostname = os.Hostname
)

// ResolvConf contains value of resolver configuration file.
//
// Reference: "man resolv.conf" in Linux.
type ResolvConf struct {
	// OptMisc contains other options with string key and boolean value.
	OptMisc map[string]bool

	// Local domain name.
	// Most queries for names within this domain can use short names
	// relative to the local domain.  If set to '.', the root domain
	// is considered.  If no domain entry is present, the domain is
	// determined from the local hostname returned by gethostname(2);
	// the domain part is taken to be everything after the first '.'.
	// Finally, if the hostname does not contain a domain part, the
	// root domain is assumed.
	Domain string

	// Search list for host-name lookup.
	// The search list is normally determined from the local domain
	// name; by default, it contains only the local domain name.
	// This may be changed by listing the desired domain search path
	// following the search keyword with spaces or tabs separating
	// the names.  Resolver queries having fewer than ndots dots
	// (default is 1) in them will be attempted using each component
	// of the search path in turn until a match is found.  For
	// environments with multiple subdomains please read options
	// ndots:n below to avoid man-in-the-middle attacks and
	// unnecessary traffic for the root-dns-servers.  Note that this
	// process may be slow and will generate a lot of network traffic
	// if the servers for the listed domains are not local, and that
	// queries will time out if no server is available for one of the
	// domains.
	//
	// The search list is currently limited to six domains with a
	// total of 256 characters.
	Search []string

	// Name server IP address
	// Internet address of a name server that the resolver should
	// query, either an IPv4 address (in dot notation), or an IPv6
	// address in colon (and possibly dot) notation as per RFC 2373.
	// Up to 3 name servers may be listed, one per keyword.  If there are
	// multiple servers, the resolver library queries them in the order
	// listed.  If no nameserver entries are present, the default is to
	// use the name server on the local machine.  (The algorithm used is
	// to try a name server, and if the query times out, try the next,
	// until out of name servers, then repeat trying all the name servers
	// until a maximum number of retries are made.)
	NameServers []string

	// Sets a threshold for the number of dots which must appear in a name
	// before an initial absolute query will be made.  The default for n
	// is 1, meaning that if there are any dots in a name, the name will
	// be tried first as an absolute name before any search list elements
	// are appended to it.  The value for this option is silently capped
	// to 15.
	NDots int

	// Sets the amount of time the resolver will wait for a response from
	// a remote name server before retrying the query via a different name
	// server. This may not be the total time taken by any resolver API
	// call and there is no guarantee that a single resolver API call maps
	// to a single timeout.  Measured in seconds, the default is 5 The
	// value for this option is silently capped to 30.
	Timeout int

	// Sets the number of times the resolver will send a query to its name
	// servers before giving up and returning an error to the calling
	// application.  The default is 2. The value for this option is
	// silently capped to 5.
	Attempts int
}

// NewResolvConf open resolv.conf file in path and return the parsed records.
func NewResolvConf(path string) (rc *ResolvConf, err error) {
	var (
		logp = `NewResolvConf`

		content []byte
	)

	content, err = os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var parser = libbytes.NewParser(content, nil)

	rc = &ResolvConf{
		OptMisc: make(map[string]bool),
	}
	rc.parse(parser)

	return rc, nil
}

// Init parse resolv.conf from string.
func (rc *ResolvConf) Init(src string) {
	var parser = libbytes.NewParser([]byte(src), nil)

	rc.reset()
	rc.parse(parser)
}

func (rc *ResolvConf) reset() {
	rc.Domain = ``
	rc.Search = nil
	rc.NameServers = nil
	rc.OptMisc = make(map[string]bool)
}

// parse open and parse the resolv.conf file.
//
// Lines that contain a semicolon (;) or hash character (#) in the first
// column are treated as comments.
//
// The keyword and value must appear on a single line, and the keyword (e.g.,
// nameserver) must start the line.  The value follows the keyword, separated
// by white space.
//
// See `man resolv.conf`
func (rc *ResolvConf) parse(parser *libbytes.Parser) {
	parser.SetDelimiters([]byte{' ', '\t', ';', '#', '\n'})

	var (
		stok string
		tok  []byte
		c    byte
	)

	for {
		tok, c = parser.ReadNoSpace()
		if len(tok) == 0 {
			if c == 0 {
				break
			}
			if c == ';' || c == '#' {
				// Skip empty line or keyword without value.
				parser.SkipLine()
			}
			continue
		}

		_, c = parser.SkipHorizontalSpaces()
		if c == '\n' || c == ';' || c == '#' {
			// Skip empty line or keyword without value.
			parser.SkipLine()
			continue
		}

		tok = ascii.ToLower(tok)
		stok = string(tok)

		switch stok {
		case `domain`:
			tok, c = parser.ReadNoSpace()
			if len(tok) != 0 {
				rc.Domain = string(tok)
			}
			if c != '\n' {
				parser.SkipLine()
			}

		case `search`:
			rc.parseSearch(parser)

		case `nameserver`:
			tok, c = parser.ReadNoSpace()
			if len(tok) != 0 {
				if len(rc.NameServers) < 3 && len(tok) > 0 {
					rc.NameServers = append(rc.NameServers, string(tok))
				}
			}
			if c != '\n' {
				parser.SkipLine()
			}

		case `options`:
			rc.parseOptions(parser)

		default:
			parser.SkipLine()
		}
	}

	rc.sanitize()
}

// parseSearch parse the "search" directive using the following format,
//
//	search domain *(domain)
//
// The domain and search keywords are mutually exclusive.
// If more than one instance of these keywords is present, the last instance
// wins.
func (rc *ResolvConf) parseSearch(parser *libbytes.Parser) {
	var (
		max    = 6
		maxLen = 255

		curLen int
		tok    []byte
		c      byte
	)

	rc.Search = nil

	for {
		tok, c = parser.ReadNoSpace()

		if curLen+len(tok) > maxLen {
			break
		}

		rc.Search = append(rc.Search, string(tok))
		if len(rc.Search) == max {
			break
		}

		curLen += len(tok)

		if c == '\n' {
			break
		}
	}
	if c != '\n' {
		parser.SkipLine()
	}
}

func (rc *ResolvConf) parseOptions(parser *libbytes.Parser) {
	var (
		tok []byte
		c   byte
	)

	for {
		tok, c = parser.ReadNoSpace()
		if len(tok) == 0 {
			break
		}
		rc.parseOptionsKV(tok)
		if c == '\n' {
			break
		}
	}
	if c != '\n' {
		parser.SkipLine()
	}
}

func (rc *ResolvConf) parseOptionsKV(opt []byte) {
	var (
		kv  = bytes.SplitN(opt, []byte{':'}, 2)
		key = string(kv[0])

		value string
	)
	if len(kv) == 2 {
		value = string(kv[1])
	}

	switch key {
	case `ndots`:
		rc.NDots, _ = strconv.Atoi(value)
	case `timeout`:
		rc.Timeout, _ = strconv.Atoi(value)
	case `attempts`:
		rc.Attempts, _ = strconv.Atoi(value)
	default:
		rc.OptMisc[key] = true
	}
}

func (rc *ResolvConf) sanitize() {
	// Sanitize domain name
	if len(rc.Domain) == 0 {
		rc.Domain, _ = getHostname()
	}
	if len(rc.Domain) > 0 {
		names := strings.Split(rc.Domain, ".")
		if len(names) > 1 {
			rc.Domain = strings.Join(names[1:], ".")
		}
	}

	// The search keyword of a system's resolv.conf file can be overridden
	// on a per-process basis by setting the environment variable
	// LOCALDOMAIN to a space-separated list of search domains.
	envLocalDomain := os.Getenv(envLocaldomain)
	if len(envLocalDomain) > 0 {
		rc.Search = strings.Split(envLocalDomain, " ")
		if len(rc.Search) > 6 {
			rc.Search = rc.Search[:6]
		}
	}

	if rc.NDots == 0 {
		rc.NDots = 1
	} else if rc.NDots > 15 {
		rc.NDots = 15
	}
	if rc.Timeout == 0 {
		rc.Timeout = 5
	} else if rc.Timeout > 30 {
		rc.Timeout = 30
	}
	if rc.Attempts == 0 {
		rc.Attempts = 2
	} else if rc.Attempts > 5 {
		rc.Attempts = 5
	}
}

// PopulateQuery given a domain name to be resolved, generate list of names
// to be queried based on registered Search in the resolv.conf.
// The dname itself will be on top of the list.
// If the number of dots in dname less than NDots then each Search domain will
// be appended as suffix and added to the list.
func (rc *ResolvConf) PopulateQuery(dname string) (queries []string) {
	var (
		s     string
		ndots int
		r     rune
	)

	for _, r = range dname {
		if r == '.' {
			ndots++
			continue
		}
	}

	queries = append(queries, dname)
	if ndots >= rc.NDots {
		return queries
	}
	for _, s = range rc.Search {
		queries = append(queries, dname+"."+s)
	}
	return queries
}

// WriteTo write the ResolvConf into w.
func (rc *ResolvConf) WriteTo(w io.Writer) (n int64, err error) {
	var bb bytes.Buffer

	if len(rc.Domain) > 0 {
		fmt.Fprintf(&bb, "domain %s\n", rc.Domain)
	}

	var k string

	if len(rc.Search) > 0 {
		fmt.Fprint(&bb, `search`)
		for _, k = range rc.Search {
			bb.WriteString(` ` + k)
		}
		bb.WriteByte('\n')
	}
	for _, k = range rc.NameServers {
		fmt.Fprintf(&bb, "nameserver %s\n", k)
	}

	if rc.NDots > 0 {
		fmt.Fprintf(&bb, "options ndots:%d\n", rc.NDots)
	}
	if rc.Timeout > 0 {
		fmt.Fprintf(&bb, "options timeout:%d\n", rc.Timeout)
	}
	if rc.Attempts > 0 {
		fmt.Fprintf(&bb, "options attempts:%d\n", rc.Attempts)
	}

	if len(rc.OptMisc) > 0 {
		var keys []string
		for k = range rc.OptMisc {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k = range keys {
			fmt.Fprintf(&bb, "options %s\n", k)
		}
	}

	var nwrite int

	nwrite, err = w.Write(bb.Bytes())
	if err != nil {
		return int64(nwrite), fmt.Errorf(`WriteTo: %w`, err)
	}

	return int64(nwrite), nil
}
