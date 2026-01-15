// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2016 Mhd Sulhan <ms@kilabit.info>

// Package resampling provide common interface, constants, and methods for
// resampling modules.
package resampling

import (
	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
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
