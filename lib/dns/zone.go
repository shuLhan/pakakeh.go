// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/reflect"
)

// Zone represent a group of domain names shared a single root domain.
// A Zone contains at least one SOA record.
type Zone struct {
	// Records contains mapping between domain name and its resource
	// records.
	Records map[string][]*ResourceRecord `json:"-"`

	SOA   *RDataSOA
	rrSOA *ResourceRecord

	Path string `json:"-"`

	// The base domain of zone.
	// It must be absolute domain, end with period.
	Origin string

	messages []*Message
}

// NewZone create and initialize new zone.
func NewZone(file, origin string) (zone *Zone) {
	origin = strings.ToLower(toDomainAbsolute(origin))
	zone = &Zone{
		Records: make(map[string][]*ResourceRecord),
		SOA:     NewRDataSOA(origin, ``),
		Path:    file,
		Origin:  origin,
	}
	return zone
}

// LoadZoneDir load DNS record from zone formatted files in
// directory "dir".
// On success, it will return map of zone Origin and its content as list
// of Message.
// On fail, it will return possible partially parse zone file and an error.
func LoadZoneDir(dir string) (zoneFiles map[string]*Zone, err error) {
	if len(dir) == 0 {
		return nil, nil
	}

	var (
		d            *os.File
		zone         *Zone
		fi           os.FileInfo
		fis          []os.FileInfo
		name         string
		zoneFilePath string
	)

	d, err = os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("LoadZoneDir: %w", err)
	}

	fis, err = d.Readdir(0)
	if err != nil {
		err = d.Close()
		if err != nil {
			return nil, fmt.Errorf("LoadZoneDir: %w", err)
		}
		return nil, fmt.Errorf("LoadZoneDir: %w", err)
	}

	zoneFiles = make(map[string]*Zone)

	for _, fi = range fis {
		if fi.IsDir() {
			continue
		}

		// Ignore file that start with "." .
		name = fi.Name()
		if name[0] == '.' {
			continue
		}

		zoneFilePath = filepath.Join(dir, name)

		zone, err = ParseZoneFile(zoneFilePath, "", 0)
		if err != nil {
			return zoneFiles, fmt.Errorf("LoadZoneDir %q: %w", dir, err)
		}

		zoneFiles[zone.Origin] = zone
	}

	err = d.Close()
	if err != nil {
		return zoneFiles, fmt.Errorf(" LoadZoneDir %q: %w", dir, err)
	}

	return zoneFiles, nil
}

// ParseZone parse zone from raw bytes.
func ParseZone(content []byte, origin string, ttl uint32) (zone *Zone, err error) {
	var (
		logp = `ParseZone`
		zp   *zoneParser
	)

	if ttl <= 0 {
		ttl = DefaultSoaMinimumTtl
	}

	zone = NewZone(``, origin)
	zone.SOA.Minimum = ttl

	zp = newZoneParser(content, zone)

	err = zp.parse()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	zp.zone = nil

	return zone, nil
}

// ParseZoneFile parse zone file.
// The file name will be assumed as origin if parameter origin or $ORIGIN is
// not set.
func ParseZoneFile(file, origin string, ttl uint32) (zone *Zone, err error) {
	var (
		logp    = `ParseZoneFile`
		content []byte
	)

	if len(origin) == 0 {
		origin = filepath.Base(file)
	}

	content, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf(`%s: %q: %w`, logp, file, err)
	}

	zone, err = ParseZone(content, origin, ttl)
	if err != nil {
		return nil, fmt.Errorf(`%s: %q: %w`, logp, file, err)
	}

	zone.Path = file

	return zone, nil
}

// toDomainAbsolute add the period '.' to the end of domain d, if its not
// exist, to make it absolute domain.
func toDomainAbsolute(d string) string {
	var n = len(d)
	if n == 0 {
		d = `.`
	} else if d[n-1] != '.' {
		d += `.`
	}
	return d
}

// Add add new ResourceRecord to Zone.
func (zone *Zone) Add(rr *ResourceRecord) (err error) {
	var logp = `Add`
	err = zone.add(rr)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	zone.onUpdate()
	return nil
}

func (zone *Zone) add(rr *ResourceRecord) (err error) {
	var (
		msg *Message
		soa *RDataSOA
	)

	if rr.Type == RecordTypeSOA && rr.Name == zone.Origin {
		soa, _ = rr.Value.(*RDataSOA)
		if soa != nil {
			var cloneSoa = *soa
			zone.SOA = &cloneSoa
			zone.SOA.init()
		}
	} else {
		zone.recordAdd(rr)
	}

	for _, msg = range zone.messages {
		if msg.Question.Name != rr.Name {
			continue
		}
		if msg.Question.Type != rr.Type {
			continue
		}
		if msg.Question.Class != rr.Class {
			continue
		}
		return msg.AddAnswer(rr)
	}

	msg = &Message{
		Header: MessageHeader{
			IsAA:    true,
			QDCount: 1,
			ANCount: 1,
		},
		Question: MessageQuestion{
			Name:  rr.Name,
			Type:  rr.Type,
			Class: rr.Class,
		},
		Answer: []ResourceRecord{*rr},
	}
	zone.messages = append(zone.messages, msg)
	return nil
}

