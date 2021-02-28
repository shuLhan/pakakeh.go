// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

type Y struct {
	String string `ini:"::string"`
	Int    int    `ini:"::int"`
}

type X struct {
	Struct       Y  `ini:"section:struct"`
	PtrStruct    *Y `ini:"section:ptr_struct"`
	PtrStructNil *Y `ini:"section:ptr_struct_nil"`

	String   string        `ini:"section::string"`
	Int      int           `ini:"section::int"`
	Bool     bool          `ini:"section::bool"`
	Duration time.Duration `ini:"section::duration"`
	Time     time.Time     `ini:"section::time" layout:"2006-01-02 15:04:05"`

	PtrString   *string        `ini:"section:pointer:string"`
	PtrInt      *int           `ini:"section:pointer:int"`
	PtrBool     *bool          `ini:"section:pointer:bool"`
	PtrDuration *time.Duration `ini:"section:pointer:duration"`
	PtrTime     *time.Time     `ini:"section:pointer:time" layout:"2006-01-02 15:04:05"`

	SliceStruct   []Y             `ini:"slice:struct"`
	SliceString   []string        `ini:"slice::string"`
	SliceInt      []int           `ini:"slice::int"`
	SliceBool     []bool          `ini:"slice::bool"`
	SliceDuration []time.Duration `ini:"slice::duration"`
	SliceTime     []time.Time     `ini:"slice::time" layout:"2006-01-02 15:04:05"`

	SlicePtrStruct   []*Y             `ini:"slice:ptr_struct"`
	SlicePtrString   []*string        `ini:"slice:ptr:string"`
	SlicePtrInt      []*int           `ini:"slice:ptr:int"`
	SlicePtrBool     []*bool          `ini:"slice:ptr:bool"`
	SlicePtrDuration []*time.Duration `ini:"slice:ptr:duration"`
	SlicePtrTime     []*time.Time     `ini:"slice:ptr:time" layout:"2006-01-02 15:04:05"`
}

func TestIni_Unmarshal(t *testing.T) {
	iniText := `
[section "struct"]
string = struct
int = 1

[section "ptr_struct"]
string = ptr_struct
int = 2

[section "ptr_struct_nil"]
string = ptr_struct_nil
int = 3

[section]
string = a string
int = 4
bool = true
duration = 4m
time = 2021-02-28 00:12:04

[section "pointer"]
string = pointer to string
int = 5
bool = true
duration = 5m
time = 2021-02-28 00:12:05
`

	got := &X{
		PtrStruct: &Y{},
	}

	err := Unmarshal([]byte(iniText), got)
	if err != nil {
		t.Fatal(err)
	}

	ptrString := "pointer to string"
	ptrInt := 5
	ptrBool := true
	ptrDuration := time.Duration(5 * time.Minute)
	ptrTime := time.Date(2021, time.February, 28, 0, 12, 5, 0, time.UTC)

	exp := &X{
		Struct: Y{
			String: "struct",
			Int:    1,
		},
		PtrStruct: &Y{
			String: "ptr_struct",
			Int:    2,
		},
		PtrStructNil: &Y{
			String: "ptr_struct_nil",
			Int:    3,
		},

		String:   "a string",
		Int:      4,
		Bool:     true,
		Duration: time.Duration(4 * time.Minute),
		Time:     time.Date(2021, time.February, 28, 0, 12, 4, 0, time.UTC),

		PtrString:   &ptrString,
		PtrInt:      &ptrInt,
		PtrBool:     &ptrBool,
		PtrDuration: &ptrDuration,
		PtrTime:     &ptrTime,
	}

	test.Assert(t, "Unmarshal", exp, got, true)
}

func TestIni_Unmarshal_sliceOfStruct(t *testing.T) {
	iniText := `
[slice "struct"]
string = struct 0
int = 1

[slice "struct"]
string = struct 1
int = 2
`
	got := &X{}

	err := Unmarshal([]byte(iniText), got)
	if err != nil {
		t.Fatal(err)
	}

	exp := &X{
		SliceStruct: []Y{{
			String: "struct 0",
			Int:    1,
		}, {
			String: "struct 1",
			Int:    2,
		}},
	}

	test.Assert(t, "Unmarshal slice of struct", exp, got, true)
}

func TestIni_Unmarshal_sliceOfPrimitive(t *testing.T) {
	iniText := `
[slice]
string = string 0
int = 1
int = 2
bool = true
duration = 1s
time = 2021-02-28 03:56:01

[slice]
string = string 1
string = string 2
int = 3
bool = false
duration = 2s
time = 2021-02-28 03:56:02
`
	got := &X{}

	err := Unmarshal([]byte(iniText), got)
	if err != nil {
		t.Fatal(err)
	}

	exp := &X{
		SliceString: []string{"string 0", "string 1", "string 2"},
		SliceInt:    []int{1, 2, 3},
		SliceBool:   []bool{true, false},
		SliceDuration: []time.Duration{
			time.Second,
			2 * time.Second,
		},
		SliceTime: []time.Time{
			time.Date(2021, time.February, 28, 3, 56, 1, 0, time.UTC),
			time.Date(2021, time.February, 28, 3, 56, 2, 0, time.UTC),
		},
	}

	test.Assert(t, "Unmarshal slice of primitive", exp, got, true)
}

func TestIni_Unmarshal_sliceOfPointer(t *testing.T) {
	iniText := `
[slice "ptr_struct"]
string = ptr_struct 0
int = 1

[slice "ptr_struct"]
string = ptr_struct 1
int = 2

[slice "ptr"]
string = string 0
int = 1
bool = true
duration = 1s
time = 2021-02-28 03:56:01

[slice "ptr"]
string = string 1
int = 2
bool = false
duration = 2s
time = 2021-02-28 03:56:02
`
	got := &X{}

	err := Unmarshal([]byte(iniText), got)
	if err != nil {
		t.Fatal(err)
	}

	str0 := "string 0"
	str1 := "string 1"
	int0 := 1
	int1 := 2
	bool0 := true
	bool1 := false
	dur0 := time.Second
	dur1 := time.Second * 2
	time0 := time.Date(2021, time.February, 28, 3, 56, 1, 0, time.UTC)
	time1 := time.Date(2021, time.February, 28, 3, 56, 2, 0, time.UTC)
	exp := &X{
		SlicePtrStruct: []*Y{{
			String: "ptr_struct 0",
			Int:    1,
		}, {
			String: "ptr_struct 1",
			Int:    2,
		}},
		SlicePtrString:   []*string{&str0, &str1},
		SlicePtrInt:      []*int{&int0, &int1},
		SlicePtrBool:     []*bool{&bool0, &bool1},
		SlicePtrDuration: []*time.Duration{&dur0, &dur1},
		SlicePtrTime:     []*time.Time{&time0, &time1},
	}

	test.Assert(t, "Unmarshal slice of pointer", exp, got, true)
}
