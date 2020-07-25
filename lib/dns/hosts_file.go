package dns

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"

	"github.com/shuLhan/share/lib/ascii"
	libio "github.com/shuLhan/share/lib/io"
	libnet "github.com/shuLhan/share/lib/net"
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
	Path     string
	Name     string
	Messages []*Message
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

		hostsFile, err := ParseHostsFile(hostsFilePath)
		if err != nil {
			return hostsFiles, fmt.Errorf("LoadHostsDir %q: %w", dir, err)
		}

		hostsFiles[name] = hostsFile
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
func ParseHostsFile(path string) (hostsFile *HostsFile, err error) {
	if len(path) == 0 {
		path = GetSystemHosts()
	}

	reader, err := libio.NewReader(path)
	if err != nil {
		return nil, fmt.Errorf("ParseHostsFile %q: %w", path, err)
	}

	hostsFile = &HostsFile{
		Path: path,
		Name: filepath.Base(path),
	}

	hostsFile.Messages = parse(reader)

	return hostsFile, nil
}

func newMessage(addr, hname []byte) *Message {
	if !libnet.IsHostnameValid(hname, false) {
		return nil
	}
	ip := net.ParseIP(string(addr))
	if ip == nil {
		return nil
	}

	qtype := QueryTypeA
	for x := 0; x < len(addr); x++ {
		if addr[x] == ':' {
			qtype = QueryTypeAAAA
			break
		}
	}

	ascii.ToLower(&hname)
	rrName := make([]byte, len(hname))
	copy(rrName, hname)

	msg := &Message{
		Header: SectionHeader{
			IsAA:    true,
			QDCount: 1,
			ANCount: 1,
		},
		Question: SectionQuestion{
			Name:  hname,
			Type:  qtype,
			Class: QueryClassIN,
		},
		Answer: []ResourceRecord{{
			Name:  rrName,
			Type:  qtype,
			Class: QueryClassIN,
			TTL:   defaultTTL,
			Text:  addr,
		}},
	}

	_, err := msg.Pack()
	if err != nil {
		return nil
	}

	return msg
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
func parse(reader *libio.Reader) (msgs []*Message) {
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
				msg := newMessage(addr, hname)
				if msg != nil {
					msgs = append(msgs, msg)
				}
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

	return msgs
}
