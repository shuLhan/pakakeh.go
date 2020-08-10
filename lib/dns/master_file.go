// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dns

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	libio "github.com/shuLhan/share/lib/io"
)

//
// MasterFile represent content of single master file.
//
type MasterFile struct {
	Path     string
	Name     string
	Records  masterRecords
	messages []*Message
}

func newMasterFile(file, name string) *MasterFile {
	return &MasterFile{
		Path:    file,
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
// Messages return all pre-generated DNS messages.
//
func (mf *MasterFile) Messages() []*Message {
	return mf.messages
}
