// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package dns

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
)

// List of known parameter names for SVCB.
const (
	svcbKeyNameMandatory     = `mandatory`
	svcbKeyNameALPN          = `alpn`
	svcbKeyNameNoDefaultALPN = `no-default-alpn`
	svcbKeyNamePort          = `port`
	svcbKeyNameIpv4hint      = `ipv4hint`
	svcbKeyNameEch           = `ech`
	svcbKeyNameIpv6hint      = `ipv6hint`
)

const (
	svcbKeyIDMandatory     int = 0
	svcbKeyIDALPN          int = 1
	svcbKeyIDNoDefaultALPN int = 2
	svcbKeyIDPort          int = 3
	svcbKeyIDIpv4hint      int = 4
	svcbKeyIDEch           int = 5
	svcbKeyIDIpv6hint      int = 6
)

// RDataSVCB the resource record for type 64 [SVCB RR].
// Format of SVCB RDATA,
//
//	+-------------+
//	| SvcPriority | 2-octets.
//	+-------------+
//	/ TargetName  / A <domain-name>.
//	/             /
//	+-------------+
//	/ SvcParams   / A <character-string>.
//	/             /
//	+-------------+
//
// SVCB RR has two modes: AliasMode and ServiceMode.
// SvcPriority with value 0 indicates SVCB RR as AliasMode.
// SvcParams SHALL be used only for ServiceMode.
//
// The SvcParams contains the SVCB parameter key and value.
// Format of SvcParams,
//
//	+-------------------+
//	| SvcParamKey       | ; 2-octets.
//	+-------------------+
//	| SvcParamKeyLength | ; 2-octets, indicates the length of SvcParamValue.
//	+-------------------+
//	/ SvcParamValue     / ; Dynamic value based on the key.
//	/                   /
//	+-------------------+
//
// The RDATA considered malformed if:
//
//   - RDATA end at SvcParamKeyLength with non-zero value.
//   - SvcParamKey are not in increasing numeric order, for example: 1, 3, 2.
//   - Contains duplicate SvcParamKey.
//   - Contains invalid SvcParamValue format.
//
// Currently, there are six known keys,
//
//   - mandatory (0): define list of keys that must be exists on TargetName.
//     Each value is stored as 2-octets of its numeric ID.
//   - alpn (1): define list of Application-Layer Protocol Negotiation
//     (ALPN) supported by TargetName.
//     Each alpn is stored as combination of 2-octets length and its value.
//   - no-default-alpn (2): indicates that no default ALPN exists on
//     TargetName.
//     This key does not have value.
//   - port (3): define TCP or UDP port of TargetName.
//     The port value is encoded in 2-octets.
//   - ipv4hint (4): contains list of IPv4 addresses of TargetName.
//     Each IPv4 address is encoded in 4-octets.
//   - ech (5): Reserved.
//   - ipv6hint (6): contains list of IPv6 addresses of TargetName.
//     Each IPv6 address is encoded in 8-octets.
//
// A generic key can be defined in zone file by prefixing the number with
// string "key".
// For example,
//
//	key123="hello"
//
// will be encoded in RDATA as 123 (2-octets), followed by 5 (length of
// value, 2-octets), and followed by "hello" (5-octets).
//
// # Example
//
// The domain "example.com" provides a service "foo.example.org" with
// priority 16 and with two mandatory parameters: "alpn" and "ipv4hint".
//
//	example.com.   SVCB   16 foo.example.org. (
//	                           alpn=h2,h3-19 mandatory=ipv4hint,alpn
//	                           ipv4hint=192.0.2.1
//	                         )
//
// The above zone record when encoded to RDATA (displayed in decimal for
// readability),
//
//	+----+-----------------+
//	| 16 / foo.example.org /
//	+----+-----------------+
//	; SvcPriority=16               (2 octets)
//	; TargetName="foo.example.org" (domain-name, max 255 octects)
//	+---+---+---+---+
//	| 0 | 4 | 1 | 4 |
//	+---+---+---+---+
//	; SvcParamKey=0 (mandatory)  (2 octets)
//	; length=4                   (2 octets)
//	; value[0]: 1 (alpn)         (2 octets)
//	; value[1]: 4 (ipv4hint)     (2 octets)
//	+---+---+---+----+---+-------+
//	| 1 | 9 | 2 | h2 | 5 | h3-19 |
//	+---+---+---+----+---+-------+
//	; SvcParamKey=1 (alpn)              (2 octets)
//	; length=9                          (2 octets)
//	; value[0]: length=2, value="h2"    (1 + 2 octets)
//	; value[1]: length=5, value="h3-19" (1 + 5 octets)
//	+---+---+-----------+
//	| 4 | 4 | 192.0.2.1 |
//	+---+---+-----------+
//	; SvcParamKey=4 (ipv4hint)  (2 octets)
//	; length=4                  (2 octets)
//	; value="192.0.2.1"         (4 octets)
//
// [SVCB RR]: https://datatracker.ietf.org/doc/html/rfc9460
type RDataSVCB struct {
	// Params contains service parameters indexed by key's ID.
	Params map[int][]string

	TargetName string
	Priority   uint16
}