// Delete the zone file from storage.
func (zone *Zone) Delete() (err error) {
	return os.Remove(zone.Path)
}

// Messages return all pre-generated DNS messages.
func (zone *Zone) Messages() []*Message {
	return zone.messages
}

// Remove a ResourceRecord from zone file.
// If the RR is SOA it will reset the value back to default.
func (zone *Zone) Remove(rr *ResourceRecord) (err error) {
	var logp = `Remove`

	if rr.Type == RecordTypeSOA {
		zone.SOA = NewRDataSOA(zone.Origin, ``)
	} else {
		if zone.recordRemove(rr) {
			err = zone.Save()
			if err != nil {
				return fmt.Errorf(`%s: %w`, logp, err)
			}
		}
	}
	zone.onUpdate()
	return nil
}

// Save the content of zone records to file defined by Zone.Path.
// The zone content will be different with original file, since it does not
// preserve comment and indentation.
func (zone *Zone) Save() (err error) {
	var (
		logp = `Save`
		out  *os.File
	)

	out, err = os.OpenFile(zone.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf(`%s: %s: %w`, logp, zone.Path, err)
	}

	_, err = zone.WriteTo(out)
	if err != nil {
		err = fmt.Errorf(`%s: %s: %w`, logp, zone.Path, err)
	}

	var errc = out.Close()
	if errc != nil {
		if err == nil {
			return fmt.Errorf(`%s: %s: %w`, logp, zone.Path, errc)
		}
	}
	return err
}

