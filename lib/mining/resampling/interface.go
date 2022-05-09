// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package resampling provide common interface, constants, and methods for
// resampling modules.
package resampling

import (
	"github.com/shuLhan/share/lib/dsv"
	"github.com/shuLhan/share/lib/tabula"
)

const (
	// DefaultK nearest neighbors.
	DefaultK = 5
	// DefaultPercentOver sampling.
	DefaultPercentOver = 100
)

// Interface define common methods used by resampling module.
type Interface interface {
	GetSynthetics() tabula.DatasetInterface
}

// WriteSynthetics will write synthetic samples in resampling module `ri` into
// `file`.
func WriteSynthetics(ri Interface, file string) (e error) {
	writer, e := dsv.NewWriter("")
	if nil != e {
		return
	}

	e = writer.OpenOutput(file)
	if e != nil {
		return
	}

	sep := dsv.DefSeparator
	_, e = writer.WriteRawDataset(ri.GetSynthetics(), &sep)
	if e != nil {
		return
	}

	return writer.Close()
}
