// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"fmt"
	"io"
	"reflect"
)

// BigEndianBuffer provides backing storage for writing (most of) Go native
// types into binary in big-endian order.
// The zero value of BigEndianBuffer is ready to use.
//
// The following basic types are supported for Write and Read: bool, byte,
// int, float, complex, and string.
// The slice and array are also supported as long as the slice's element type
// is one of basic types.
//
// For the string, slice, and array, the Write operation write the dynamic
// length first as 4 bytes of uint32 and then followed by its content.
//
// For struct, each exported field is written in order.
// Field with type pointer have single byte flag to indicate whether the value
// is nil during write or not.
// If pointer field is nil, the flag will be set to 0, otherwise it will be
// set to 1.
type BigEndianBuffer struct {
	storage
}

// NewBigEndianBuffer creates and initializes a new [BigEndianBuffer] using
// bin as its initial contents.
// The new [BigEndianBuffer] takes control of the bin and the caller should
// not use bin after this call.
func NewBigEndianBuffer(bin []byte) *BigEndianBuffer {
	return &BigEndianBuffer{
		storage: storage{
			bin:     bin,
			off:     int64(len(bin)),
			endness: EndianBig,
		},
	}
}

// Bytes return the backing storage.
func (beb *BigEndianBuffer) Bytes() []byte {
	return beb.bin
}

// Read the binary value into data by its type.
//
// Like any function that needs to set any type, the instance of parameter
// data must be passed as pointer, including for slice, excepts as noted
// below.
//
// Slice with length can be read without passing it as pointer,
//
//	s := make([]int, 5)
//	Read(s)  // OK.
//
// But array with size cannot be read without passing it as pointer,
//
//	a := [5]int{}
//	Read(a)   // Fail, panic: reflect.Value.Addr of unaddressable value.
//	Read(&a)  // OK.
func (beb *BigEndianBuffer) Read(data any) (n int, err error) {
	var logp = `Read`
	var refval = reflect.ValueOf(data)
	var refkind = refval.Kind()
	if refkind == reflect.Pointer {
		refval = reflect.Indirect(refval)
		refkind = refval.Kind()
	}
	if isKindIgnored(refkind) {
		return 0, nil
	}
	if !refval.CanAddr() {
		if refkind == reflect.Slice && !refval.IsNil() {
			// Slice that has been created with make can be read,
			// even if the length is 0.
		} else {
			// The passed data must be pointer to variable.
			return 0, fmt.Errorf(`%s: expecting pointer to %T`,
				logp, data)
		}
	}

	beb.endness = EndianBig
	n, err = decode(EndianBig, beb.bin[beb.off:], data)
	if err != nil {
		return 0, fmt.Errorf(`%s: %w`, logp, err)
	}

	beb.off += int64(n)

	return n, nil
}

// Reset the internal storage, start from empty again.
func (beb *BigEndianBuffer) Reset() {
	beb.bin = beb.bin[:0]
	beb.off = 0
}

// Seek move the write and read position to the offset off.
// It will return an error if the offset out of range.
func (beb *BigEndianBuffer) Seek(off int64, whence int) (ret int64, err error) {
	var logp = `Seek`
	var size = int64(len(beb.bin))
	if whence == io.SeekStart {
		if off > size {
			return 0, fmt.Errorf(`%s: offset %d out of range (0-%d)`,
				logp, off, size)
		}
		beb.off = off
		return beb.off, nil
	}
	if whence == io.SeekCurrent {
		off = beb.off + off
		if off > size || off < 0 {
			return 0, fmt.Errorf(`%s: offset %d out of range (max %d)`,
				logp, off, size)
		}
		beb.off = off
		return beb.off, nil
	}
	if whence == io.SeekEnd {
		off = size - off
		if off > size || off < 0 {
			return 0, fmt.Errorf(`%s: offset %d out of range (max %d)`,
				logp, off, size)
		}
		beb.off = off
		return beb.off, nil
	}
	return 0, fmt.Errorf(`%s: invalid whence %d`, logp, whence)
}

// Write any data to binary format.
// The following type of data are ignored: Invalid, Uintptr, Chan, Func,
// Interface, Map; and will return with (0, nil).
func (beb *BigEndianBuffer) Write(data any) (n int, err error) {
	beb.storage.endness = EndianBig
	n, err = beb.storage.encode(data)
	if err != nil {
		return 0, fmt.Errorf(`Write: %w`, err)
	}
	return n, nil
}
