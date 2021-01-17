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

//
// ZoneFile represent content of single zone file.
// A zone file contains at least one SOA record.
//
type ZoneFile struct {
	Path     string `json:"-"`
	Name     string
	SOA      ResourceRecord
	Records  zoneRecords
	messages []*Message
}

//
// NewZoneFile create and initialize new zone file.
//
func NewZoneFile(file, name string) *ZoneFile {
	return &ZoneFile{
		Path: file,
		Name: name,
		SOA: ResourceRecord{
			Type: QueryTypeSOA,
			Value: &RDataSOA{
				MName: name,
			},
		},
		Records: make(zoneRecords),
	}
}

//
// LoadMasterDir load DNS record from master (zone) formatted files in
// directory "dir".
// On success, it will return map of file name and ZoneFile content as list
// of Message.
// On fail, it will return possible partially parse master file and an error.
//
func LoadMasterDir(dir string) (zoneFiles map[string]*ZoneFile, err error) {
	if len(dir) == 0 {
		return nil, nil
	}

	d, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("LoadMasterDir: %w", err)
	}

	fis, err := d.Readdir(0)
	if err != nil {
		err = d.Close()
		if err != nil {
			return nil, fmt.Errorf("LoadMasterDir: %w", err)
		}
		return nil, fmt.Errorf("LoadMasterDir: %w", err)
	}

	zoneFiles = make(map[string]*ZoneFile)

	for x := 0; x < len(fis); x++ {
		if fis[x].IsDir() {
			continue
		}

		// Ignore file that start with "." .
		name := fis[x].Name()
		if name[0] == '.' {
			continue
		}

		zoneFilePath := filepath.Join(dir, name)

		zoneFile, err := ParseZoneFile(zoneFilePath, "", 0)
		if err != nil {
			return zoneFiles, fmt.Errorf("LoadMasterDir %q: %w", dir, err)
		}

		zoneFiles[name] = zoneFile
	}

	err = d.Close()
	if err != nil {
		return zoneFiles, fmt.Errorf(" LoadMasterDir %q: %w", dir, err)
	}

	return zoneFiles, nil
}

//
// ParseZoneFile parse zone file and return it as list of Message.
// The file name will be assumed as origin if parameter origin or $ORIGIN is
// not set.
//
func ParseZoneFile(file, origin string, ttl uint32) (*ZoneFile, error) {
	var err error

	m := newMasterParser(file)
	m.ttl = ttl

	if len(origin) > 0 {
		m.origin = origin
	} else {
		m.origin = filepath.Base(file)
	}

	m.origin = strings.ToLower(m.origin)

	m.reader, err = libio.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("ParseZoneFile %q: %w", file, err)
	}

	err = m.parse()
	if err != nil {
		return nil, fmt.Errorf("ParseZoneFile %q: %w", file, err)
	}

	m.zone.Name = m.origin

	zone := m.zone
	m.zone = nil
	return zone, nil
}

