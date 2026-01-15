// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.

package dsv

import (
	"testing"
)

const (
	testdataSimpleRead = "testdata/config_simpleread.dsv"
)

// doInit create read-write object.
func doInit(t *testing.T, fcfg string) (rw *ReadWriter, e error) {
	// Initialize dsv
	rw, e = New(fcfg, nil)

	if nil != e {
		t.Fatal(e)
	}

	return
}

// TestReadWriter test reading and writing DSV.
func TestReadWriter(t *testing.T) {
	rw, _ := doInit(t, "testdata/config.dsv")

	doReadWrite(t, &rw.Reader, &rw.Writer, expectation, true)

	e := rw.Close()
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, rw.GetOutput(), "testdata/expected.dat")
}

// TestReadWriter test reading and writing DSV.
func TestReadWriterAll(t *testing.T) {
	rw, _ := doInit(t, "testdata/config.dsv")

	rw.SetMaxRows(-1)

	doReadWrite(t, &rw.Reader, &rw.Writer, expectation, false)

	e := rw.Close()
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, rw.GetOutput(), "testdata/expected.dat")
}

func TestSimpleReadWrite(t *testing.T) {
	fcfg := testdataSimpleRead

	reader, e := SimpleRead(fcfg, nil)
	if e != nil {
		t.Fatal(e)
	}

	fout := "testdata/output.dat"
	fexp := "testdata/expected.dat"

	_, e = SimpleWrite(reader, fcfg)
	if e != nil {
		t.Fatal(e)
	}

	assertFile(t, fexp, fout)
}

func TestSimpleMerge(t *testing.T) {
	fcfg1 := testdataSimpleRead
	fcfg2 := testdataSimpleRead

	reader, e := SimpleMerge(fcfg1, fcfg2, nil, nil)
	if e != nil {
		t.Fatal(e)
	}

	_, e = SimpleWrite(reader, fcfg1)
	if e != nil {
		t.Fatal(e)
	}

	fexp := "testdata/expected_simplemerge.dat"
	fout := "testdata/output.dat"

	assertFile(t, fexp, fout)
}