// AddParam add parameter to service binding.
// It will return an error if key already exist or contains invalid value.
func (svcb *RDataSVCB) AddParam(key string, listValue []string) (err error) {
	var logp = `AddParam`

	var keyid = svcbKeyID(key)
	if keyid < 0 {
		return fmt.Errorf(`%s: unknown key %q`, logp, key)
	}

	var isExist bool

	_, isExist = svcb.Params[keyid]
	if isExist {
		return fmt.Errorf(`%s: duplicate key %q`, logp, key)
	}

	switch keyid {
	case svcbKeyIDMandatory:
		var (
			listKeyID = map[int]struct{}{}
			name      string
			gotid     int
		)
		for _, name = range listValue {
			gotid = svcbKeyID(name)
			if gotid < 0 {
				return fmt.Errorf(`%s: invalid mandatory key %q`, logp, name)
			}
			_, isExist = listKeyID[gotid]
			if isExist {
				return fmt.Errorf(`%s: duplicate mandatory key %q`, logp, name)
			}
			listKeyID[gotid] = struct{}{}
		}
		svcb.Params[keyid] = listValue

	case svcbKeyIDALPN:
		var name string
		for _, name = range listValue {
			if len(name) > math.MaxUint8 {
				return fmt.Errorf(`%s: ALPN value must not exceed %d: %q`, logp, math.MaxUint8, name)
			}
		}
		svcb.Params[keyid] = listValue

	case svcbKeyIDNoDefaultALPN:
		if len(listValue) != 0 {
			return fmt.Errorf(`%s: key no-default-alpn must not have values`, logp)
		}
		svcb.Params[keyid] = listValue

	case svcbKeyIDPort:
		if len(listValue) == 0 {
			return fmt.Errorf(`%s: missing port value`, logp)
		}
		if len(listValue) > 1 {
			return fmt.Errorf(`%s: multiple port values %q`, logp, listValue)
		}

		var port int64

		port, err = strconv.ParseInt(listValue[0], 10, 16)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
		if port < 0 || port > math.MaxUint16 {
			return fmt.Errorf(`%s: invalid port value %q`, logp, listValue[0])
		}
		svcb.Params[keyid] = listValue

	case svcbKeyIDIpv4hint, svcbKeyIDIpv6hint:
		if len(listValue) == 0 {
			return fmt.Errorf(`%s: missing %q value`, logp, key)
		}
		var (
			val string
			ip  net.IP
		)
		for _, val = range listValue {
			ip = net.ParseIP(val)
			if ip == nil {
				return fmt.Errorf(`%s: invalid IP %q`, logp, val)
			}
		}
		svcb.Params[keyid] = listValue

	case svcbKeyIDEch:
		// NO-OP.

	default:
		svcb.Params[keyid] = listValue
	}

	return nil
}