func (zone *Zone) saveListRR(out io.Writer, dname string, listRR []*ResourceRecord) (total int, err error) {
	var (
		suffixOrigin = "." + zone.Origin

		hinfo *RDataHINFO
		minfo *RDataMINFO
		mx    *RDataMX
		rr    *ResourceRecord
		srv   *RDataSRV
		wks   *RDataWKS
		v     string
		n     int
		x     int
		ok    bool
	)

	for x, rr = range listRR {
		if x > 0 {
			dname = "\t"
		}
		switch rr.Type {
		case RecordTypeA, RecordTypeNULL, RecordTypeAAAA:
			n, err = fmt.Fprintf(out, "%s %d %s %s %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				RecordTypeNames[rr.Type], rr.Value.(string))

		case RecordTypeTXT:
			n, err = fmt.Fprintf(out, "%s %d %s %s %q\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				RecordTypeNames[rr.Type], rr.Value.(string))

		case RecordTypeNS, RecordTypeCNAME, RecordTypeMB,
			RecordTypeMG, RecordTypeMR:
			v, ok = rr.Value.(string)
			if !ok {
				err = errors.New("invalid record value for " +
					RecordTypeNames[rr.Type])
				break
			}

			if v == zone.Origin {
				v = "@"
			} else {
				v = strings.TrimSuffix(v, suffixOrigin)
			}
			n, err = fmt.Fprintf(out, "%s %d %s %s %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				RecordTypeNames[rr.Type], v)

		case RecordTypePTR:
			v, ok = rr.Value.(string)
			if !ok {
				err = errors.New("invalid record value for " +
					RecordTypeNames[rr.Type])
				break
			}
			v = strings.TrimSuffix(v, suffixOrigin)
			n, err = fmt.Fprintf(out, "%s %d IN PTR %s\n",
				rr.Name, rr.TTL, v)

		case RecordTypeWKS:
			wks, ok = rr.Value.(*RDataWKS)
			if !ok {
				err = errors.New("invalid record value for WKS")
				break
			}
			n, err = fmt.Fprintf(out,
				"%s %d %s WKS %s %d %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				wks.Address, wks.Protocol, wks.BitMap)

		case RecordTypeHINFO:
			hinfo, ok = rr.Value.(*RDataHINFO)
			if !ok {
				err = errors.New("invalid record value for HINFO")
				break
			}
			n, err = fmt.Fprintf(out,
				"%s %d %s HINFO %s %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				hinfo.CPU, hinfo.OS)

		case RecordTypeMINFO:
			minfo, ok = rr.Value.(*RDataMINFO)
			if !ok {
				err = errors.New("invalid record value for MINFO")
				break
			}
			n, err = fmt.Fprintf(out,
				"%s %d %s MINFO %s %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				strings.TrimSuffix(minfo.RMailBox, suffixOrigin),
				strings.TrimSuffix(minfo.EmailBox, suffixOrigin))

		case RecordTypeMX:
			mx, ok = rr.Value.(*RDataMX)
			if !ok {
				err = errors.New("invalid record value for MX")
				break
			}
			v = strings.TrimSuffix(mx.Exchange, suffixOrigin)
			n, err = fmt.Fprintf(out,
				"%s %d %s MX %d %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				mx.Preference, v)

		case RecordTypeSRV:
			srv, ok = rr.Value.(*RDataSRV)
			if !ok {
				err = errors.New("invalid record value for SRV")
				break
			}
			v = strings.TrimSuffix(srv.Target, suffixOrigin)
			n, err = fmt.Fprintf(out,
				"%s %d %s SRV %d %d %d %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				srv.Priority, srv.Weight, srv.Port, v)
		}
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// WriteTo write the zone as text into w.
// The result of WriteTo will be different with original content of zone file,
// since it does not preserve comment and indentation.
func (zone *Zone) WriteTo(out io.Writer) (total int64, err error) {
	var (
		logp = `Write`
		n    int
	)

	n, _ = fmt.Fprintf(out, "$ORIGIN %s\n", zone.Origin)
	total += int64(n)

	if len(zone.SOA.MName) > 0 {
		n, err = fmt.Fprintf(out,
			"@ SOA %s %s %d %d %d %d %d\n",
			zone.SOA.MName, zone.SOA.RName, zone.SOA.Serial, zone.SOA.Refresh,
			zone.SOA.Retry, zone.SOA.Expire, zone.SOA.Minimum)
		if err != nil {
			return total, fmt.Errorf(`%s: %w`, logp, err)
		}
		total += int64(n)
	}

	// Save the origin records first.
	var listRR = zone.Records[zone.Origin]
	if len(listRR) > 0 {
		n, err = zone.saveListRR(out, `@`, listRR)
		if err != nil {
			return total, fmt.Errorf(`%s: %w`, logp, err)
		}
		total += int64(n)
	}

	// Save the records ordered by name.
	var (
		names = make([]string, 0, len(zone.Records))

		dname string
	)
	for dname = range zone.Records {
		if dname == zone.Origin {
			continue
		}
		names = append(names, dname)
	}
	sort.Strings(names)

	for _, dname = range names {
		listRR = zone.Records[dname]
		dname = strings.TrimSuffix(dname, `.`+zone.Origin)
		n, err = zone.saveListRR(out, dname, listRR)
		if err != nil {
			return total, fmt.Errorf(`%s: %w`, logp, err)
		}
		total += int64(n)
	}
	return total, nil
}

// onUpdate handle when a record inserted, updated, or removed from zone.
// Basically, it set the SOA serial to current epoch or increase by one if
// the current serial and epoch are equal.
func (zone *Zone) onUpdate() {
	var serial = uint32(timeNow().Unix())
	if zone.SOA.Serial == serial {
		serial++
	} else if zone.SOA.Serial > serial {
		serial = zone.SOA.Serial + 1
	}
	zone.SOA.Serial = serial
}

// recordAdd a ResourceRecord into the zone.
func (zone *Zone) recordAdd(rr *ResourceRecord) {
	var (
		listRR = zone.Records[rr.Name]

		in *ResourceRecord
		x  int
	)

	// Replace the RR if its type is SOA because only one SOA
	// should exist per domain name.
	if rr.Type == RecordTypeSOA {
		for x, in = range listRR {
			if in.Type != RecordTypeSOA {
				continue
			}
			listRR[x] = rr
			return
		}
	}
	listRR = append(listRR, rr)
	zone.Records[rr.Name] = listRR
}

// recordRemove remove a ResourceRecord from list by its Name and Value.
// It will return true if the RR exist and removed.
func (zone *Zone) recordRemove(rr *ResourceRecord) bool {
	var (
		listRR = zone.Records[rr.Name]
		nlist  = len(listRR)

		in *ResourceRecord
		x  int
	)

	for x, in = range listRR {
		if in.Type != rr.Type {
			continue
		}
		if in.Class != rr.Class {
			continue
		}
		if !reflect.IsEqual(in.Value, rr.Value) {
			continue
		}
		copy(listRR[x:], listRR[x+1:])
		listRR[nlist-1] = nil
		zone.Records[rr.Name] = listRR[:nlist-1]
		return true
	}
	return false
}

func (zone *Zone) soaRecord() (rrsoa *ResourceRecord) {
	if zone.rrSOA == nil {
		zone.rrSOA = &ResourceRecord{
			Value: zone.SOA,
			Name:  zone.Origin,
			Type:  RecordTypeSOA,
			Class: RecordClassIN,
			TTL:   zone.SOA.Minimum,
		}
	}
	return zone.rrSOA
}
