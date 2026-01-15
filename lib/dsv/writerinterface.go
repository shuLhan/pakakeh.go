// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.

package dsv

// WriterInterface is an interface for writing DSV data to file.
type WriterInterface interface {
	ConfigInterface
	GetOutput() string
	SetOutput(path string)
	OpenOutput(file string) error
	Flush() error
	Close() error
}

// OpenWriter configuration file and initialize the attributes.
func OpenWriter(writer WriterInterface, fcfg string) (e error) {
	e = ConfigOpen(writer, fcfg)
	if e != nil {
		return
	}

	return InitWriter(writer)
}

// InitWriter initialize writer by opening output file.
func InitWriter(writer WriterInterface) error {
	out := writer.GetOutput()

	// Exit immediately if no output file is defined in config.
	if len(out) == 0 {
		return ErrNoOutput
	}

	writer.SetOutput(ConfigCheckPath(writer, out))

	return writer.OpenOutput("")
}
