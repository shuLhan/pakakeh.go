// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"unsafe"
)

// storage provides the backing storage for BigEndianBuffer.
type storage struct {
	bin     []byte
	off     int64
	endness Endian
}

func (stor *storage) encode(data any) (n int, err error) {
	switch val := data.(type) {
	case bool:
		if val {
			stor.bin = append(stor.bin[:stor.off], 1)
		} else {
			stor.bin = append(stor.bin[:stor.off], 0)
		}
		stor.off++
		return 1, nil
	case *bool:
		if *val {
			stor.bin = append(stor.bin[:stor.off], 1)
		} else {
			stor.bin = append(stor.bin[:stor.off], 0)
		}
		stor.off++
		return 1, nil

	case int8:
		stor.bin = append(stor.bin[:stor.off], byte(val))
		stor.off++
		return 1, nil
	case *int8:
		stor.bin = append(stor.bin[:stor.off], byte(*val))
		stor.off++
		return 1, nil
	case uint8:
		stor.bin = append(stor.bin[:stor.off], val)
		stor.off++
		return 1, nil
	case *uint8:
		stor.bin = append(stor.bin[:stor.off], *val)
		stor.off++
		return 1, nil

	case int16:
		n = stor.encodeUint16(uint16(val))
		return n, nil
	case *int16:
		n = stor.encodeUint16(uint16(*val))
		return n, nil
	case uint16:
		n = stor.encodeUint16(val)
		return n, nil
	case *uint16:
		n = stor.encodeUint16(*val)
		return n, nil

	case int32:
		n = stor.encodeUint32(uint32(val))
		return n, nil
	case *int32:
		n = stor.encodeUint32(uint32(*val))
		return n, nil
	case uint32:
		n = stor.encodeUint32(val)
		return n, nil
	case *uint32:
		n = stor.encodeUint32(*val)
		return n, nil

	case int64:
		n = stor.encodeUint64(uint64(val))
		return n, nil
	case *int64:
		n = stor.encodeUint64(uint64(*val))
		return n, nil
	case uint64:
		n = stor.encodeUint64(val)
		return n, nil

	case *uint64:
		n = stor.encodeUint64(*val)
		return n, nil

	case int:
		if unsafe.Sizeof(val) == 4 {
			n = stor.encodeUint32(uint32(val))
		} else {
			n = stor.encodeUint64(uint64(val))
		}
		return n, nil
	case *int:
		if unsafe.Sizeof(*val) == 4 {
			n = stor.encodeUint32(uint32(*val))
		} else {
			n = stor.encodeUint64(uint64(*val))
		}
		return n, nil
	case uint:
		if unsafe.Sizeof(val) == 4 {
			n = stor.encodeUint32(uint32(val))
		} else {
			n = stor.encodeUint64(uint64(val))
		}
		return n, nil
	case *uint:
		if unsafe.Sizeof(*val) == 4 {
			n = stor.encodeUint32(uint32(*val))
		} else {
			n = stor.encodeUint64(uint64(*val))
		}
		return n, nil

	case float32:
		n = stor.encodeUint32(math.Float32bits(val))
		return n, nil

	case *float32:
		n = stor.encodeUint32(math.Float32bits(*val))
		return n, nil

	case float64:
		n = stor.encodeUint64(math.Float64bits(val))
		return n, nil

	case *float64:
		n = stor.encodeUint64(math.Float64bits(*val))
		return n, nil

	case complex64:
		n = stor.encodeComplex64(val)
		return n, nil

	case *complex64:
		n = stor.encodeComplex64(*val)
		return n, nil

	case complex128:
		n = stor.encodeComplex128(val)
		return n, nil
	case *complex128:
		n = stor.encodeComplex128(*val)
		return n, nil

	case string:
		n = stor.encodeSliceByte([]byte(val))
		return n, nil

	case *string:
		n = stor.encodeSliceByte([]byte(*val))
		return n, nil

	case []bool:
		n = 4 + len(val)
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, v := range val {
			if v {
				tmp[x] = 1
			} else {
				tmp[x] = 0
			}
		}
		stor.off += int64(n)
		return n, nil

	case []int8:
		n = 4 + len(val)
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, v := range val {
			tmp[x] = uint8(v)
		}
		stor.off += int64(n)
		return n, nil
	case []uint8:
		n = stor.encodeSliceByte(val)
		return n, nil

	case []int16:
		n = 4 + (2 * len(val))
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, v := range val {
			SetUint16(stor.endness, tmp[x*2:], uint16(v))
		}
		stor.off += int64(n)
		return n, nil

	case []uint16:
		n = 4 + (2 * len(val))
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, v := range val {
			SetUint16(stor.endness, tmp[x*2:], v)
		}
		stor.off += int64(n)
		return n, nil

	case []int32:
		n = 4 + (4 * len(val))
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, v := range val {
			SetUint32(stor.endness, tmp[x*4:], uint32(v))
		}
		stor.off += int64(n)
		return n, nil

	case []uint32:
		n = 4 + (4 * len(val))
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, v := range val {
			SetUint32(stor.endness, tmp[x*4:], v)
		}
		stor.off += int64(n)
		return n, nil

	case []int64:
		n = 4 + (8 * len(val))
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, v := range val {
			SetUint64(stor.endness, tmp[x*8:], uint64(v))
		}
		stor.off += int64(n)
		return n, nil

	case []uint64:
		n = 4 + (8 * len(val))
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, v := range val {
			SetUint64(stor.endness, tmp[x*8:], v)
		}
		stor.off += int64(n)
		return n, nil

	case []int:
		if unsafe.Sizeof(int(1)) == 4 {
			n = 4 + (4 * len(val))
			stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
			SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
			var tmp = stor.bin[stor.off+4:]
			for x, v := range val {
				SetUint32(stor.endness, tmp[x*4:], uint32(v))
			}
		} else {
			n = 4 + (8 * len(val))
			stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
			SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
			var tmp = stor.bin[stor.off+4:]
			for x, v := range val {
				SetUint64(stor.endness, tmp[x*8:], uint64(v))
			}
		}
		stor.off += int64(n)
		return n, nil

	case []uint:
		if unsafe.Sizeof(uint(1)) == 4 {
			n = 4 + (4 * len(val))
			stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
			SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
			var tmp = stor.bin[stor.off+4:]
			for x, v := range val {
				SetUint32(stor.endness, tmp[x*4:], uint32(v))
			}
		} else {
			n = 4 + (8 * len(val))
			stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
			SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
			var tmp = stor.bin[stor.off+4:]
			for x, v := range val {
				SetUint64(stor.endness, tmp[x*8:], uint64(v))
			}
		}
		stor.off += int64(n)
		return n, nil

	case []float32:
		n = 4 + (4 * len(val))
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, f32 := range val {
			var ui32 = math.Float32bits(f32)
			SetUint32(stor.endness, tmp[x*4:], ui32)
		}
		stor.off += int64(n)
		return n, nil

	case []float64:
		n = 4 + (8 * len(val))
		stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
		SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
		var tmp = stor.bin[stor.off+4:]
		for x, f64 := range val {
			var ui64 = math.Float64bits(f64)
			SetUint64(stor.endness, tmp[x*8:], ui64)
		}
		stor.off += int64(n)
		return n, nil

	default:
		var refval = reflect.ValueOf(data)
		var refkind = refval.Kind()

		if refkind == reflect.Pointer {
			refval = reflect.Indirect(refval)
			refkind = refval.Kind()
		}
		if isKindIgnored(refkind) {
			return 0, nil
		}

		// We have eliminated most of unused types, and left with Array,
		// Slice, String, Struct, and basic types.

		switch refkind {
		case reflect.Array:
			n, err = stor.encodeSlice(refval)
		case reflect.Slice:
			n, err = stor.encodeSlice(refval)
		case reflect.Struct:
			n, err = stor.encodeStruct(refval)
		default:
			return 0, fmt.Errorf(`unsupported type %T`, data)
		}
	}
	return n, err
}

