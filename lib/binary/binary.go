// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package binary complement the standard [binary] package.
package binary

import (
	"fmt"
	"math"
	"reflect"
	"time"
)

var timeNow = func() time.Time {
	return time.Now().UTC()
}

// Endian define the byte order to convert data to binary and vice versa.
type Endian = byte

const (
	// EndianBig define big-endian order, where bytes stored in
	// left-to-right.
	// For example int16=32767 {0x7F 0xFF} will be stored as is.
	EndianBig Endian = 1

	// EndianLittle define little-endian order, where byte stored in
	// right-to-left.
	// For example int16=32767 {0x7F 0xFF} will be stored as {0xFF 0x7F}
	EndianLitte = 2
)

// DecodeString convert binary to string.
func DecodeString(endness Endian, bin []byte) (val string) {
	var length = int(DecodeUint32(endness, bin))
	val = string(bin[4 : 4+length])
	return val
}

// DecodeUint16 convert binary to uint16.
func DecodeUint16(endness Endian, bin []byte) (val uint16) {
	if endness == EndianBig {
		val = uint16(bin[1]) | uint16(bin[0])<<8
	} else {
		val = uint16(bin[0]) | uint16(bin[1])<<8
	}
	return val
}

// DecodeUint32 convert binary to uint32.
func DecodeUint32(endness Endian, bin []byte) (val uint32) {
	if endness == EndianBig {
		val = uint32(bin[3]) |
			uint32(bin[2])<<8 |
			uint32(bin[1])<<16 |
			uint32(bin[0])<<24
	} else {
		val = uint32(bin[0]) |
			uint32(bin[1])<<8 |
			uint32(bin[2])<<16 |
			uint32(bin[3])<<24
	}
	return val
}

// DecodeUint64 convert binary to uint64.
func DecodeUint64(endness Endian, bin []byte) (val uint64) {
	if endness == EndianBig {
		val = uint64(bin[7]) |
			uint64(bin[6])<<8 |
			uint64(bin[5])<<16 |
			uint64(bin[4])<<24 |
			uint64(bin[3])<<32 |
			uint64(bin[2])<<40 |
			uint64(bin[1])<<48 |
			uint64(bin[0])<<56
	} else {
		val = uint64(bin[0]) |
			uint64(bin[1])<<8 |
			uint64(bin[2])<<16 |
			uint64(bin[3])<<24 |
			uint64(bin[4])<<32 |
			uint64(bin[5])<<40 |
			uint64(bin[6])<<48 |
			uint64(bin[7])<<56
	}
	return val
}

// EncodeString convert string to binary as slice of byte.
// String is encoded by writing the length of string (number of bytes) and
// then followed by content of string in bytes.
//
// The endian-ness does not affect on how to write the content of string.
func EncodeString(endness Endian, val string) (bin []byte) {
	var length = len(val)
	var binsize = EncodeUint32(endness, uint32(length))
	bin = make([]byte, 0, 4+length)
	bin = append(bin, binsize...)
	bin = append(bin, []byte(val)...)
	return bin
}

// EncodeUint16 convert uint16 to binary as slice of byte.
func EncodeUint16(endness Endian, val uint16) (bin []byte) {
	bin = make([]byte, 2)
	if endness == EndianBig {
		bin[0] = byte(val >> 8)
		bin[1] = byte(val)
	} else {
		bin[0] = byte(val)
		bin[1] = byte(val >> 8)
	}
	return bin
}

// EncodeUint32 convert uint32 to binary as slice of byte.
func EncodeUint32(endness Endian, val uint32) (bin []byte) {
	bin = make([]byte, 4)
	if endness == EndianBig {
		bin[0] = byte(val >> 24)
		bin[1] = byte(val >> 16)
		bin[2] = byte(val >> 8)
		bin[3] = byte(val)
	} else {
		bin[0] = byte(val)
		bin[1] = byte(val >> 8)
		bin[2] = byte(val >> 16)
		bin[3] = byte(val >> 24)
	}
	return bin
}

// EncodeUint64 convert uint64 to binary as slice of byte.
func EncodeUint64(endness Endian, val uint64) (bin []byte) {
	bin = make([]byte, 8)
	if endness == EndianBig {
		bin[0] = byte(val >> 56)
		bin[1] = byte(val >> 48)
		bin[2] = byte(val >> 40)
		bin[3] = byte(val >> 32)
		bin[4] = byte(val >> 24)
		bin[5] = byte(val >> 16)
		bin[6] = byte(val >> 8)
		bin[7] = byte(val)
	} else {
		bin[0] = byte(val)
		bin[1] = byte(val >> 8)
		bin[2] = byte(val >> 16)
		bin[3] = byte(val >> 24)
		bin[4] = byte(val >> 32)
		bin[5] = byte(val >> 40)
		bin[6] = byte(val >> 48)
		bin[7] = byte(val >> 56)
	}
	return bin
}

