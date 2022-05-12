// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	libio "github.com/shuLhan/share/lib/io"
)

// List of known hosts file by OS.
const (
	HostsFilePOSIX   = "/etc/hosts"
	HostsFileWindows = "C:\\Windows\\System32\\Drivers\\etc\\hosts"
	defaultTTL       = 604800 // 7 days
)

// HostsFile represent content of single hosts file.
type HostsFile struct {
	out     *os.File
	Path    string `json:"-"`
	Name    string
	Records []*ResourceRecord `json:"-"`
}

// NewHostsFile create and store the host records in file defined by "path".
func NewHostsFile(path string, records []*ResourceRecord) (
	hfile *HostsFile, err error,
) {
	hfile = &HostsFile{
		Path:    path,
		Name:    filepath.Base(path),
		Records: records,
	}

	err = hfile.Save()

	return hfile, err
}

// GetSystemHosts return path to system hosts file.
func GetSystemHosts() string {
	if runtime.GOOS == "windows" {
		return HostsFileWindows
	}
	return HostsFilePOSIX
}

// LoadHostsDir load all of hosts formatted files inside a directory.
// On success, it will return map of filename and the content of hosts file as
// list of Message.
// On fail, it will return partial loadeded hosts files and an error.
func LoadHostsDir(dir string) (hostsFiles map[string]*HostsFile, err error) {
	if len(dir) == 0 {
		return nil, nil
	}

	var (
		d             *os.File
		hfile         *HostsFile
		fi            os.FileInfo
		fis           []os.FileInfo
		name          string
		hostsFilePath string
		errClose      error
	)

	d, err = os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("LoadHostsDir %q: %w", dir, err)
	}

	fis, err = d.Readdir(0)
	if err != nil {
		log.Println("dns: LoadHostsDir: ", err)
		errClose = d.Close()
		if errClose != nil {
			log.Println("dns: LoadHostsDir: ", errClose)
		}
		return nil, fmt.Errorf("LoadHostsDir %q: %w", dir, err)
	}

	hostsFiles = make(map[string]*HostsFile)

	for _, fi = range fis {
		if fi.IsDir() {
			continue
		}

		// Ignore file that start with "." .
		name = fi.Name()
		if name[0] == '.' {
			continue
		}

		hostsFilePath = filepath.Join(dir, name)

		hfile, err = ParseHostsFile(hostsFilePath)
		if err != nil {
			return hostsFiles, fmt.Errorf("LoadHostsDir %q: %w", dir, err)
		}

		hostsFiles[name] = hfile
	}

	err = d.Close()
	if err != nil {
		return hostsFiles, fmt.Errorf("LoadHostsDir %q: %w", dir, err)
	}

	return hostsFiles, nil
}

// ParseHostsFile parse the content of hosts file as packed DNS message.
// If path is empty, it will load from the system hosts file.
func ParseHostsFile(path string) (hfile *HostsFile, err error) {
	var (
		reader *libio.Reader
	)

	if len(path) == 0 {
		path = GetSystemHosts()
	}

	reader, err = libio.NewReader(path)
	if err != nil {
		return nil, fmt.Errorf("ParseHostsFile %q: %w", path, err)
	}

	hfile = &HostsFile{
		Path:    path,
		Name:    filepath.Base(path),
		Records: parse(reader),
	}

	return hfile, nil
}

// Fields of the entry are separated by any number of blanks and/or tab
// characters.
// Text from a "#" character until the end of the line is a comment, and is
// ignored.
// Host names may contain only alphanumeric characters, minus signs ("-"), and
// periods (".").
// They must begin with an alphabetic character and end with an alphanumeric
// character.
// Optional aliases provide for name changes, alternate spellings,
// shorter hostnames, or generic hostnames (for example, localhost). [1]
//
// [1] man 5 hosts
func parse(reader *libio.Reader) (listRR []*ResourceRecord) {
	var (
		seps  = []byte{'\t', '\v', ' '}
		terms = []byte{'\n', '\f', '#'}

		rr     *ResourceRecord
		addr   []byte
		hname  []byte
		rtype  RecordType
		c      byte
		isTerm bool
	)

	for {
		c = reader.SkipSpaces()
		if c == 0 {
			break
		}
		if c == '#' {
			reader.SkipLine()
			continue
		}

		addr, isTerm, c = reader.ReadUntil(seps, terms)
		if isTerm {
			if c == 0 {
				break
			}
			if c == '#' {
				reader.SkipLine()
			}
			continue
		}

		for {
			c = reader.SkipSpaces()
			if c == 0 {
				break
			}
			if c == '#' {
				reader.SkipLine()
				break
			}
			hname, isTerm, c = reader.ReadUntil(seps, terms)
			if len(hname) > 0 {
				rtype = RecordTypeFromAddress(addr)
				if rtype == 0 {
					continue
				}
				rr = &ResourceRecord{
					Name:  string(bytes.ToLower(hname)),
					Type:  rtype,
					Class: RecordClassIN,
					TTL:   defaultTTL,
					Value: string(addr),
				}
				listRR = append(listRR, rr)
			}
			if isTerm {
				if c == 0 {
					break
				}
				if c == '#' {
					reader.SkipLine()
				}
				break
			}
		}
	}

	return listRR
}

