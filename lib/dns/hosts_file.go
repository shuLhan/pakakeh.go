package dns

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	libio "github.com/shuLhan/share/lib/io"
)

// List of known hosts file by OS.
const (
	HostsFilePOSIX   = "/etc/hosts"
	HostsFileWindows = "C:\\Windows\\System32\\Drivers\\etc\\hosts"
	defaultTTL       = 604800 // 7 days
)

//
// HostsFile represent content of single hosts file.
//
type HostsFile struct {
	Path    string `json:"-"`
	Name    string
	Records []*ResourceRecord `json:"-"`
	out     *os.File
}

//
// NewHostsFile create and store the host records in file defined by "path".
//
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

//
// GetSystemHosts return path to system hosts file.
//
func GetSystemHosts() string {
	if runtime.GOOS == "windows" {
		return HostsFileWindows
	}
	return HostsFilePOSIX
}

//
// LoadHostsDir load all of hosts formatted files inside a directory.
// On success, it will return map of filename and the content of hosts file as
// list of Message.
// On fail, it will return partial loadeded hosts files and an error.
//
func LoadHostsDir(dir string) (hostsFiles map[string]*HostsFile, err error) {
	if len(dir) == 0 {
		return nil, nil
	}

	d, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("LoadHostsDir %q: %w", dir, err)
	}

	fis, err := d.Readdir(0)
	if err != nil {
		log.Println("dns: LoadHostsDir: ", err)
		errClose := d.Close()
		if errClose != nil {
			log.Println("dns: LoadHostsDir: ", errClose)
		}
		return nil, fmt.Errorf("LoadHostsDir %q: %w", dir, err)
	}

	hostsFiles = make(map[string]*HostsFile)

	for x := 0; x < len(fis); x++ {
		if fis[x].IsDir() {
			continue
		}

		// Ignore file that start with "." .
		name := fis[x].Name()
		if name[0] == '.' {
			continue
		}

		hostsFilePath := filepath.Join(dir, name)

		hfile, err := ParseHostsFile(hostsFilePath)
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

//
// ParseHostsFile parse the content of hosts file as packed DNS message.
// If path is empty, it will load from the system hosts file.
//
func ParseHostsFile(path string) (hfile *HostsFile, err error) {
	if len(path) == 0 {
		path = GetSystemHosts()
	}

	reader, err := libio.NewReader(path)
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

//
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
//
func parse(reader *libio.Reader) (listRR []*ResourceRecord) {
	var (
		seps  = []byte{'\t', '\v', ' '}
		terms = []byte{'\n', '\f', '#'}
	)

	for {
		c := reader.SkipSpaces()
		if c == 0 {
			break
		}
		if c == '#' {
			reader.SkipLine()
			continue
		}

		addr, isTerm, c := reader.ReadUntil(seps, terms)
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
			c := reader.SkipSpaces()
			if c == 0 {
				break
			}
			if c == '#' {
				reader.SkipLine()
				break
			}
			hname, isTerm, c := reader.ReadUntil(seps, terms)
			if len(hname) > 0 {
				qtype := GetQueryTypeFromAddress(addr)
				if qtype == 0 {
					continue
				}
				rr := &ResourceRecord{
					Name:  string(bytes.ToLower(hname)),
					Type:  qtype,
					Class: QueryClassIN,
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

//
// AppendAndSaveRecord append new record and save it to hosts file.
//
func (hfile *HostsFile) AppendAndSaveRecord(rr *ResourceRecord) (err error) {
	f, err := os.OpenFile(
		hfile.Path,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0600,
	)
	if err != nil {
		return err
	}

	ipAddress, ok := rr.Value.(string)
	if ok {
		_, err = fmt.Fprintf(f, "%s %s\n", ipAddress, rr.Name)
	}

	errClose := f.Close()
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

//
// Delete the hosts file from the storage.
//
func (hfile *HostsFile) Delete() (err error) {
	return os.RemoveAll(hfile.Path)
}

//
// Names return all hosts domain names.
//
func (hfile *HostsFile) Names() (names []string) {
	names = make([]string, 0, len(hfile.Records))

	for _, rr := range hfile.Records {
		names = append(names, rr.Name)
	}

	return names
}

//
// RemoveRecord remove single record from hosts file by domain name.
// It will return true if record found and removed.
//
func (hfile *HostsFile) RemoveRecord(dname string) bool {
	for x := 0; x < len(hfile.Records); x++ {
		if hfile.Records[x].Name != dname {
			continue
		}
		copy(hfile.Records[x:], hfile.Records[x+1:])
		hfile.Records[len(hfile.Records)-1] = nil
		hfile.Records = hfile.Records[:len(hfile.Records)-1]
		return true
	}
	return false
}

//
// Save the hosts records into the file defined by field "Path".
//
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

	for _, rr := range hfile.Records {
		if len(rr.Name) == 0 || rr.Value == nil {
			continue
		}
		ipAddress, ok := rr.Value.(string)
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