//
// Add add new ResourceRecord to ZoneFile.
//
func (zone *ZoneFile) Add(rr *ResourceRecord) (err error) {
	if rr.Type == QueryTypeSOA {
		zone.SOA = *rr
	} else {
		zone.Records.add(rr)
	}

	for _, msg := range zone.messages {
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

	msg := &Message{
		Header: SectionHeader{
			IsAA:    true,
			QDCount: 1,
			ANCount: 1,
		},
		Question: SectionQuestion{
			Name:  rr.Name,
			Type:  rr.Type,
			Class: rr.Class,
		},
		Answer: []ResourceRecord{*rr},
	}
	zone.messages = append(zone.messages, msg)
	return nil
}

//
// Delete the zone file from storage.
//
func (zone *ZoneFile) Delete() (err error) {
	return os.Remove(zone.Path)
}

//
// Messages return all pre-generated DNS messages.
//
func (zone *ZoneFile) Messages() []*Message {
	return zone.messages
}

//
// Remove a ResourceRecord from zone file.
//
func (zone *ZoneFile) Remove(rr *ResourceRecord) (err error) {
	if rr.Type == QueryTypeSOA {
		zone.SOA = ResourceRecord{
			Type:  QueryTypeSOA,
			Value: &RDataSOA{},
		}
	} else {
		if zone.Records.remove(rr) {
			err = zone.Save()
		}
	}
	return err
}

//
// Save the content of zone records to file defined by path.
//
func (zone *ZoneFile) Save() (err error) {
	out, err := os.OpenFile(zone.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600)
	if err != nil {
		return err
	}

	var (
		names  []string
		listRR []*ResourceRecord
	)

	fmt.Fprintf(out, "$ORIGIN %s.\n", zone.Name)

	soa, ok := zone.SOA.Value.(*RDataSOA)
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
	for dname := range zone.Records {
		if dname == zone.Name {
			continue
		}
		names = append(names, dname)
	}
	sort.Strings(names)

	for _, dname := range names {
		listRR := zone.Records[dname]
		dname = strings.TrimSuffix(dname, "."+zone.Name)
		err = zone.saveListRR(out, dname, listRR)
		if err != nil {
			break
		}
	}
out:
	errc := out.Close()
	if errc != nil {
		if err == nil {
			err = errc
		}
	}
	return err
}

func (zone *ZoneFile) saveListRR(
	out *os.File, dname string, listRR []*ResourceRecord,
) (err error) {
	for x, rr := range listRR {
		if x > 0 {
			dname = "\t"
		}
		switch rr.Type {
		case QueryTypeA, QueryTypeNULL, QueryTypeTXT,
			QueryTypeAAAA:
			_, err = fmt.Fprintf(out, "%s %d %s %s %s\n",
				dname, rr.TTL, QueryClassName[rr.Class],
				QueryTypeNames[rr.Type], rr.Value.(string))

		case QueryTypeNS, QueryTypeCNAME, QueryTypeMB,
			QueryTypeMG, QueryTypeMR:
			v, ok := rr.Value.(string)
			if !ok {
				err = errors.New("invalid record value for " +
					QueryTypeNames[rr.Type])
				break
			}
			if strings.HasSuffix(v, zone.Name) {
				v = strings.TrimSuffix(v, "."+zone.Name)
			} else {
				v += "."
			}
			_, err = fmt.Fprintf(out, "%s %d %s %s %s\n",
				dname, rr.TTL, QueryClassName[rr.Class],
				QueryTypeNames[rr.Type], v)

		case QueryTypePTR:
			v, ok := rr.Value.(string)
			if !ok {
				err = errors.New("invalid record value for " +
					QueryTypeNames[rr.Type])
				break
			}
			if strings.HasSuffix(v, zone.Name) {
				v = strings.TrimSuffix(v, "."+zone.Name)
			} else {
				v += "."
			}
			_, err = fmt.Fprintf(out, "%s. %d IN PTR %s\n",
				rr.Name, rr.TTL, v)

		case QueryTypeWKS:
			wks, ok := rr.Value.(*RDataWKS)
			if !ok {
				err = errors.New("invalid record value for WKS")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s WKS %s %d %s\n",
				dname, rr.TTL, QueryClassName[rr.Class],
				wks.Address, wks.Protocol, wks.BitMap)

		case QueryTypeHINFO:
			hinfo, ok := rr.Value.(*RDataHINFO)
			if !ok {
				err = errors.New("invalid record value for HINFO")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s HINFO %s %s\n",
				dname, rr.TTL, QueryClassName[rr.Class],
				hinfo.CPU, hinfo.OS)

		case QueryTypeMINFO:
			minfo, ok := rr.Value.(*RDataMINFO)
			if !ok {
				err = errors.New("invalid record value for MINFO")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s MINFO %s %s\n",
				dname, rr.TTL, QueryClassName[rr.Class],
				minfo.RMailBox, minfo.EmailBox)

		case QueryTypeMX:
			mx, ok := rr.Value.(*RDataMX)
			if !ok {
				err = errors.New("invalid record value for MX")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s MX %d %s.\n",
				dname, rr.TTL, QueryClassName[rr.Class],
				mx.Preference, mx.Exchange)

		case QueryTypeSRV:
			srv, ok := rr.Value.(*RDataSRV)
			if !ok {
				err = errors.New("invalid record value for SRV")
				break
			}
			_, err = fmt.Fprintf(out,
				"%s %d %s SRV %d %d %d %s.\n",
				dname, rr.TTL, QueryClassName[rr.Class],
				srv.Priority, srv.Weight,
				srv.Port, srv.Target)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