// AppendAndSaveRecord append new record and save it to hosts file.
func (hfile *HostsFile) AppendAndSaveRecord(rr *ResourceRecord) (err error) {
	var (
		f         *os.File
		ipAddress string
		errClose  error
		ok        bool
	)

	f, err = os.OpenFile(
		hfile.Path,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0600,
	)
	if err != nil {
		return err
	}

	ipAddress, ok = rr.Value.(string)
	if ok {
		_, err = fmt.Fprintf(f, "%s %s\n", ipAddress, rr.Name)
	}

	errClose = f.Close()
	if errClose != nil {
		if err == nil {
			err = errClose
		}
	}

	if err == nil {
		hfile.Records = append(hfile.Records, rr)
	}

	return err
}

// Delete the hosts file from the storage.
func (hfile *HostsFile) Delete() (err error) {
	return os.RemoveAll(hfile.Path)
}

// Get the first resource record that match with domain name and/or value.
// The value parameter is optional, if its empty, then only the first record
// that match with domain name that will be returned.
//
// If no record matched, it will return nil.
func (hfile *HostsFile) Get(dname, value string) (rr *ResourceRecord) {
	var (
		rrValue string
		ok      bool
	)

	dname = strings.ToLower(dname)

	if len(value) != 0 {
		value = strings.ToLower(value)
	}

	for _, rr = range hfile.Records {
		if rr.Name != dname {
			continue
		}
		if len(value) == 0 {
			return rr
		}
		rrValue, ok = rr.Value.(string)
		if !ok {
			continue
		}
		if rrValue != value {
			continue
		}
		return rr
	}
	return nil
}

// Names return all hosts domain names.
func (hfile *HostsFile) Names() (names []string) {
	var (
		rr *ResourceRecord
	)

	names = make([]string, 0, len(hfile.Records))

	for _, rr = range hfile.Records {
		names = append(names, rr.Name)
	}

	return names
}

// RemoveRecord remove single record from hosts file by domain name.
// It will return true if record found and removed.
func (hfile *HostsFile) RemoveRecord(dname string) (rr *ResourceRecord) {
	var (
		x int
	)
	for x, rr = range hfile.Records {
		if rr.Name != dname {
			continue
		}
		copy(hfile.Records[x:], hfile.Records[x+1:])
		hfile.Records[len(hfile.Records)-1] = nil
		hfile.Records = hfile.Records[:len(hfile.Records)-1]
		return rr
	}
	return nil
}

// Save the hosts records into the file defined by field "Path".
func (hfile *HostsFile) Save() (err error) {
	if hfile.out == nil {
		hfile.out, err = os.OpenFile(
			hfile.Path,
			os.O_CREATE|os.O_TRUNC|os.O_RDWR,
			0600,
		)
	} else {
		err = hfile.out.Truncate(0)
	}
	if err != nil {
		return err
	}

	var (
		rr        *ResourceRecord
		ipAddress string
		ok        bool
	)

	for _, rr = range hfile.Records {
		if len(rr.Name) == 0 || rr.Value == nil {
			continue
		}
		ipAddress, ok = rr.Value.(string)
		if !ok {
			continue
		}
		_, err = fmt.Fprintf(hfile.out, "%s %s\n", ipAddress, rr.Name)
		if err != nil {
			return err
		}
	}

	err = hfile.out.Close()
	hfile.out = nil
	if err != nil {
		return err
	}

	return nil
}
