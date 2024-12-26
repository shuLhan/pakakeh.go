// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary_test

import (
	"bytes"
	"io"
	"math"
	"reflect"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/binary"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestBigEndianBuffer(t *testing.T) {
	var listCase = []struct {
		desc   string
		data   any
		expBin []byte
		expN   int
	}{{
		desc:   `bool=true`,
		data:   bool(true),
		expBin: []byte{0x01},
		expN:   1,
	}, {
		desc:   `bool=false`,
		data:   bool(false),
		expBin: []byte{0x00},
		expN:   1,
	}, {
		desc:   `int8=127`,
		data:   int8(127),
		expBin: []byte{0x7f},
		expN:   1,
	}, {
		desc:   `uint8=255`,
		data:   uint8(255),
		expBin: []byte{0xff},
		expN:   1,
	}, {
		desc:   `int16=32767`,
		data:   int16(math.MaxInt16),
		expBin: []byte{0x7f, 0xff},
		expN:   2,
	}, {
		desc:   `uint16=65535`,
		data:   uint16(math.MaxUint16),
		expBin: []byte{0xff, 0xff},
		expN:   2,
	}, {
		desc:   `int32=2147483647`,
		data:   int32(math.MaxInt32),
		expBin: []byte{0x7f, 0xff, 0xff, 0xff},
		expN:   4,
	}, {
		desc:   `uint32=4294967295`,
		data:   uint32(math.MaxUint32),
		expBin: []byte{0xff, 0xff, 0xff, 0xff},
		expN:   4,
	}, {
		desc: `int64=math.MaxInt64`,
		data: int64(math.MaxInt64),
		expBin: []byte{
			0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `uint64=math.MaxUint64`,
		data: uint64(math.MaxUint64),
		expBin: []byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `int=math.MaxInt`,
		data: int(math.MaxInt),
		expBin: []byte{
			0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `uint=math.MaxUint`,
		data: uint(math.MaxUint),
		expBin: []byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `float32=math.Float32`,
		data: float32(math.MaxFloat32),
		expBin: []byte{
			0x7f, 0x7f, 0xff, 0xff,
		},
		expN: 4,
	}, {
		desc: `float64=math.Float64`,
		data: float64(math.MaxFloat64),
		expBin: []byte{
			0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `complex64=complex(1, 2)`,
		data: complex(float32(1), float32(2)),
		expBin: []byte{
			0x3f, 0x80, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00,
		},
		expN: 8,
	}, {
		desc: `complex128=complex(3,4)`,
		data: complex(float64(3), float64(4)),
		expBin: []byte{
			0x40, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x40, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		expN: 16,
	}, {
		desc: `string="日本語"`,
		data: string(`日本語`),
		expBin: []byte{
			0x00, 0x00, 0x00, 0x09,
			0xe6, 0x97, 0xa5, 0xe6, 0x9c, 0xac, 0xe8, 0xaa, 0x9e,
		},
		expN: 13,
	}}

	var beb binary.BigEndianBuffer

	for _, tcase := range listCase {
		beb.Reset()

		n, err := beb.Write(tcase.data)
		if err != nil {
			t.Fatal(err)
		}

		var gotBin = beb.Bytes()
		test.Assert(t, tcase.desc+` Write n`, tcase.expN, n)
		test.Assert(t, tcase.desc+` Write Bytes`, tcase.expBin, gotBin)

		beb.Seek(0, io.SeekStart)

		// Create new(T) based on the type of tcase.data.
		var gotData = newWithReflect(tcase.data)
		n, err = beb.Read(gotData)
		if err != nil {
			t.Fatal(err)
		}

		// Change the *T to T back for assert.
		gotData = reflect.Indirect(reflect.ValueOf(gotData)).
			Interface()
		test.Assert(t, tcase.desc+` Read n`, tcase.expN, n)
		test.Assert(t, tcase.desc+` Read data`, tcase.data, gotData)
	}
}

func TestBigEndianBuffer_Array(t *testing.T) {
	var listCase = []struct {
		desc   string
		data   any
		expBin []byte
		expN   int
	}{{
		desc: `[2]bool`,
		data: [2]bool{false, true},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x01,
		},
		expN: 6,
	}, {
		desc: `[3]int32`,
		data: [3]int32{1, 2, 3},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x03,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x03,
		},
		expN: 16,
	}, {
		desc: `[2]uint64`,
		data: [2]uint64{100, 200},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc8,
		},
		expN: 20,
	}, {
		desc: `empty array`,
		data: [0]int{},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x00,
		},
		expN: 4,
	}}

	var beb binary.BigEndianBuffer

	for _, tcase := range listCase {
		beb.Reset()

		t.Run(`Write `+tcase.desc, func(t *testing.T) {
			n, err := beb.Write(tcase.data)
			if err != nil {
				t.Fatal(err)
			}

			var gotBin = beb.Bytes()
			test.Assert(t, `n`, tcase.expN, n)
			test.Assert(t, `Bytes`, tcase.expBin, gotBin)
		})

		t.Run(`Read `+tcase.desc, func(t *testing.T) {
			beb.Seek(0, io.SeekStart)

			// Create pointer to array based on the type of
			// tcase.data.
			var gotData = newWithReflect(tcase.data)
			n, err := beb.Read(gotData)
			if err != nil {
				t.Fatal(err)
			}

			// Change the *T to T back for assert.
			gotData = reflect.Indirect(reflect.ValueOf(gotData)).
				Interface()

			test.Assert(t, `n`, tcase.expN, n)
			test.Assert(t, `data`, tcase.data, gotData)
		})
	}
}

