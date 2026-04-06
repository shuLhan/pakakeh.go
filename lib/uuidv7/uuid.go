// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2026 M. Shulhan <ms@kilabit.info>

// Package uuidv7 implements UUID version 7 as defined in RFC 9562.
package uuidv7

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

const hexTable = `0123456789ABCDEF`
const variantMask uint64 = 0x8000_0000_0000_0000

// Now defines the function variable that return the current time in UTC.
// This identifier is exported to simplify testing.
var Now = func() time.Time {
	return time.Now().UTC()
}

// Rand defines the function variable that return a random uint64.
// This identifier is exported to simplify testing.
var Rand = func() (v uint64) {
	b := make([]byte, 8)
	rand.Read(b)
	v, _ = binary.Uvarint(b)
	return v
}

// UUID defines the container for UUID version 7 that satisfy the
// [database/sql], [encoding/gob], and [encoding/json].
type UUID struct {
	// The high 8 octets store the unix timestamp in milliseconds (6
	// octets), the version (4 bits), and clock precision (12 bits).
	high uint64

	// The low 8 octets store the variant (2 bits) and random number (62
	// bits).
	low uint64
}

// Generate generates new UUID version 7.
func Generate() (id UUID) {
	now := Now()
	generate(&now, &id)
	return id
}

func generate(t *time.Time, id *UUID) {
	unixns := uint64(t.UnixNano())           // 1773066002123456700
	millis := uint64(unixns / 1e6)           // 1773066002123
	id.high = millis << 16                   // 0x019c_d2f8_1ecb_0000
	id.high |= 0x7000                        // 0x019c_d2f8_1ecb_7000
	prec := unixns - (millis * 1e6)          // 456700
	clockPrec := uint64((prec * 4096) / 1e6) // 1870
	id.high |= clockPrec                     // 0x019c_d2f8_1ecb_774e
	rand := Rand()                           // 0xffff_ffff_ffff_ffff
	rand >>= 2                               // 0x3fff_ffff_ffff_ffff
	id.low = variantMask | rand              // 0xbfff_ffff_ffff_ffff
}

// Parse parses the UUID formated string and return it as non-zero id.
// If the string s is invalid UUID it will return zero id.
func Parse(s string) (id UUID) {
	id.UnmarshalText([]byte(s))
	return id
}

// Bytes returns id as a stream of binary.
func (id UUID) Bytes() (data []byte) {
	data = make([]byte, 0, binary.MaxVarintLen64*2)
	data = binary.BigEndian.AppendUint64(data, id.high)
	data = binary.BigEndian.AppendUint64(data, id.low)
	return data
}

// Equal returns nil if id and v are equals.
// This method implements [git.sr.ht/~shulhan/pakakeh.go/lib/reflect.Equaler]
// interface.
func (id *UUID) Equal(v any) (err error) {
	ptr, ok := v.(*UUID)
	if !ok {
		other, ok := v.(UUID)
		if !ok {
			return fmt.Errorf(`uuidv7: Equal: want type %T, got %T`, &id, v)
		}
		ptr = &other
	}
	if id.high != ptr.high {
		return fmt.Errorf(`uuidv7: not equal, want %s, got %s`, id.String(), ptr.String())
	}
	if id.low != ptr.low {
		return fmt.Errorf(`uuidv7: not equal, want %s, got %s`, id.String(), ptr.String())
	}
	return nil
}

// IsEqual returns true if id equal with other.
func (id UUID) IsEqual(other UUID) bool {
	return id.high == other.high && id.low == other.low
}

// IsZero returns true if all bits is zero.
func (id UUID) IsZero() bool {
	return id.high == 0 && id.low == 0
}

// MarshalBinary encodes the id to binary for [encoding/gob].
func (id UUID) MarshalBinary() (data []byte, err error) {
	data = id.Bytes()
	return data, nil
}

// MarshalText encodes the id to JSON for [encoding/json].
func (id UUID) MarshalText() (data []byte, err error) {
	v := id.String()
	return []byte(v), nil
}

// UnmarshalBinary decodes the data into id for [encoding/gob].
func (id *UUID) UnmarshalBinary(data []byte) (err error) {
	if len(data) != 16 {
		return fmt.Errorf(`uuidv7: invalid length %d`, len(data))
	}
	ver := data[6] >> 4
	if ver != 7 {
		return fmt.Errorf(`uuidv7: invalid version %d`, ver)
	}
	id.high = binary.BigEndian.Uint64(data[:8])
	id.low = binary.BigEndian.Uint64(data[8:])
	return nil
}

// UnmarshalText decodes the JSON data into id for [encoding/json].
func (id *UUID) UnmarshalText(data []byte) (err error) {
	src := make([]byte, 0, 32)
	for _, c := range data {
		if c == '-' {
			continue
		}
		src = append(src, c)
	}
	l := len(src)
	if l != 32 {
		return fmt.Errorf(`uuidv7: invalid length %d`, l)
	}
	if src[12] != '7' {
		return fmt.Errorf(`uuidv7: invalid version %c`, data[12])
	}

	l = hex.DecodedLen(len(src[:16]))
	dst := make([]byte, l)
	_, err = hex.Decode(dst, src[:16])
	if err != nil {
		return fmt.Errorf(`uuidv7: %w`, err)
	}
	high := binary.BigEndian.Uint64(dst)

	_, err = hex.Decode(dst, src[16:])
	if err != nil {
		return fmt.Errorf(`uuidv7: %w`, err)
	}
	id.high = high
	id.low = binary.BigEndian.Uint64(dst)
	return nil
}

// Scan scans the raw database value into id.
// This method implement [database/sql.Scanner] interface.
// Column with NULL value will returns no error set it to zero UUID.
func (id *UUID) Scan(src any) (err error) {
	switch v := src.(type) {
	case []byte:
		err = id.UnmarshalText(v)
		if err != nil {
			return err
		}
	case nil:
		id.high = 0
		id.low = 0
		return nil
	default:
		return fmt.Errorf(`uuidv7: Scan: invalid type %T`, src)
	}
	return nil
}

// String returns the human readable format of UUID in the form of
// "0xxxxxxx-xxxx-7xxx-8xxx-xxxxxxxxxxxx".
func (id UUID) String() string {
	buf := make([]byte, 36)
	sw := 56
	x := 0
	for sw >= 0 {
		b := byte(id.high >> sw)
		sw -= 8
		buf[x] = hexTable[b>>4]
		buf[x+1] = hexTable[b&0x0F]
		x += 2
		if x == 8 || x == 13 {
			buf[x] = '-'
			x++
		}
	}
	buf[x] = '-'
	x++
	sw = 56
	for sw >= 0 {
		b := byte(id.low >> sw)
		sw -= 8
		buf[x] = hexTable[b>>4]
		buf[x+1] = hexTable[b&0x0F]
		x += 2
		if x == 23 {
			buf[x] = '-'
			x++
		}
	}
	return string(buf)
}

// Time returns the Unix epoch timestamp in milliseconds stored in the UUID.
func (id *UUID) Time() time.Time {
	millis := id.high >> 16
	t := time.UnixMilli(int64(millis))
	return t
}

// Value returns the value for sending it to the database.
// This method implements the [driver.Valuer] interface.
func (id UUID) Value() (v driver.Value, err error) {
	v = id.String()
	return v, nil
}
