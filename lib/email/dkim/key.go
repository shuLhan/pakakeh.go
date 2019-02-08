// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dkim

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/shuLhan/share/lib/dns"
)

var dnsClientPool *dns.UDPClientPool // nolint: gochecknoglobals

//
// Key represent a DKIM key record.
//
type Key struct {
	// REQUIRED fields.

	// Public contains public key data.
	// ("p=", base64, REQUIRED)
	Public []byte

	// RECOMMENDED fields.

	// Version of DKIM key record.
	// ("v=", plain-text, RECOMMENDED, default is "DKIM1")
	Version []byte

	// OPTIONAL fields.

	// Type of key.
	// ("k=", plain-text, OPTIONAL, default is "rsa").
	Type KeyType

	// HashAlgs contains list of hash algorithm that might be used.
	// ("h=", plain-text colon-separated,  OPTIONAL, defaults to allowing
	// all algorithms)
	HashAlgs []HashAlg

	// Notes that might be interest to human.
	// ("n=", qp-section, OPTIONAL, default is empty)
	Notes []byte

	// Services contains list of service types to which this record
	// applies.
	// ("s=", plain-text colon-separated, OPTIONAL, default is "*").
	Services [][]byte

	// Flags contains list of flags.
	// ("t=", plain-text colon-separated, OPTIONAL, default is no flags set)
	Flags []KeyFlag

	// RSA contains parsed Public key.
	RSA *rsa.PublicKey

	// ExpiredAt define time when the key will be expired.
	// This is a local value derived from lookup time + RR TTL.
	ExpiredAt int64
}

//
// LookupKey DKIM (public) key using specific query method and DKIM domain
// name (selector plus SDID).
//
func LookupKey(qmethod QueryMethod, dname string) (key *Key, err error) {
	if len(dname) == 0 {
		return nil, nil
	}
	if qmethod.Type == QueryTypeDNS {
		key, err = lookupDNS(qmethod.Option, dname)
	}
	return key, err
}

//
// ParseTXT parse DNS TXT resource record into Key.
//
func ParseTXT(txt []byte, ttl uint32) (key *Key, err error) {
	p := newParser(txt)

	key = &Key{}
	for {
		tag, err := p.fetchTag()
		if err != nil {
			return nil, err
		}
		if tag == nil {
			break
		}
		err = key.set(tag)
		if err != nil {
			return nil, err
		}
	}

	key.ExpiredAt = time.Now().Unix() + int64(ttl)

	return key, nil
}

func lookupDNS(opt QueryOption, dname string) (key *Key, err error) {
	if opt == QueryOptionTXT {
		key, err = lookupDNSTXT(dname)
	}
	return key, err
}

func lookupDNSTXT(dname string) (key *Key, err error) {
	if dnsClientPool == nil {
		err = newDNSClientPool()
		if err != nil {
			return nil, err
		}
	}

	dnsClient := dnsClientPool.Get()

	dnsMsg, err := dnsClient.Lookup(dns.QueryTypeTXT, dns.QueryClassIN,
		[]byte(dname))
	if err != nil {
		dnsClientPool.Put(dnsClient)
		return nil, fmt.Errorf("dkim: LookupKey: " + err.Error())
	}
	if len(dnsMsg.Answer) == 0 {
		dnsClientPool.Put(dnsClient)
		return nil, fmt.Errorf("dkim: LookupKey: empty answer on '%s'", dname)
	}

	dnsClientPool.Put(dnsClient)

	answers := dnsMsg.FilterAnswers(dns.QueryTypeTXT)
	if len(answers) == 0 {
		return nil, fmt.Errorf("dkim: LookupKey: no TXT record on '%s'", dname)
	}
	if len(answers) != 1 {
		return nil, fmt.Errorf("dkim: LookupKey: multiple TXT records on '%s'", dname)
	}

	txt := answers[0].RData().([]byte)

	return ParseTXT(txt, answers[0].TTL)
}

func newDNSClientPool() (err error) {
	var ns []string

	if len(DefaultNameServers) > 0 {
		ns = DefaultNameServers
	} else {
		ns = dns.GetSystemNameServers("")
		if len(ns) == 0 {
			ns = append(ns, "1.1.1.1")
		}
	}

	dnsClientPool, err = dns.NewUDPClientPool(ns)

	return err
}

//
// Pack the key to be used in DNS TXT record.
//
func (key *Key) Pack() string {
	var bb strings.Builder

	if len(key.Version) > 0 {
		bb.WriteString("v=")
		bb.Write(key.Version)
		bb.WriteByte(';')
	}
	if len(key.Public) > 0 {
		if bb.Len() > 0 {
			bb.WriteByte(' ')
		}
		bb.WriteString("p=")
		bb.Write(key.Public)
		bb.WriteByte(';')
	}

	bb.WriteString(" k=")
	bb.Write(keyTypeNames[key.Type])
	bb.WriteByte(';')

	if len(key.HashAlgs) > 0 {
		bb.WriteString(" h=")
		bb.Write(packHashAlgs(key.HashAlgs))
		bb.WriteByte(';')
	}
	if len(key.Notes) > 0 {
		bb.WriteString(" n=")
		bb.Write(key.Notes)
		bb.WriteByte(';')
	}
	if len(key.Services) > 0 {
		bb.WriteString(" s=")
		bb.Write(bytes.Join(key.Services, sepColon))
		bb.WriteByte(';')
	}
	if len(key.Flags) > 0 {
		bb.WriteString(" t=")
		bb.Write(packKeyFlags(key.Flags))
	}

	return bb.String()
}

//
// IsExpired will return true if key ExpiredAt time is less than current time;
// otherwise it will return false.
//
func (key *Key) IsExpired() bool {
	return key.ExpiredAt < time.Now().Unix()
}

func (key *Key) set(t *tag) (err error) {
	if t == nil {
		return nil
	}
	switch t.key {
	case tagDNSPublicKey:
		pkey, err := base64.RawStdEncoding.DecodeString(string(t.value))
		if err != nil {
			err = fmt.Errorf("dkim: error decode public key: " + err.Error())
			return err
		}
		pk, err := x509.ParsePKIXPublicKey(pkey)
		if err != nil {
			err = fmt.Errorf("dkim: error parsing public key: " + err.Error())
			return err
		}

		key.RSA = pk.(*rsa.PublicKey)
		key.Public = t.value

	case tagDNSVersion:
		key.Version = t.value

	case tagDNSHashAlgs:
		key.HashAlgs = unpackHashAlgs(t.value)

	case tagDNSKeyType:
		key.Type = parseKeyType(t.value)

	case tagDNSNotes:
		key.Notes = t.value

	case tagDNSServices:
		key.Services = bytes.Split(t.value, sepColon)

	case tagDNSFlags:
		key.Flags = unpackKeyFlags(t.value)
	}
	return err
}