// SetUint16 write an uint16 value val into bin.
// This function assume the bin has enought storage, otherwise it would be
// panic.
func SetUint16(endness Endian, bin []byte, val uint16) {
	if endness == EndianBig {
		bin[0] = byte(val >> 8)
		bin[1] = byte(val)
	} else {
		bin[0] = byte(val)
		bin[1] = byte(val >> 8)
	}
}

// SetUint32 write an uint32 value val into bin.
// This function assume the bin has enought storage, otherwise it would be
// panic.
func SetUint32(endness Endian, bin []byte, val uint32) {
	if endness == EndianBig {
		bin[0] = byte(val >> 24)
		bin[1] = byte(val >> 16)
		bin[2] = byte(val >> 8)
		bin[3] = byte(val)
	} else {
		bin[0] = byte(val)
		bin[1] = byte(val >> 8)
		bin[2] = byte(val >> 16)
		bin[3] = byte(val >> 24)
	}
}

// SetUint64 write an uint64 value val into bin.
// This function assume the bin has enought storage, otherwise it would be
// panic.
func SetUint64(endness Endian, bin []byte, val uint64) {
	if endness == EndianBig {
		bin[0] = byte(val >> 56)
		bin[1] = byte(val >> 48)
		bin[2] = byte(val >> 40)
		bin[3] = byte(val >> 32)
		bin[4] = byte(val >> 24)
		bin[5] = byte(val >> 16)
		bin[6] = byte(val >> 8)
		bin[7] = byte(val)
	} else {
		bin[0] = byte(val)
		bin[1] = byte(val >> 8)
		bin[2] = byte(val >> 16)
		bin[3] = byte(val >> 24)
		bin[4] = byte(val >> 32)
		bin[5] = byte(val >> 40)
		bin[6] = byte(val >> 48)
		bin[7] = byte(val >> 56)
	}
}

func decode(endness Endian, bin []byte, val any) (n int, err error) {
	var refval = reflect.ValueOf(val)
	refval = reflect.Indirect(refval)
	var refkind = refval.Kind()

	if isKindIgnored(refkind) {
		return 0, nil
	}
	if refkind == reflect.Array {
		return decodeArray(endness, bin, refval)
	}
	if refkind == reflect.Slice {
		return decodeSlice(endness, bin, refval)
	}
	if refkind == reflect.Struct {
		return decodeStruct(endness, bin, refval)
	}

	switch v := val.(type) {
	case *bool:
		var b = bin[0]
		if b == 1 {
			*v = true
		}
		n = 1

	case *int8:
		*v = int8(bin[0])
		n = 1
	case *uint8:
		*v = uint8(bin[0])
		n = 1

	case *int16:
		var ui16 = DecodeUint16(endness, bin[:2])
		*v = int16(ui16)
		n = 2
	case *uint16:
		var ui16 = DecodeUint16(endness, bin[:2])
		*v = ui16
		n = 2

	case *int32:
		var ui32 = DecodeUint32(endness, bin[:4])
		*v = int32(ui32)
		n = 4
	case *uint32:
		var ui32 = DecodeUint32(endness, bin[:4])
		*v = ui32
		n = 4

	case *int64:
		var ui64 = DecodeUint64(endness, bin[:8])
		*v = int64(ui64)
		n = 8
	case *uint64:
		*v = DecodeUint64(endness, bin[:8])
		n = 8

	case *int:
		if refval.Type().Size() == 4 {
			var ui32 = DecodeUint32(endness, bin[:4])
			*v = int(ui32)
			n = 4
		} else {
			var ui64 = DecodeUint64(endness, bin[:8])
			*v = int(ui64)
			n = 8
		}
	case *uint:
		if refval.Type().Size() == 4 {
			var ui32 = DecodeUint32(endness, bin[:4])
			*v = uint(ui32)
			n = 4
		} else {
			var ui64 = DecodeUint64(endness, bin[:8])
			*v = uint(ui64)
			n = 8
		}

	case *float32:
		var ui32 = DecodeUint32(endness, bin[:4])
		*v = math.Float32frombits(ui32)
		n = 4
	case *float64:
		var ui64 = DecodeUint64(endness, bin[:8])
		*v = math.Float64frombits(ui64)
		n = 8

	case *complex64:
		var ui32 uint32 = DecodeUint32(endness, bin[:4])
		var re float32 = math.Float32frombits(ui32)

		ui32 = DecodeUint32(endness, bin[4:8])
		var im float32 = math.Float32frombits(ui32)

		*v = complex(re, im)
		n = 8

	case *complex128:
		var ui64 uint64 = DecodeUint64(endness, bin[:8])
		var re float64 = math.Float64frombits(ui64)

		ui64 = DecodeUint64(endness, bin[8:16])
		var im float64 = math.Float64frombits(ui64)

		*v = complex(re, im)
		n = 16

	case *string:
		*v = DecodeString(endness, bin)
		n = 4 + len(*v)

	default:
		return 0, fmt.Errorf(`unsupported type %T`, v)
	}

	return n, nil
}

