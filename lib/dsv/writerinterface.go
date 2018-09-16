// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dsv

//
// WriterInterface is an interface for writing DSV data to file.
//
type WriterInterface interface {
	ConfigInterface
	GetOutput() string
	SetOutput(path string)
	OpenOutput(file string) error
	Flush() error
	Close() error
}

//
// OpenWriter configuration file and initialize the attributes.
//
func OpenWriter(writer WriterInterface, fcfg string) (e error) {
	e = ConfigOpen(writer, fcfg)
	if e != nil {
		return
	}

	return InitWriter(writer)
}

//
// InitWriter initialize writer by opening output file.
//
func InitWriter(writer WriterInterface) error {
	out := writer.GetOutput()

	// Exit immediately if no output file is defined in config.
	if "" == out {
		return ErrNoOutput
	}

	writer.SetOutput(ConfigCheckPath(writer, out))

	return writer.OpenOutput("")
}