// WriteTo write the SVCB record as zone format to out.
func (svcb *RDataSVCB) WriteTo(out io.Writer) (_ int64, err error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, `SVCB %d %s`, svcb.Priority, svcb.TargetName)

	var (
		keys = svcb.keys()

		keyid int
	)
	for _, keyid = range keys {
		buf.WriteByte(' ')

		if keyid == svcbKeyIDNoDefaultALPN {
			buf.WriteString(svcbKeyNameNoDefaultALPN)
			continue
		}

		svcb.writeParam(&buf, keyid)
	}
	buf.WriteByte('\n')

	var n int

	n, err = out.Write(buf.Bytes())

	return int64(n), err
}

func (svcb *RDataSVCB) getParamKey(zp *zoneParser) (_ []byte, err error) {
	var logp = `getParamKey`

	for {
		err = zp.next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}
		if len(zp.token) != 0 {
			break
		}
	}
	return zp.token, nil
}

func (svcb *RDataSVCB) getParamValue(zp *zoneParser) (val []byte, err error) {
	var (
		logp = `getParamValue`

		lenToken int
		isQuoted bool
	)

	for {
		err = zp.next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		val = append(val, zp.token...)

		if isQuoted {
			lenToken = len(zp.token)
			if lenToken != 0 && zp.token[lenToken-1] == '"' {
				if lenToken >= 2 && zp.token[lenToken-2] == '\\' {
					// Double-quote is escaped.
					continue
				}
				break
			}
			continue
		}
		if zp.token[0] == '"' {
			isQuoted = true
			continue
		}
		break
	}

	if isQuoted {
		val = val[1 : len(val)-1]
	}

	return val, nil
}

// keys return the list of sorted parameter key.
func (svcb *RDataSVCB) keys() (listKey []int) {
	var key int
	for key = range svcb.Params {
		listKey = append(listKey, key)
	}
	sort.Ints(listKey)
	return listKey
}

func (svcb *RDataSVCB) pack(msg *Message) (n int) {
	n = len(msg.packet)

	msg.packet = binary.BigEndian.AppendUint16(msg.packet, svcb.Priority)

	_ = msg.packDomainName([]byte(svcb.TargetName), false)

	var (
		sortedKeys = svcb.keys()

		listValue []string
		keyid     int
	)
	for _, keyid = range sortedKeys {
		listValue = svcb.Params[keyid]

		switch keyid {
		case svcbKeyIDMandatory:
			svcb.packMandatory(msg, listValue)

		case svcbKeyIDALPN:
			svcb.packALPN(msg, listValue)

		case svcbKeyIDNoDefaultALPN:
			msg.packet = binary.BigEndian.AppendUint16(msg.packet,
				uint16(svcbKeyIDNoDefaultALPN))

		case svcbKeyIDPort:
			svcb.packPort(msg, listValue)

		case svcbKeyIDIpv4hint:
			svcb.packIpv4hint(msg, listValue)

		case svcbKeyIDEch:
			// NO-OP.

		case svcbKeyIDIpv6hint:
			svcb.packIpv6hint(msg, listValue)

		default:
			svcb.packGenericValue(keyid, msg, listValue)
		}
	}

	n = len(msg.packet) - n
	return n
}

func (svcb *RDataSVCB) packMandatory(msg *Message, listValue []string) {
	msg.packet = binary.BigEndian.AppendUint16(msg.packet,
		uint16(svcbKeyIDMandatory))

	var total = 2 * len(listValue)
	msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(total))

	var (
		listKeyID = make([]int, 0, len(listValue))
		keyName   string
		keyid     int
	)
	for _, keyName = range listValue {
		keyid = svcbKeyID(keyName)
		listKeyID = append(listKeyID, keyid)
	}
	sort.Ints(listKeyID)
	for _, keyid = range listKeyID {
		msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(keyid))
	}
}