func decodeArray(endness Endian, bin []byte, refval reflect.Value) (
	n int, err error,
) {
	var reftype = refval.Type()
	var elkind = reftype.Elem().Kind()
	if isKindIgnored(elkind) {
		return 0, fmt.Errorf(`unsupported element type %T`,
			refval.Interface())
	}

	// Read and compare the length from parameter and stored.
	var reflen = refval.Len()
	var storedLen = int(DecodeUint32(endness, bin[:4]))

	if reflen != storedLen {
		return 0, fmt.Errorf(`expecting slice/array length %d, got %d`,
			storedLen, reflen)
	}

	var total = 4
	bin = bin[4:]

	for x := range storedLen {
		var elval = refval.Index(x)

		n, err = decode(endness, bin, elval.Addr().Interface())
		if err != nil {
			return total, err
		}

		bin = bin[n:]
		total += n
	}

	return total, nil
}

// decodeSlice read the slice values from bin.
func decodeSlice(endness Endian, bin []byte, refval reflect.Value) (
	n int, err error,
) {
	var logp = `decodeSlice`

	if !refval.CanAddr() {
		if refval.IsNil() {
			return 0, fmt.Errorf(`%s: expecting initialized slice, got nil`, logp)
		}
		return decodeArray(endness, bin, refval)
	}

	var eltype = refval.Type().Elem()
	var elkind = eltype.Kind()
	if isKindIgnored(elkind) {
		// The element of slice is not supported.
		return 0, nil
	}

	// Read the length.
	var sliceLen = int(DecodeUint32(endness, bin[:4]))
	var slice = reflect.MakeSlice(refval.Type(), 0, sliceLen)
	var total = 4
	bin = bin[4:]

	for range sliceLen {
		// Create new(T) to be filled.
		var elval = reflect.New(eltype)
		n, err = decode(endness, bin, elval.Interface())
		if err != nil {
			return total, err
		}

		// Convert *T to T back.
		elval = reflect.Indirect(elval)
		slice = reflect.Append(slice, elval)
		bin = bin[n:]
		total += n
	}

	// Finally, set the reflect value to the slice we created.
	refval.Set(slice)

	return total, nil
}

func decodeStruct(endness Endian, bin []byte, refval reflect.Value) (
	total int, err error,
) {
	var logp = `decodeStruct`
	var visibleFields = reflect.VisibleFields(refval.Type())
	var n int
	for _, sfield := range visibleFields {
		if !sfield.IsExported() {
			continue
		}
		var fval = refval.FieldByIndex(sfield.Index)
		var fkind = fval.Type().Kind()
		if fkind != reflect.Pointer {
			if !fval.CanAddr() {
				return total, fmt.Errorf(`%s: field %q is unsetabble`,
					logp, sfield.Name)
			}
			fval = fval.Addr()
		} else {
			var fflag byte
			n, _ = decode(endness, bin, &fflag)
			bin = bin[n:]
			total += n
			if fflag == 0 {
				// Skip pointer with nil.
				continue
			}
			var newval = reflect.New(fval.Type().Elem())
			fval.Set(newval)
		}
		n, err = decode(endness, bin, fval.Interface())
		if err != nil {
			return total, err
		}
		bin = bin[n:]
		total += n
	}
	return total, nil
}

func isKindIgnored(refkind reflect.Kind) bool {
	return refkind == reflect.Invalid || refkind == reflect.Uintptr ||
		refkind == reflect.Chan || refkind == reflect.Func ||
		refkind == reflect.Interface || refkind == reflect.Map
}
