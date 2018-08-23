package dns

import (
	"io/ioutil"
	"runtime"

	libtext "github.com/shuLhan/share/lib/text"
)

const (
	HostsFilePOSIX   = "/etc/hosts"
	HostsFileWindows = "C:\\Windows\\System32\\Drivers\\etc\\hosts"
	defaultTTL       = 604800 // 7 days
)

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
// HostsLoad parse the content of hosts file as packed DNS message.
// If path is empty, it will load from the system hosts file.
//
func HostsLoad(path string) (msgs []*Message, err error) {
	if len(path) == 0 {
		path = GetSystemHosts()
	}

	in, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	msgs = parse(in)

	return
}

func newMessage(addr, hname *[]byte) *Message {
	newAddr := make([]byte, len(*addr))
	newHName := make([]byte, len(*hname))
	copy(newAddr, *addr)
	copy(newHName, *hname)

	qtype := QueryTypeA
	for x := 0; x < len(newAddr); x++ {
		if newAddr[x] == ':' {
			qtype = QueryTypeAAAA
			break
		}
	}

	msg := &Message{
		Header: &SectionHeader{
			QDCount: 1,
			ANCount: 1,
		},
		Question: &SectionQuestion{
			Name:  newHName,
			Type:  qtype,
			Class: QueryClassIN,
		},
		Answer: []*ResourceRecord{{
			Name:  newHName,
			Type:  qtype,
			Class: QueryClassIN,
			TTL:   defaultTTL,
			Text: &RDataText{
				v: newAddr,
			},
		}},
	}

	_, err := msg.MarshalBinary()
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
func parse(in []byte) (msgs []*Message) {
	var ok bool
	addr := make([]byte, 0, 32)
	hname := make([]byte, 0, 32)

	for x := 0; x < len(in); x++ {
		if libtext.IsSpace(in[x]) {
			continue
		}
		if in[x] == '#' {
			x = skipLine(x, in)
			continue
		}

		addr = addr[:0]
		x, ok = parseIPAddress(&addr, x, in)
		if !ok {
			x = skipLine(x, in)
			continue
		}

		hname = hname[:0]

		for ; x < len(in); x++ {
			x = skipBlanks(x, in)

			if in[x] == '\n' {
				break
			}

			if len(hname) > 0 {
				msg := newMessage(&addr, &hname)
				if msg != nil {
					msgs = append(msgs, msg)
				}
			}

			hname = hname[:0]
			x, ok = parseHostname(&hname, x, in)
			if !ok {
				hname = hname[:0]
				break
			}
		}

		if len(hname) > 0 {
			msg := newMessage(&addr, &hname)
			if msg != nil {
				msgs = append(msgs, msg)
			}
		}
		x = skipLine(x, in)
	}

	return
}

//
// parseIPAddress from input 'in' start from index 'x'.
// It will return true if address contains valid IPv4 or IPv6 characters;
// otherwise it will return false.
//
func parseIPAddress(addr *[]byte, x int, in []byte) (int, bool) {
	x, isIPv4, isIPv6 := parseDigitOrHex(addr, x, in)

	if isIPv4 {
		for ; x < len(in); x++ {
			if in[x] == ' ' || in[x] == '\t' {
				break
			}
			if in[x] == '.' || libtext.IsDigit(in[x]) {
				*addr = append(*addr, in[x])
				continue
			}
			return x, false
		}
		return x, true
	}
	if isIPv6 {
		for ; x < len(in); x++ {
			if in[x] == ' ' || in[x] == '\t' {
				break
			}
			if in[x] == ':' || libtext.IsHex(in[x]) {
				*addr = append(*addr, in[x])
				continue
			}
			return x, false
		}
		return x, true
	}

	return x, false
}

func parseDigitOrHex(addr *[]byte, x int, in []byte) (xx int, isIPv4, isIPv6 bool) {
	for ; x < len(in); x++ {
		if libtext.IsDigit(in[x]) {
			*addr = append(*addr, in[x])
			continue
		}
		if libtext.IsHex(in[x]) {
			*addr = append(*addr, in[x])
			x++
			return x, false, true
		}
		if in[x] == '.' {
			*addr = append(*addr, in[x])
			x++
			return x, true, false
		}
		if in[x] == ':' {
			*addr = append(*addr, in[x])
			x++
			return x, false, true
		}
		break
	}
	return x, false, false
}

func parseHostname(hname *[]byte, x int, in []byte) (int, bool) {
	if !libtext.IsAlnum(in[x]) {
		return x, false
	}
	*hname = append(*hname, in[x])
	x++
	for ; x < len(in); x++ {
		if libtext.IsSpace(in[x]) {
			return x, true
		}
		if in[x] == '-' || in[x] == '.' || libtext.IsAlnum(in[x]) {
			*hname = append(*hname, in[x])
			continue
		}
		break
	}
	return x, false
}

func skipBlanks(x int, in []byte) int {
	for ; x < len(in); x++ {
		if in[x] == ' ' || in[x] == '\t' {
			continue
		}
		break
	}
	return x
}

func skipLine(x int, in []byte) int {
	for ; x < len(in); x++ {
		if in[x] != '\n' {
			continue
		}
		break
	}
	return x
}