func TestBigEndianBuffer_IgnoredType(t *testing.T) {
	var iface io.Writer = &bytes.Buffer{}

	var listCase = []struct {
		data any
		desc string
	}{{
		desc: `nil`,
	}, {
		desc: `uintptr`,
		data: uintptr(0xFF),
	}, {
		desc: `Chan`,
		data: make(chan int, 1),
	}, {
		desc: `Func`,
		data: func() {},
	}, {
		desc: `Interface`,
		data: iface,
	}, {
		desc: `Map`,
		data: map[int]int{},
	}}

	var beb binary.BigEndianBuffer
	var expBin []byte
	for _, tcase := range listCase {
		beb.Reset()

		n, err := beb.Write(tcase.data)
		if err != nil {
			t.Fatal(err)
		}

		var got = beb.Bytes()
		test.Assert(t, `n`, 0, n)
		test.Assert(t, `Bytes`, expBin, got)
	}
}

func TestBigEndianBuffer_Pointer(t *testing.T) {
	var listCase = []struct {
		desc   string
		data   func() any
		expBin []byte
		expN   int
	}{{
		desc:   `*bool=true`,
		data:   func() any { val := bool(true); return &val },
		expBin: []byte{0x01},
		expN:   1,
	}, {
		desc:   `*bool=false`,
		data:   func() any { val := bool(false); return &val },
		expBin: []byte{0x00},
		expN:   1,
	}, {
		desc:   `*int8=127`,
		data:   func() any { val := int8(127); return &val },
		expBin: []byte{0x7f},
		expN:   1,
	}, {
		desc:   `*uint8=255`,
		data:   func() any { val := uint8(255); return &val },
		expBin: []byte{0xff},
		expN:   1,
	}, {
		desc:   `*int16=32767`,
		data:   func() any { val := int16(math.MaxInt16); return &val },
		expBin: []byte{0x7f, 0xff},
		expN:   2,
	}, {
		desc:   `*uint16=65535`,
		data:   func() any { val := uint16(math.MaxUint16); return &val },
		expBin: []byte{0xff, 0xff},
		expN:   2,
	}, {
		desc:   `*int32=2147483647`,
		data:   func() any { val := int32(math.MaxInt32); return &val },
		expBin: []byte{0x7f, 0xff, 0xff, 0xff},
		expN:   4,
	}, {
		desc:   `*uint32=4294967295`,
		data:   func() any { val := uint32(math.MaxUint32); return &val },
		expBin: []byte{0xff, 0xff, 0xff, 0xff},
		expN:   4,
	}, {
		desc: `*int64=math.MaxInt64`,
		data: func() any { val := int64(math.MaxInt64); return &val },
		expBin: []byte{
			0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `*uint64=math.MaxUint64`,
		data: func() any { val := uint64(math.MaxUint64); return &val },
		expBin: []byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `*int=math.MaxInt`,
		data: func() any { val := int(math.MaxInt); return &val },
		expBin: []byte{
			0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `*uint=math.MaxUint`,
		data: func() any { val := uint(math.MaxUint); return &val },
		expBin: []byte{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `*float32=math.Float32`,
		data: func() any { val := float32(math.MaxFloat32); return &val },
		expBin: []byte{
			0x7f, 0x7f, 0xff, 0xff,
		},
		expN: 4,
	}, {
		desc: `*float64=math.Float64`,
		data: func() any { val := float64(math.MaxFloat64); return &val },
		expBin: []byte{
			0x7f, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		},
		expN: 8,
	}, {
		desc: `*complex64=complex(1, 2)`,
		data: func() any { val := complex(float32(1), float32(2)); return &val },
		expBin: []byte{
			0x3f, 0x80, 0x00, 0x00, 0x40, 0x00, 0x00, 0x00,
		},
		expN: 8,
	}, {
		desc: `*complex128=complex(3,4)`,
		data: func() any { val := complex(float64(3), float64(4)); return &val },
		expBin: []byte{
			0x40, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x40, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		},
		expN: 16,
	}, {
		desc: `*string="日本語"`,
		data: func() any { val := string(`日本語`); return &val },
		expBin: []byte{
			0x00, 0x00, 0x00, 0x09,
			0xe6, 0x97, 0xa5, 0xe6, 0x9c, 0xac, 0xe8, 0xaa, 0x9e,
		},
		expN: 13,
	}}

	var beb binary.BigEndianBuffer

	for _, tcase := range listCase {
		beb.Reset()

		var data = tcase.data()
		n, err := beb.Write(data)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, tcase.desc+` Write n`, tcase.expN, n)
		var gotBin = beb.Bytes()
		test.Assert(t, tcase.desc+` Write Bytes`, tcase.expBin, gotBin)

		beb.Seek(0, io.SeekStart)

		// Create new(T) based on the type of tcase.data.
		var gotData = newWithReflect(data)
		n, err = beb.Read(gotData)
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, tcase.desc+` Read n`, tcase.expN, n)
		test.Assert(t, tcase.desc+` Read data`, data, gotData)
	}
}

func TestBigEndianBuffer_Slices(t *testing.T) {
	var listCase = []struct {
		desc   string
		data   any
		expBin []byte
		expN   int
	}{{
		desc: `[]bool`,
		data: []bool{false, true},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x01,
		},
		expN: 6,
	}, {
		desc: `[]int8`,
		data: []int8{1, 2, 3},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x03,
			0x01, 0x02, 0x03,
		},
		expN: 7,
	}, {
		desc: `[]byte`,
		data: []byte{1, 2, 3},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x03,
			0x01, 0x02, 0x03,
		},
		expN: 7,
	}, {
		desc: `[]int16`,
		data: []int16{1, 2, 3},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x03,
			0x00, 0x01, 0x00, 0x02, 0x00, 0x03,
		},
		expN: 10,
	}, {
		desc: `[]uint16`,
		data: []uint16{1, 2, 3},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x03,
			0x00, 0x01, 0x00, 0x02, 0x00, 0x03,
		},
		expN: 10,
	}, {
		desc: `[]int32`,
		data: []int32{1, 2, 3},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x03,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x03,
		},
		expN: 16,
	}, {
		desc: `[]uint32`,
		data: []uint32{1, 2, 3},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x03,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x03,
		},
		expN: 16,
	}, {
		desc: `[]int64`,
		data: []int64{100, 200},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc8,
		},
		expN: 20,
	}, {
		desc: `[]uint64`,
		data: []uint64{100, 200},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc8,
		},
		expN: 20,
	}, {
		desc: `[]int`,
		data: []int{},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x00,
		},
		expN: 4,
	}, {
		desc: `[]uint`,
		data: []uint{100, 200},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xc8,
		},
		expN: 20,
	}, {
		desc: `[]float32`,
		data: []float32{3.45, 6.78},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x02,
			0x40, 0x5c, 0xcc, 0xcd,
			0x40, 0xd8, 0xf5, 0xc3,
		},
		expN: 12,
	}, {
		desc: `[]float64`,
		data: []float64{3.45, 6.78},
		expBin: []byte{
			0x00, 0x00, 0x00, 0x02,
			0x40, 0x0b, 0x99, 0x99, 0x99, 0x99, 0x99, 0x9a,
			0x40, 0x1b, 0x1e, 0xb8, 0x51, 0xeb, 0x85, 0x1f,
		},
		expN: 20,
	}}

	var beb binary.BigEndianBuffer

	for _, tcase := range listCase {
		beb.Reset()

		t.Run(`Write `+tcase.desc, func(t *testing.T) {
			n, err := beb.Write(tcase.data)
			if err != nil {
				t.Fatal(err)
			}

			var gotBin = beb.Bytes()
			test.Assert(t, `n`, tcase.expN, n)
			test.Assert(t, `Bytes`, tcase.expBin, gotBin)
		})

		t.Run(`Read `+tcase.desc, func(t *testing.T) {
			beb.Seek(0, io.SeekStart)

			// Create pointer to slices based on the type of
			// tcase.data.
			var gotData = newWithReflect(tcase.data)
			n, err := beb.Read(gotData)
			if err != nil {
				t.Fatal(err)
			}

			// Change the *T to T back for assert.
			gotData = reflect.Indirect(reflect.ValueOf(gotData)).
				Interface()
			test.Assert(t, `n`, tcase.expN, n)
			test.Assert(t, `data`, tcase.data, gotData)
		})

		t.Run(`Read with length `+tcase.desc, func(t *testing.T) {
			beb.Seek(0, io.SeekStart)

			// Create slices based on the type of tcase.data.

			var sliceType = reflect.TypeOf(tcase.data)
			var sliceLen = reflect.ValueOf(tcase.data).Len()
			var gotData = reflect.
				MakeSlice(sliceType, sliceLen, sliceLen).
				Interface()

			n, err := beb.Read(gotData)
			if err != nil {
				t.Fatal(err)
			}

			test.Assert(t, `n`, tcase.expN, n)
			test.Assert(t, `data`, tcase.data, gotData)
		})
	}
}

