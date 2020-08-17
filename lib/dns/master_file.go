// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	libio "github.com/shuLhan/share/lib/io"
)

//
// MasterFile represent content of single master file.
//
type MasterFile struct {
	path     string
	Name     string
	Records  masterRecords
	messages []*Message
}

//
// NewMasterFile create and initialize new master file.
//
func NewMasterFile(file, name string) *MasterFile {
	return &MasterFile{
		path:    file,
		Name:    name,
		Records: make(masterRecords),
	}
}

//
// LoadMasterDir load DNS record from master (zone) formatted files in
// directory "dir".
// On success, it will return map of file name and MasterFile content as list
// of Message.
// On fail, it will return possible partially parse master files and an error.
//
func LoadMasterDir(dir string) (masterFiles map[string]*MasterFile, err error) {
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

	masterFiles = make(map[string]*MasterFile)

	for x := 0; x < len(fis); x++ {
		if fis[x].IsDir() {
			continue
		}

		// Ignore file that start with "." .
		name := fis[x].Name()
		if name[0] == '.' {
			continue
		}

		masterFilePath := filepath.Join(dir, name)

		masterFile, err := ParseMasterFile(masterFilePath, "", 0)
		if err != nil {
			return masterFiles, fmt.Errorf("LoadMasterDir %q: %w", dir, err)
		}

		masterFiles[name] = masterFile
	}

	err = d.Close()
	if err != nil {
		return masterFiles, fmt.Errorf(" LoadMasterDir %q: %w", dir, err)
	}

	return masterFiles, nil
}

//
// ParseMasterFile parse master file and return it as list of Message.
// The file name will be assumed as origin if parameter origin or $ORIGIN is
// not set.
//
func ParseMasterFile(file, origin string, ttl uint32) (*MasterFile, error) {
	var err error

	m := newMasterParser(file)
	m.ttl = ttl

	if len(origin) > 0 {
		m.origin = origin
	} else {
		m.origin = path.Base(file)
	}

	m.origin = strings.ToLower(m.origin)

	m.reader, err = libio.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("ParseMasterFile %q: %w", file, err)
	}

	err = m.parse()
	if err != nil {
		return nil, fmt.Errorf("ParseMasterFile %q: %w", file, err)
	}

	m.out.Name = m.origin

	mf := m.out
	m.out = nil
	return mf, nil
}

//
// AddRR add new ResourceRecord to MasterFile.
//
func (mf *MasterFile) AddRR(rr *ResourceRecord) (err error) {
	mf.Records.add(rr)

	for _, msg := range mf.messages {
		if msg.Question.Name != rr.Name {
			continue
		}
		if msg.Question.Type != rr.Type {
			continue
		}
		if msg.Question.Class != rr.Class {
			continue
		}
		return msg.AddRR(rr)
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
	mf.messages = append(mf.messages, msg)
	return nil
}

//
// Delete the master file from storage.
//
func (mf *MasterFile) Delete() (err error) {
	return os.Remove(mf.path)
}

//
// Messages return all pre-generated DNS messages.
//
func (mf *MasterFile) Messages() []*Message {
	return mf.messages
}

//
// Remove a ResourceRecord from master file.
//
func (mf *MasterFile) Remove(rr *ResourceRecord) (err error) {
	isExist := mf.Records.remove(rr)
	if isExist {
		err = mf.Save()
	}
	return err
}

//
// Save the content of master records to file defined by path.
//
func (mf *MasterFile) Save() (err error) {
	out, err := os.OpenFile(mf.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC,
		0600)
	if err != nil {
		return err
	}

	var names []string

	fmt.Fprintf(out, "$ORIGIN %s.\n", mf.Name)

	// Save the origin records first.
	listRR := mf.Records[mf.Name]
	if len(listRR) > 0 {
		err = mf.saveListRR(out, "@", listRR)
		if err != nil {
			goto out
		}
	}

	// Save the records ordered by name.
	names = make([]string, 0, len(mf.Records))
	for dname := range mf.Records {
		if dname == mf.Name {
			continue
		}
		names = append(names, dname)
	}
	sort.Strings(names)

	for _, dname := range names {
		listRR := mf.Records[dname]
		dname = strings.TrimSuffix(dname, "."+mf.Name)
		err = mf.saveListRR(out, dname, listRR)
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

func (mf *MasterFile) saveListRR(
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
			QueryTypeMG, QueryTypeMR, QueryTypePTR:
			v, ok := rr.Value.(string)
			if !ok {
				err = errors.New("invalid record value for " +
					QueryTypeNames[rr.Type])
				break
			}
			if strings.HasSuffix(v, mf.Name) {
				v = strings.TrimSuffix(v, mf.Name)
			} else {
				v += "."
			}
			_, err = fmt.Fprintf(out, "%s %d %s %s %s\n",
				dname, rr.TTL, QueryClassName[rr.Class],
				QueryTypeNames[rr.Type], v)

		case QueryTypeSOA:
			soa, ok := rr.Value.(*RDataSOA)
			if !ok {
				err = errors.New("invalid record value for SOA")
				break
			}
			_, err = fmt.Fprintf(out,
				"@ SOA %s %s %d %d %d %d %d\n",
				soa.MName, soa.RName, soa.Serial, soa.Refresh,
				soa.Retry, soa.Expire, soa.Minimum)

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
