// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package mock provide a mocking for standard output and standard error.
package mock

import (
	"io"
	"io/ioutil"
	"log"
	"os"
)

var (
	_stderr *os.File //nolint: gochecknoglobals
	_stdin  *os.File //nolint: gochecknoglobals
	_stdout *os.File //nolint: gochecknoglobals
)

//
// Close all mocked output and/or error.
//
func Close() {
	if _stderr != nil {
		err := _stderr.Close()
		if err != nil {
			log.Printf("! Close: %s\n", err)
		}
		err = os.Remove(_stderr.Name())
		if err != nil {
			log.Printf("! Close: os.Remove: %s\n", err)
		}
	}
	if _stdin != nil {
		err := _stdin.Close()
		if err != nil {
			log.Printf("! Close: %s\n", err)
		}
		err = os.Remove(_stdin.Name())
		if err != nil {
			log.Printf("! Close: os.Remove: %s\n", err)
		}
	}
	if _stdout != nil {
		err := _stdout.Close()
		if err != nil {
			log.Printf("! Close: %s\n", err)
		}
		err = os.Remove(_stdout.Name())
		if err != nil {
			log.Printf("! Close: os.Remove: %s\n", err)
		}
	}
}

//
// Error get stream of standard error as string.
//
func Error() string {
	if _stderr == nil {
		return ""
	}

	ResetStderr(false)

	bs, err := ioutil.ReadAll(_stderr)
	if err != nil {
		log.Fatal(err)
	}

	return string(bs)
}

//
// Output get stream of standard output.
//
func Output() string {
	if _stdout == nil {
		return ""
	}

	ResetStdout(false)

	bs, err := ioutil.ReadAll(_stdout)
	if err != nil {
		log.Fatal(err)
	}

	return string(bs)
}

//
// Stderr mock standard error to temporary file.
//
func Stderr() *os.File {
	var err error

	_stderr, err = ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}

	return _stderr
}

//
// Stdin mock the standar input using temporary file.
//
func Stdin() *os.File {
	var err error

	_stdin, err = ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}

	return _stdin
}

//
// Stdout mock standard output to temporary file.
//
func Stdout() *os.File {
	var err error

	_stdout, err = ioutil.TempFile("", "")
	if err != nil {
		log.Fatal(err)
	}

	return _stdout
}

//
// Reset all mocked standard output and error.
//
func Reset(truncate bool) {
	ResetStderr(truncate)
	ResetStdout(truncate)
}

//
// ResetStderr reset mocked standard error offset back to 0.
// If truncated is true, it also reset the size to 0.
//
func ResetStderr(truncate bool) {
	if _stderr == nil {
		return
	}

	_, err := _stderr.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}
	if truncate {
		err = _stderr.Truncate(0)
		if err != nil {
			log.Fatal(err)
		}
	}
}

//
// ResetStdin reset mocked standard input offset back to 0.
// If truncated is true, it also reset the size to 0.
//
func ResetStdin(truncate bool) {
	if _stdin == nil {
		return
	}

	_, err := _stdin.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}
	if truncate {
		err = _stdin.Truncate(0)
		if err != nil {
			log.Fatal(err)
		}

	}
}

//
// ResetStdout reset mocked standard output offset back to 0.
// If truncated is true, it also reset the size to 0.
//
func ResetStdout(truncate bool) {
	if _stdout == nil {
		return
	}

	_, err := _stdout.Seek(0, io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}
	if truncate {
		err = _stdout.Truncate(0)
		if err != nil {
			log.Fatal(err)
		}
	}
}
