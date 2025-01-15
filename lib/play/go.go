// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"time"
)

// GoOptions define the options for running and test [Go] code.
type GoOptions struct {
	// Root directory of where the Go code to be written, run, or test.
	// Default to [os.UserCacheDir] if its not set.
	Root string

	// Version define the Go tool version in go.mod to be used to run the
	// code.
	// Default to package [GoVersion] if its not set.
	Version string

	// Timeout define the maximum time the program can be run until it
	// gets terminated.
	// Default to package [Timeout] if its not set.
	Timeout time.Duration
}

func (opts *GoOptions) init() {
	if len(opts.Root) == 0 {
		opts.Root = userCacheDir
	}
	if len(opts.Version) == 0 {
		opts.Version = GoVersion
	}
	if opts.Timeout <= 0 {
		opts.Timeout = Timeout
	}
}

// Go define the type that can format, run, and test Go code.
type Go struct {
	opts GoOptions
}

// NewGo create and initialize new Go tools.
func NewGo(opts GoOptions) (playgo *Go) {
	opts.init()
	playgo = &Go{
		opts: opts,
	}
	return playgo
}