func TestBigEndianBuffer_Struct(t *testing.T) {
	var vfloat32 float32 = 3

	type testStruct struct {
		PtrFloat32 *float32
		MapIntInt  map[int]int
		String     string
		SliceByte  []byte
		Int32      int32
		Bool       bool
	}

	var listCase = []struct {
		desc   string
		data   any
		expBin []byte
		expN   int
	}{{
		desc: `empty struct`,
		data: testStruct{},
		expBin: []byte{
			0x00,                   // PtrFloat32
			0x00, 0x00, 0x00, 0x00, // String
			0x00, 0x00, 0x00, 0x00, // SliceByte
			0x00, 0x00, 0x00, 0x00, // Int32
			0x00, // Bool
		},
		expN: 14,
	}, {
		desc: `all set`,
		data: testStruct{
			PtrFloat32: &vfloat32,
			String:     `hello`,
			SliceByte:  []byte{4, 5},
			Int32:      2,
			Bool:       true,
		},
		expBin: []byte{
			0x01, // PtrFloat32
			0x40, 0x40, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x05, // String
			0x68, 0x65, 0x6c, 0x6c, 0x6f,
			0x00, 0x00, 0x00, 0x02, // SliceByte
			0x04, 0x05,
			0x00, 0x00, 0x00, 0x02, // Int32
			0x01, // Bool
		},
		expN: 25,
	}}

	var beb binary.BigEndianBuffer

	for _, tcase := range listCase {
		beb.Reset()

		n, err := beb.Write(tcase.data)
		if err != nil {
			t.Fatal(err)
		}

		var gotBin = beb.Bytes()
		test.Assert(t, tcase.desc+` Write n`, tcase.expN, n)
		test.Assert(t, tcase.desc+` Write Bytes`, tcase.expBin, gotBin)

		beb.Seek(0, io.SeekStart)

		// Create new(T) based on the type of tcase.data.
		var gotData = newWithReflect(tcase.data)
		n, err = beb.Read(gotData)
		if err != nil {
			t.Fatal(err)
		}

		// Change the *T to T back for assert.
		gotData = reflect.Indirect(reflect.ValueOf(gotData)).
			Interface()
		test.Assert(t, tcase.desc+` Read n`, tcase.expN, n)
		test.Assert(t, tcase.desc+` Read data`, tcase.data, gotData)
	}
}

