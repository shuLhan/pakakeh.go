// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"reflect"
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestUnpackStruct(t *testing.T) {
	type U struct {
		String string `ini:"::string"`
		Int    int    `ini:"::int"`
	}

	type T struct {
		Time time.Time `ini:"section::time" layout:"2006-01-02 15:04:05"`

		PtrStruct    *U
		PtrStructNil *U

		PtrString *string `ini:"section:pointer"`
		PtrInt    *int    `ini:"section:pointer"`

		MapString map[string]string `ini:"section:mapstring"`
		MapInt    map[string]int    `ini:"section:mapint"`

		String string `ini:"section::string"`

		SliceString []string `ini:"section:slice:string"`
		SliceInt    []int    `ini:"section:slice:int"`
		SliceUint   []uint   `ini:"section:slice:uint"`
		SliceStruct []U      `ini:"slice:OfStruct"`
		SliceBool   []bool   `ini:"section:slice:bool"`

		Struct U

		Duration time.Duration `ini:"section::duration"`
		Int      int           `ini:"section::int"`
		Bool     bool          `ini:"section::bool"`
	}

	var v interface{} = &T{
		PtrStruct: &U{},
	}

	rtype := reflect.TypeOf(v)
	rval := reflect.ValueOf(v)
	rtype = rtype.Elem()
	rval = rval.Elem()
	got := unpackTagStructField(rtype, rval)

	exp := []string{
		"::int",
		"::string",
		"section::bool",
		"section::duration",
		"section::int",
		"section::string",
		"section::time",
		"section:mapint",
		"section:mapstring",
		"section:pointer",
		"section:slice:bool",
		"section:slice:int",
		"section:slice:string",
		"section:slice:uint",
		"slice:OfStruct",
	}

	test.Assert(t, "unpackTagStructField", exp, got.keys())
}

func TestUnpackStruct_embedded(t *testing.T) {
	type A struct {
		X int  `ini:"a::x"`
		Y bool `ini:"a::y"`
	}

	type B struct {
		A
		Z float64 `ini:"b::z"`
	}

	type C struct {
		B
		XX byte `ini:"c::xx"`
	}

	var v interface{} = &C{}

	rtype := reflect.TypeOf(v)
	rval := reflect.ValueOf(v)
	rtype = rtype.Elem()
	rval = rval.Elem()
	got := unpackTagStructField(rtype, rval)

	exp := []string{
		"a::x",
		"a::y",
		"b::z",
		"c::xx",
	}
	test.Assert(t, "unpackTagStructField: embedded", exp, got.keys())
}
