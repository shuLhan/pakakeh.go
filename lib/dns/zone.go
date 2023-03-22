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

	libio "github.com/shuLhan/share/lib/io"
)

// Zone represent a group of domain names shared a single root domain.
// A Zone contains at least one SOA record.
type Zone struct {
	Records  ZoneRecords `json:"-"`
	Path     string      `json:"-"`
	Name     string
	messages []*Message
	SOA      RDataSOA
}

// NewZone create and initialize new zone.
func NewZone(file, name string) *Zone {
	return &Zone{
		Path: file,
		Name: name,
		SOA: RDataSOA{
			MName: name,
		},
		Records: make(ZoneRecords),
	}
}

// LoadZoneDir load DNS record from zone formatted files in
// directory "dir".
// On success, it will return map of file name and Zone content as list
// of Message.
// On fail, it will return possible partially parse zone file and an error.
func LoadZoneDir(dir string) (zoneFiles map[string]*Zone, err error) {
	if len(dir) == 0 {
		return nil, nil
	}

	var (
		d            *os.File
		zoneFile     *Zone
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

		zoneFile, err = ParseZoneFile(zoneFilePath, "", 0)
		if err != nil {
			return zoneFiles, fmt.Errorf("LoadZoneDir %q: %w", dir, err)
		}

		zoneFiles[name] = zoneFile
	}

	err = d.Close()
	if err != nil {
		return zoneFiles, fmt.Errorf(" LoadZoneDir %q: %w", dir, err)
	}

	return zoneFiles, nil
}

// ParseZoneFile parse zone file.
// The file name will be assumed as origin if parameter origin or $ORIGIN is
// not set.
func ParseZoneFile(file, origin string, ttl uint32) (zone *Zone, err error) {
	var (
		logp = `ParseZoneFile`
		m    = newZoneParser(nil)
	)

	m.ttl = ttl

	if len(origin) > 0 {
		m.origin = origin
	} else {
		m.origin = filepath.Base(file)
	}

	m.origin = strings.ToLower(m.origin)
	if m.origin[len(m.origin)-1] != '.' {
		m.origin += `.`
	}

	m.reader, err = libio.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf(`%s: %q: %w`, logp, file, err)
	}

	err = m.parse()
	if err != nil {
		return nil, fmt.Errorf(`%s: %q: %w`, logp, file, err)
	}

	m.zone.Name = m.origin

	zone = m.zone
	m.zone = nil
	return zone, nil
}

// Add add new ResourceRecord to Zone.
func (zone *Zone) Add(rr *ResourceRecord) (err error) {
	var (
		msg *Message
		soa *RDataSOA
	)

	if rr.Type == RecordTypeSOA {
		soa, _ = rr.Value.(*RDataSOA)
		if soa != nil {
			zone.SOA = *soa
		}
	} else {
		zone.Records.add(rr)
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
func (zone *Zone) Remove(rr *ResourceRecord) (err error) {
	if rr.Type == RecordTypeSOA {
		zone.SOA = RDataSOA{}
	} else {
		if zone.Records.remove(rr) {
			err = zone.Save()
		}
	}
	return err
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
		suffixOrigin = "." + zone.Name

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

			if v == zone.Name {
				v = "@"
			} else if strings.HasSuffix(v, suffixOrigin) {
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
			if strings.HasSuffix(v, suffixOrigin) {
				v = strings.TrimSuffix(v, suffixOrigin)
			}
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
				minfo.RMailBox, minfo.EmailBox)

		case RecordTypeMX:
			mx, ok = rr.Value.(*RDataMX)
			if !ok {
				err = errors.New("invalid record value for MX")
				break
			}
			v = mx.Exchange
			if strings.HasSuffix(v, suffixOrigin) {
				v = strings.TrimSuffix(v, suffixOrigin)
			}
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
			v = srv.Target
			if strings.HasSuffix(v, suffixOrigin) {
				v = strings.TrimSuffix(v, suffixOrigin)
			}
			n, err = fmt.Fprintf(out,
				"%s %d %s SRV %d %d %d %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				srv.Priority, srv.Weight,
				srv.Port, v)
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
func (zone *Zone) WriteTo(out io.Writer) (total int, err error) {
	var (
		logp = `Write`
		n    int
	)

	n, _ = fmt.Fprintf(out, "$ORIGIN %s\n", zone.Name)
	total += n

	if len(zone.SOA.MName) > 0 {
		n, err = fmt.Fprintf(out,
			"@ SOA %s %s %d %d %d %d %d\n",
			zone.SOA.MName, zone.SOA.RName, zone.SOA.Serial, zone.SOA.Refresh,
			zone.SOA.Retry, zone.SOA.Expire, zone.SOA.Minimum)
		if err != nil {
			return total, fmt.Errorf(`%s: %w`, logp, err)
		}
		total += n
	}

	// Save the origin records first.
	var listRR = zone.Records[zone.Name]
	if len(listRR) > 0 {
		n, err = zone.saveListRR(out, `@`, listRR)
		if err != nil {
			return total, fmt.Errorf(`%s: %w`, logp, err)
		}
		total += n
	}

	// Save the records ordered by name.
	var (
		names = make([]string, 0, len(zone.Records))

		dname string
	)
	for dname = range zone.Records {
		if dname == zone.Name {
			continue
		}
		names = append(names, dname)
	}
	sort.Strings(names)

	for _, dname = range names {
		listRR = zone.Records[dname]
		dname = strings.TrimSuffix(dname, `.`+zone.Name)
		n, err = zone.saveListRR(out, dname, listRR)
		if err != nil {
			return total, fmt.Errorf(`%s: %w`, logp, err)
		}
		total += n
	}
	return total, nil
}