func TestBigEndianBuffer_ReadFailed(t *testing.T) {
	var vint int

	var listCase = []struct {
		desc        string
		dataRead    any
		dataWrite   any
		expErrRead  string
		expErrWrite string
	}{{
		desc:       `dataRead=nil`,
		dataWrite:  bool(true),
		expErrRead: `Read: expecting pointer to bool`,
	}, {
		desc:       `dataRead not pointer`,
		dataWrite:  bool(true),
		dataRead:   vint,
		expErrRead: `Read: expecting pointer to int`,
	}, {
		desc:        `[]map[int]int`,
		dataWrite:   []map[int]int{},
		expErrWrite: `Write: unsupported type []map[int]int`,
	}}

	var beb binary.BigEndianBuffer

	for _, tcase := range listCase {
		beb.Reset()

		_, err := beb.Write(tcase.dataWrite)
		if err != nil {
			test.Assert(t, `Write error`, tcase.expErrWrite,
				err.Error())
		}

		_, err = beb.Seek(0, io.SeekStart)
		if err != nil {
			t.Fatal(err)
		}
		_, err = beb.Read(tcase.dataRead)
		if err != nil {
			test.Assert(t, `Read error`, tcase.expErrRead,
				err.Error())
		}
	}
}

// newWithReflect create new value with non-pointer type of in.
func newWithReflect(in any) (out any) {
	var refval = reflect.Indirect(reflect.ValueOf(in))
	return reflect.New(refval.Type()).Interface()
}