func (stor *storage) encodeComplex64(val complex64) (n int) {
	n = 8
	stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
	SetUint32(stor.endness, stor.bin[stor.off:],
		math.Float32bits(float32(real(val))))
	SetUint32(stor.endness, stor.bin[stor.off+4:],
		math.Float32bits(float32(imag(val))))
	stor.off += int64(n)
	return n
}

func (stor *storage) encodeComplex128(val complex128) (n int) {
	n = 16
	stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
	SetUint64(stor.endness, stor.bin[stor.off:],
		math.Float64bits(float64(real(val))))
	SetUint64(stor.endness, stor.bin[stor.off+8:],
		math.Float64bits(float64(imag(val))))
	stor.off += int64(n)
	return n
}

func (stor *storage) encodeSlice(refval reflect.Value) (
	total int, err error,
) {
	var elkind = refval.Type().Elem().Kind()
	if isKindIgnored(elkind) {
		// The element of array or slice is not supported.
		return 0, fmt.Errorf(`unsupported type %T`, refval.Interface())
	}

	var (
		size = refval.Len()
		n    int
	)
	stor.encodeUint32(uint32(size))
	total = 4
	for x := range size {
		var el reflect.Value = refval.Index(x)
		n, err = stor.encode(el.Interface())
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func (stor *storage) encodeStruct(refval reflect.Value) (
	total int, err error,
) {
	var visibleFields = reflect.VisibleFields(refval.Type())
	var n int
	for _, sfield := range visibleFields {
		if !sfield.IsExported() {
			continue
		}
		var fval = refval.FieldByIndex(sfield.Index)
		if fval.Kind() == reflect.Pointer {
			// Field with nil value will be written only as single
			// byte 0.
			var fflag byte
			if fval.IsNil() {
				fflag = 0
			} else {
				fflag = 1
			}
			n, _ = stor.encode(fflag)
			total += n
			if fflag == 0 {
				continue
			}
		}
		n, err = stor.encode(fval.Interface())
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func (stor *storage) encodeSliceByte(val []byte) (n int) {
	n = 4 + len(val)
	stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
	SetUint32(stor.endness, stor.bin[stor.off:], uint32(len(val)))
	copy(stor.bin[stor.off+4:], val)
	stor.off += int64(n)
	return n
}

func (stor *storage) encodeUint16(val uint16) (n int) {
	n = 2
	stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
	SetUint16(stor.endness, stor.bin[stor.off:], val)
	stor.off += int64(n)
	return n
}

func (stor *storage) encodeUint32(val uint32) (n int) {
	n = 4
	stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
	SetUint32(stor.endness, stor.bin[stor.off:], val)
	stor.off += int64(n)
	return n
}

func (stor *storage) encodeUint64(val uint64) (n int) {
	n = 8
	stor.bin = slices.Grow(stor.bin[:stor.off], n)[:int(stor.off)+n]
	SetUint64(stor.endness, stor.bin[stor.off:], val)
	stor.off += int64(n)
	return n
}
