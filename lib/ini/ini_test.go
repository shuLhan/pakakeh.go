// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

const (
	testdataInputIni          = "testdata/input.ini"
	testdataVarWithoutSection = "testdata/var_without_section.ini"
)

type StructA struct {
	X int  `ini:"a::x"`
	Y bool `ini:"a::y"`
}

type StructB struct {
	StructA
	Z float64 `ini:"b::z"`
}

type StructC struct {
	StructB
	XX byte `ini:"c::xx"`
}

type StructMap struct {
	Amap map[string]string `ini:"test:map"`
}

type Y struct {
	String string `ini:"::string"`
	Int    int    `ini:"::int"`
}

type X struct {
	Time time.Time `ini:"section::time" layout:"2006-01-02 15:04:05"`

	PtrBool     *bool          `ini:"section:pointer:bool"`
	PtrDuration *time.Duration `ini:"section:pointer:duration"`
	PtrInt      *int           `ini:"section:pointer:int"`
	PtrString   *string        `ini:"section:pointer:string"`
	PtrTime     *time.Time     `ini:"section:pointer:time" layout:"2006-01-02 15:04:05"`

	PtrStruct    *Y `ini:"section:ptr_struct"`
	PtrStructNil *Y `ini:"section:ptr_struct_nil"`

	Struct Y `ini:"section:struct"`

	String string `ini:"section::string"`

	SliceStruct []Y `ini:"slice:struct"`

	SlicePtrBool     []*bool          `ini:"slice:ptr:bool"`
	SlicePtrDuration []*time.Duration `ini:"slice:ptr:duration"`
	SlicePtrInt      []*int           `ini:"slice:ptr:int"`
	SlicePtrString   []*string        `ini:"slice:ptr:string"`
	SlicePtrStruct   []*Y             `ini:"slice:ptr_struct"`
	SlicePtrTime     []*time.Time     `ini:"slice:ptr:time" layout:"2006-01-02 15:04:05"`

	SliceBool     []bool          `ini:"slice::bool"`
	SliceDuration []time.Duration `ini:"slice::duration"`
	SliceInt      []int           `ini:"slice::int"`
	SliceString   []string        `ini:"slice::string"`
	SliceTime     []time.Time     `ini:"slice::time" layout:"2006-01-02 15:04:05"`

	Duration time.Duration `ini:"section::duration"`
	Int      int           `ini:"section::int"`
	Bool     bool          `ini:"section::bool"`
}

func TestData(t *testing.T) {
	var (
		listTestData []*test.Data
		tdata        *test.Data
		err          error
	)

	listTestData, err = test.LoadDataDir("testdata/struct")
	if err != nil {
		t.Fatal(err)
	}

	for _, tdata = range listTestData {
		t.Run(tdata.Name, func(t *testing.T) {
			var (
				kind   = tdata.Flag["kind"]
				input  = tdata.Input["default"]
				expOut = tdata.Output["default"]
				gotX   = &X{}
				gotC   = &StructC{}
				gotMap = &StructMap{}

				obj    interface{}
				gotOut []byte
				err    error
			)

			switch kind {
			case "":
				return
			case "embedded":
				obj = gotC
			case "map":
				obj = gotMap
			case "struct":
				obj = gotX
			}

			err = Unmarshal(input, obj)
			if err != nil {
				t.Fatal(err)
			}

			gotOut, err = Marshal(obj)
			if err != nil {
				t.Fatal(err)
			}

			test.Assert(t, string(tdata.Desc), string(expOut), string(gotOut))
		})
	}
}

func TestOpen(t *testing.T) {
	cases := []struct {
		desc   string
		inFile string
		expErr string
	}{{
		desc:   "With no file",
		expErr: "open : no such file or directory",
	}, {
		desc:   "With variable without section",
		inFile: testdataVarWithoutSection,
		expErr: "variable without section, line 7 at testdata/var_without_section.ini",
	}, {
		desc:   "With valid file",
		inFile: "testdata/input.ini",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}
	}
}

func TestSave(t *testing.T) {
	cases := []struct {
		desc    string
		inFile  string
		outFile string
		expErr  string
	}{{
		desc:   "With no file",
		expErr: "open : no such file or directory",
	}, {
		desc:   "With variable without section",
		inFile: testdataVarWithoutSection,
		expErr: "variable without section, line 7 at testdata/var_without_section.ini",
	}, {
		desc:   "With empty output file",
		inFile: testdataInputIni,
		expErr: "open : no such file or directory",
	}, {
		desc:    "With valid output file",
		inFile:  testdataInputIni,
		outFile: testdataInputIni + ".save",
	}}

	for _, c := range cases {
		t.Logf(c.desc)

		cfg, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		err = cfg.Save(c.outFile)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
		}
	}
}

func TestAddSection(t *testing.T) {
	in := &Ini{}

	cases := []struct {
		sec    *Section
		expIni *Ini
		desc   string
	}{{
		desc:   "With nil section",
		expIni: &Ini{},
	}, {
		desc: "With valid section",
		sec: &Section{
			mode:      lineModeSection,
			name:      "Test",
			nameLower: "test",
		},
		expIni: &Ini{
			secs: []*Section{{
				mode:      lineModeSection,
				name:      "Test",
				nameLower: "test",
			}},
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		in.addSection(c.sec)

		test.Assert(t, "ini", c.expIni, in)
	}
}

func TestIni_Get(t *testing.T) {
	var (
		cfg       *Ini
		listTData []*test.Data
		tdata     *test.Data
		got       string
		def       string
		tags      []string
		keys      [][]byte
		exps      [][]byte
		key       []byte
		err       error
		x         int
	)

	listTData, err = test.LoadDataDir("testdata/")
	if err != nil {
		t.Fatal(err)
	}

	for _, tdata = range listTData {
		cfg, err = Parse(tdata.Input["default"])
		if err != nil {
			t.Fatal(err)
		}

		keys = bytes.Split(tdata.Input["keys"], []byte("\n"))
		exps = bytes.Split(tdata.Output["default"], []byte("\n"))

		if len(keys) != len(exps) {
			t.Fatalf("%s: input keys length %d does not match with output %d",
				tdata.Name, len(keys), len(exps))
		}

		for x, key = range keys {
			if len(key) == 0 {
				test.Assert(t, "Get", string(exps[x]), "")
				continue
			}

			tags = ParseTag(string(key))
			def = tags[3]

			got, _ = cfg.Get(tags[0], tags[1], tags[2], def)
			got = got + "."

			test.Assert(t, "Get", string(exps[x]), got)
		}
	}
}