func (svcb *RDataSVCB) packALPN(msg *Message, listValue []string) {
	var (
		val   string
		total int
	)
	for _, val = range listValue {
		total += 1 + len(val)
	}

	msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(svcbKeyIDALPN))
	msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(total))

	for _, val = range listValue {
		msg.packet = append(msg.packet, byte(len(val)))
		msg.packet = append(msg.packet, []byte(val)...)
	}
}

func (svcb *RDataSVCB) packPort(msg *Message, listValue []string) {
	var (
		port int64
		err  error
	)

	port, err = strconv.ParseInt(listValue[0], 10, 16)
	if err != nil {
		return
	}

	msg.packet = binary.BigEndian.AppendUint16(msg.packet,
		uint16(svcbKeyIDPort))
	msg.packet = binary.BigEndian.AppendUint16(msg.packet, 2)
	msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(port))
}

func (svcb *RDataSVCB) packIpv4hint(msg *Message, listValue []string) {
	msg.packet = binary.BigEndian.AppendUint16(msg.packet,
		uint16(svcbKeyIDIpv4hint))

	var total = 4 * len(listValue)
	msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(total))

	var val string

	for _, val = range listValue {
		msg.packIPv4(val)
	}
}

func (svcb *RDataSVCB) packIpv6hint(msg *Message, listValue []string) {
	msg.packet = binary.BigEndian.AppendUint16(msg.packet,
		uint16(svcbKeyIDIpv6hint))

	var total = 16 * len(listValue)
	msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(total))

	var val string

	for _, val = range listValue {
		msg.packIPv6(val)
	}
}

func (svcb *RDataSVCB) packGenericValue(keyid int, msg *Message, listValue []string) {
	var val = strings.Join(listValue, `,`)

	msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(keyid))
	msg.packet = binary.BigEndian.AppendUint16(msg.packet, uint16(len(val)))
	msg.packet = append(msg.packet, []byte(val)...)
}

// parseParams parse parameters from zone file.
//
//	SvcParam      = SvcParamKey [ "=" SvcParamValue ]
//	SvcParamKey   = 1*63(ASCII_LETTER / ASCII_DIGIT / "-")
//	SvcParamValue = STRING
//	WSP           = " " / "\t"
//	ASCII_LETTER  = ; a-z
//	ASCII_DIGIT   = ; 0-9
func (svcb *RDataSVCB) parseParams(zp *zoneParser) (err error) {
	var (
		logp = `parseParams`

		tok []byte
	)

	zp.parser.AddDelimiters([]byte{'='})
	defer zp.parser.RemoveDelimiters([]byte{'='})

	for {
		tok, err = svcb.getParamKey(zp)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
		if len(tok) == 0 {
			break
		}

		var key = strings.ToLower(string(tok))
		if key == svcbKeyNameNoDefaultALPN {
			if zp.delim == '=' {
				return fmt.Errorf(`%s: key %q must not have value`, logp, key)
			}
			err = svcb.AddParam(key, nil)
			if err != nil {
				return fmt.Errorf(`%s: %w`, logp, err)
			}
			continue
		}

		tok, err = svcb.getParamValue(zp)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
		if len(tok) == 0 {
			return fmt.Errorf(`%s: missing value for key %q`, logp, key)
		}

		tok, err = zp.decodeString(tok)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}

		var listValue []string

		listValue, err = svcbSplitRawValue(tok)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}

		err = svcb.AddParam(key, listValue)
		if err != nil {
			return fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	return nil
}

func (svcb *RDataSVCB) unpack(packet, rdata []byte, start uint) (err error) {
	svcb.Priority = binary.BigEndian.Uint16(rdata)
	rdata = rdata[2:]
	start += 2

	var end uint
	svcb.TargetName, end, err = unpackDomainName(packet, start)
	if err != nil {
		return err
	}
	start = end - start
	rdata = rdata[start:]
	err = svcb.unpackParams(rdata)
	if err != nil {
		return err
	}

	return nil
}

