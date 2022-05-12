// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	libio "github.com/shuLhan/share/lib/io"
)

// Zone represent a group of domain names shared a single root domain.
// A Zone contains at least one SOA record.
type Zone struct {
	Records  zoneRecords
	Path     string `json:"-"`
	Name     string
	messages []*Message
	SOA      ResourceRecord
}

// NewZone create and initialize new zone.
func NewZone(file, name string) *Zone {
	return &Zone{
		Path: file,
		Name: name,
		SOA: ResourceRecord{
			Type: RecordTypeSOA,
			Value: &RDataSOA{
				MName: name,
			},
		},
		Records: make(zoneRecords),
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
		m *zoneParser = newZoneParser(file)
	)

	m.ttl = ttl

	if len(origin) > 0 {
		m.origin = origin
	} else {
		m.origin = filepath.Base(file)
	}

	m.origin = strings.ToLower(m.origin)

	m.reader, err = libio.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("ParseZone %q: %w", file, err)
	}

	err = m.parse()
	if err != nil {
		return nil, fmt.Errorf("ParseZone %q: %w", file, err)
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
	)

	if rr.Type == RecordTypeSOA {
		zone.SOA = *rr
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
		zone.SOA = ResourceRecord{
			Type:  RecordTypeSOA,
			Value: &RDataSOA{},
		}
	} else {
		if zone.Records.remove(rr) {
			err = zone.Save()
		}
	}
	return err
}

// Save the content of zone records to file defined by Path.
func (zone *Zone) Save() (err error) {
	var (
		out *os.File
	)

	out, err = os.OpenFile(zone.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	var (
		soa    *RDataSOA
		dname  string
		names  []string
		listRR []*ResourceRecord
		errc   error
		ok     bool
	)

	fmt.Fprintf(out, "$ORIGIN %s.\n", zone.Name)

	soa, ok = zone.SOA.Value.(*RDataSOA)
	if ok && len(soa.MName) > 0 {
		_, err = fmt.Fprintf(out,
			"@ SOA %s. %s. %d %d %d %d %d\n",
			soa.MName, soa.RName, soa.Serial, soa.Refresh,
			soa.Retry, soa.Expire, soa.Minimum)
		if err != nil {
			goto out
		}
	}

	// Save the origin records first.
	listRR = zone.Records[zone.Name]
	if len(listRR) > 0 {
		err = zone.saveListRR(out, "@", listRR)
		if err != nil {
			goto out
		}
	}

	// Save the records ordered by name.
	names = make([]string, 0, len(zone.Records))
	for dname = range zone.Records {
		if dname == zone.Name {
			continue
		}
		names = append(names, dname)
	}
	sort.Strings(names)

	for _, dname = range names {
		listRR = zone.Records[dname]
		dname = strings.TrimSuffix(dname, "."+zone.Name)
		err = zone.saveListRR(out, dname, listRR)
		if err != nil {
			break
		}
	}
out:
	errc = out.Close()
	if errc != nil {
		if err == nil {
			err = errc
		}
	}
	return err
}

func (zone *Zone) saveListRR(out *os.File, dname string, listRR []*ResourceRecord) (err error) {
	var (
		hinfo *RDataHINFO
		minfo *RDataMINFO
		mx    *RDataMX
		rr    *ResourceRecord
		srv   *RDataSRV
		wks   *RDataWKS
		v     string
		x     int
		ok    bool
	)

	for x, rr = range listRR {
		if x > 0 {
			dname = "\t"
		}
		switch rr.Type {
		case RecordTypeA, RecordTypeNULL, RecordTypeAAAA:
			_, err = fmt.Fprintf(out, "%s %d %s %s %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				RecordTypeNames[rr.Type], rr.Value.(string))

		case RecordTypeTXT:
			_, err = fmt.Fprintf(out, "%s %d %s %s %q\n",
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
			if strings.HasSuffix(v, zone.Name) {
				v = strings.TrimSuffix(v, "."+zone.Name)
			} else {
				v += "."
			}
			_, err = fmt.Fprintf(out, "%s %d %s %s %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				RecordTypeNames[rr.Type], v)

		case RecordTypePTR:
			v, ok = rr.Value.(string)
			if !ok {
				err = errors.New("invalid record value for " +
					RecordTypeNames[rr.Type])
				break
			}
			if strings.HasSuffix(v, zone.Name) {
				v = strings.TrimSuffix(v, "."+zone.Name)
			} else {
				v += "."
			}
			_, err = fmt.Fprintf(out, "%s. %d IN PTR %s\n",
				rr.Name, rr.TTL, v)

		case RecordTypeWKS:
			wks, ok = rr.Value.(*RDataWKS)
			if !ok {
				err = errors.New("invalid record value for WKS")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s WKS %s %d %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				wks.Address, wks.Protocol, wks.BitMap)

		case RecordTypeHINFO:
			hinfo, ok = rr.Value.(*RDataHINFO)
			if !ok {
				err = errors.New("invalid record value for HINFO")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s HINFO %s %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				hinfo.CPU, hinfo.OS)

		case RecordTypeMINFO:
			minfo, ok = rr.Value.(*RDataMINFO)
			if !ok {
				err = errors.New("invalid record value for MINFO")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s MINFO %s %s\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				minfo.RMailBox, minfo.EmailBox)

		case RecordTypeMX:
			mx, ok = rr.Value.(*RDataMX)
			if !ok {
				err = errors.New("invalid record value for MX")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s MX %d %s.\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				mx.Preference, mx.Exchange)

		case RecordTypeSRV:
			srv, ok = rr.Value.(*RDataSRV)
			if !ok {
				err = errors.New("invalid record value for SRV")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s SRV %d %d %d %s.\n",
				dname, rr.TTL, RecordClassName[rr.Class],
				srv.Priority, srv.Weight,
				srv.Port, srv.Target)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