func (svcb *RDataSVCB) unpackParams(packet []byte) (err error) {
	var keyid uint16

	for len(packet) > 0 {
		keyid = binary.BigEndian.Uint16(packet)
		packet = packet[2:]

		switch int(keyid) {
		case svcbKeyIDMandatory:
			packet, err = svcb.unpackParamMandatory(packet)

		case svcbKeyIDALPN:
			packet, err = svcb.unpackParamALPN(packet)

		case svcbKeyIDNoDefaultALPN:
			svcb.Params[int(keyid)] = nil

		case svcbKeyIDPort:
			packet, err = svcb.unpackParamPort(packet)

		case svcbKeyIDIpv4hint:
			packet, err = svcb.unpackParamIpv4hint(packet)

		case svcbKeyIDEch:
			// NO-OP.

		case svcbKeyIDIpv6hint:
			packet, err = svcb.unpackParamIpv6hint(packet)

		default:
			packet, err = svcb.unpackParamGeneric(packet, int(keyid))
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (svcb *RDataSVCB) unpackParamMandatory(packet []byte) ([]byte, error) {
	if len(packet) < 2 {
		return packet, errors.New(`missing mandatory key value`)
	}

	var size = binary.BigEndian.Uint16(packet)
	if size <= 0 {
		return packet, fmt.Errorf(`invalid mandatory length %d`, size)
	}
	packet = packet[2:]

	var (
		n = int(size) / 2

		listValue []string
	)
	for n > 0 {
		if len(packet) < 2 {
			return packet, fmt.Errorf(`missing mandatory value on index %d`, len(listValue))
		}

		var keyid = binary.BigEndian.Uint16(packet)
		packet = packet[2:]

		var keyName = svcbKeyName(int(keyid))
		listValue = append(listValue, keyName)
		n--
	}
	svcb.Params[svcbKeyIDMandatory] = listValue

	return packet, nil
}

func (svcb *RDataSVCB) unpackParamALPN(packet []byte) ([]byte, error) {
	var logp = `unpackParamALPN`

	if len(packet) < 2 {
		return packet, fmt.Errorf(`%s: missing length and value`, logp)
	}

	var total = int(binary.BigEndian.Uint16(packet))
	if total <= 0 {
		return packet, fmt.Errorf(`%s: invalid length %d`, logp, total)
	}
	packet = packet[2:]

	var listValue []string

	for total > 0 {
		if len(packet) == 0 {
			return packet, fmt.Errorf(`%s: missing value on index %d`, logp, len(listValue))
		}

		var n = int(packet[0])
		packet = packet[1:]
		total--

		if len(packet) < int(total) {
			return packet, fmt.Errorf(`%s: mismatch value length, want %d got %d`, logp, n, len(packet))
		}

		var keyName = string(packet[:n])
		packet = packet[n:]

		listValue = append(listValue, keyName)
		total -= n
	}

	svcb.Params[svcbKeyIDALPN] = listValue

	return packet, nil
}

func (svcb *RDataSVCB) unpackParamPort(packet []byte) ([]byte, error) {
	var logp = `unpackParamPort`

	if len(packet) < 4 {
		return packet, fmt.Errorf(`%s: missing value`, logp)
	}

	var u16 = binary.BigEndian.Uint16(packet)
	if u16 <= 0 {
		return packet, fmt.Errorf(`%s: invalid length %d`, logp, u16)
	}
	packet = packet[2:]

	u16 = binary.BigEndian.Uint16(packet)
	if u16 <= 0 {
		return packet, fmt.Errorf(`%s: invalid port %d`, logp, u16)
	}
	packet = packet[2:]

	var portv = strconv.FormatUint(uint64(u16), 10)
	svcb.Params[svcbKeyIDPort] = []string{portv}

	return packet, nil
}

func (svcb *RDataSVCB) unpackParamIpv4hint(packet []byte) ([]byte, error) {
	var logp = `unpackParamIpv4hint`

	if len(packet) < 2 {
		return packet, fmt.Errorf(`%s: missing value`, logp)
	}

	var size = int(binary.BigEndian.Uint16(packet))
	if size <= 0 {
		return nil, fmt.Errorf(`%s: invalid length %d`, logp, size)
	}
	packet = packet[2:]

	var (
		n         = size / 4
		listValue []string
	)
	for n > 0 {
		if len(packet) < 4 {
			return packet, fmt.Errorf(`%s: missing value on index %d`, logp, len(listValue))
		}
		var ip = net.IP(packet[0:4])
		packet = packet[4:]
		listValue = append(listValue, ip.String())
		n--
	}

	svcb.Params[svcbKeyIDIpv4hint] = listValue
	return packet, nil
}

func (svcb *RDataSVCB) unpackParamIpv6hint(packet []byte) ([]byte, error) {
	var logp = `unpackParamIpv6hint`

	if len(packet) < 2 {
		return packet, fmt.Errorf(`%s: missing value`, logp)
	}

	var size = int(binary.BigEndian.Uint16(packet))
	if size <= 0 {
		return nil, fmt.Errorf(`%s: invalid length %d`, logp, size)
	}
	packet = packet[2:]

	var (
		n         = size / 16
		listValue []string
	)
	for n > 0 {
		if len(packet) < 16 {
			return packet, fmt.Errorf(`%s: missing value on index %d`, logp, len(listValue))
		}
		var ip = net.IP(packet[:16])
		packet = packet[16:]
		listValue = append(listValue, ip.String())
		n--
	}

	svcb.Params[svcbKeyIDIpv6hint] = listValue

	return packet, nil
}

func (svcb *RDataSVCB) unpackParamGeneric(packet []byte, keyid int) ([]byte, error) {
	var logp = `unpackParamGeneric`

	if len(packet) < 2 {
		return nil, fmt.Errorf(`%s: missing parameter value`, logp)
	}

	var size = int(binary.BigEndian.Uint16(packet))
	if size <= 0 {
		return packet, fmt.Errorf(`%s: invalid length %d`, logp, size)
	}
	packet = packet[2:]

	if len(packet) < size {
		return packet, fmt.Errorf(`%s: mismatch value length, want %d got %d`,
			logp, size, len(packet))
	}

	var val = string(packet[:size])
	packet = packet[size:]

	svcb.Params[keyid] = []string{val}

	return packet, nil
}

// validate the mandatory parameter.
// Each key in mandatory value should only defined once.
func (svcb *RDataSVCB) validate() (err error) {
	var (
		listValue []string
		ok        bool
	)
	listValue, ok = svcb.Params[svcbKeyIDMandatory]
	if !ok {
		return nil
	}

	var (
		key   string
		keyid int
	)
	for _, key = range listValue {
		keyid = svcbKeyID(key)
		if keyid < 0 {
			return fmt.Errorf(`invalid key %q`, key)
		}
		if keyid == svcbKeyIDMandatory {
			return errors.New(`mandatory key must not be included in the "mandatory" value`)
		}
		_, ok = svcb.Params[keyid]
		if !ok {
			return fmt.Errorf(`missing mandatory key %q`, key)
		}
	}
	return nil
}

func (svcb *RDataSVCB) writeParam(out io.Writer, keyid int) {
	var (
		listValue = svcb.Params[keyid]

		sb        strings.Builder
		val       string
		x         int
		isEscaped bool
		isQuoted  bool
	)
	for x, val = range listValue {
		if x > 0 {
			sb.WriteByte(',')
		}
		val, isEscaped = svcbEncodeValue(val)
		if isEscaped {
			isQuoted = true
		}
		sb.WriteString(val)
	}

	var keyName = svcbKeyName(keyid)
	if isQuoted {
		fmt.Fprintf(out, `%s="%s"`, keyName, sb.String())
	} else {
		fmt.Fprintf(out, `%s=%s`, keyName, sb.String())
	}
}

// svcbEncodeValue encode the parameter value.
// A comma ',', backslash '\', or double quote '"' will be escaped using
// backslash.
// Non-printable character will be encoded as escaped octal, "\XXX", where
// XXX is the octal value of character.
func svcbEncodeValue(in string) (out string, escaped bool) {
	var (
		rawin = []byte(in)

		sb strings.Builder
		c  byte
	)
	for _, c = range rawin {
		switch {
		case c == ',', c == '\\', c == '"':
			sb.WriteString(`\\\`)
			sb.WriteByte(c)
			escaped = true
			continue

		case c == '!',
			c >= 0x23 && c <= 0x27,
			c >= 0x2A && c <= 0x3A,
			c >= 0x3C && c <= 0x5B,
			c >= 0x5D && c <= 0x7E:
			sb.WriteByte(c)

		default:
			// Write byte as escaped decimal "\XXX".
			sb.WriteString(`\` + strconv.FormatUint(uint64(c), 10))
			escaped = true
		}

	}
	return sb.String(), escaped
}

// svcbSplitRawValue split raw SVCB parameter value by comma ','.
// A comma can be escaped using backslash '\'.
// A backslash also can be escaped using backslash.
// Other than that, no escaped sequence are allowed.
func svcbSplitRawValue(raw []byte) (listValue []string, err error) {
	var (
		val   []byte
		x     int
		isEsc bool
	)
	for ; x < len(raw); x++ {
		if isEsc {
			switch raw[x] {
			case '\\':
				val = append(val, '\\')
			case ',':
				val = append(val, ',')
			default:
				return nil, fmt.Errorf(`invalid escaped character %q`, raw[x])
			}
			isEsc = false
			continue
		}
		if raw[x] == '\\' {
			isEsc = true
			continue
		}
		if raw[x] == ',' {
			listValue = append(listValue, string(val))
			val = nil
			continue
		}
		val = append(val, raw[x])
	}
	if len(val) != 0 {
		listValue = append(listValue, string(val))
	}
	return listValue, nil
}

// svcbKeyID return the key ID based on string value.
// It will return -1 if key is invalid.
func svcbKeyID(key string) int {
	switch key {
	case svcbKeyNameMandatory:
		return svcbKeyIDMandatory
	case svcbKeyNameALPN:
		return svcbKeyIDALPN
	case svcbKeyNameNoDefaultALPN:
		return svcbKeyIDNoDefaultALPN
	case svcbKeyNamePort:
		return svcbKeyIDPort
	case svcbKeyNameIpv4hint:
		return svcbKeyIDIpv4hint
	case svcbKeyNameEch:
		return svcbKeyIDEch
	case svcbKeyNameIpv6hint:
		return svcbKeyIDIpv6hint
	}
	if !strings.HasPrefix(key, `key`) {
		return -1
	}

	key = strings.TrimPrefix(key, `key`)

	var (
		keyid int64
		err   error
	)

	keyid, err = strconv.ParseInt(key, 10, 16)
	if err != nil {
		return -1
	}
	if keyid < 0 || keyid > math.MaxUint16 {
		return -1
	}
	return int(keyid)
}

func svcbKeyName(keyid int) string {
	switch keyid {
	case svcbKeyIDMandatory:
		return svcbKeyNameMandatory
	case svcbKeyIDALPN:
		return svcbKeyNameALPN
	case svcbKeyIDNoDefaultALPN:
		return svcbKeyNameNoDefaultALPN
	case svcbKeyIDPort:
		return svcbKeyNamePort
	case svcbKeyIDIpv4hint:
		return svcbKeyNameIpv4hint
	case svcbKeyIDEch:
		return svcbKeyNameEch
	case svcbKeyIDIpv6hint:
		return svcbKeyNameIpv6hint
	}
	return fmt.Sprintf(`key%d`, keyid)
}
