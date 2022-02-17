// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package mlog implement buffered multi writers of log.
// It can have zero or more normal writers (for example, os.Stdout and
// os.File) and zero or more error writers (for example, os.Stderr and
// different os.File).
//
// The MultiLogger is buffered to minimize waiting time when writing to
// multiple writers that have different latencies.
// For example, if we have one writer to os.Stdout, one writer to file, and
// one writer to network; the writer to network may take more time to finish
// than to os.Stdout and file, which may slowing down the program if we want
// to wait for all writes to finish.
//
// For this reason, do not forget to call Flush when your program exit.
//
// The default MultiLogger use time.RFC3339 as the default time layout, empty
// prefix, os.Stdout for the output writer, and os.Stderr for the error
// writer.
//
// Format of written log,
//
//	[time] [prefix] <message>
//
// The time and prefix only printed if its not empty, and the single space is
// added for convenience.
//
package mlog

import (
	"io"
	"os"
)

const (
	defTimeFormat = "2006-01-02 15:04:05 MST"
)

var defaultMLog MultiLogger = createMultiLogger(defTimeFormat, "",
	[]NamedWriter{
		NewNamedWriter("stdout", os.Stdout),
	},
	[]NamedWriter{
		NewNamedWriter("stderr", os.Stderr),
	})

//
// Errf write to all registered error writers.
// The default registered error writer is os.Stderr.
//
// If the generated string does not end with new line, it will be added.
//
func Errf(format string, v ...interface{}) {
	defaultMLog.Errf(format, v...)
}

//
// Fatalf is equal to Errf and os.Exit(1).
//
func Fatalf(format string, v ...interface{}) {
	defaultMLog.Fatalf(format, v...)
}

//
// Flush all writes in queue and wait until it finished.
//
func Flush() {
	defaultMLog.Flush()
}

//
// Outf write to all registered output writers.
// The default registered output writer is os.Stdout.
//
// If the generated string does not end with new line, it will be added.
//
func Outf(format string, v ...interface{}) {
	defaultMLog.Outf(format, v...)
}

//
// Panicf is equal to Errf followed by panic.
//
func Panicf(format string, v ...interface{}) {
	defaultMLog.Panicf(format, v...)
}

//
// PrintStack writes to error writers the stack trace returned by
// debug.Stack.
//
// This function can be useful when debugging panic using recover in the main
// program by logging the stack trace.
// For example,
//
//	err := recover()
//	if err != nil {
//		mlog.PrintStack()
//		os.Exit(1)
//	}
//
func PrintStack() {
	defaultMLog.PrintStack()
}

//
// RegisterErrorWriter register the named writer to one of error writers.
//
func RegisterErrorWriter(errw NamedWriter) {
	defaultMLog.RegisterErrorWriter(errw)
}

//
// RegisterOutputWriter register the named writer to one of output writers.
//
func RegisterOutputWriter(outw NamedWriter) {
	defaultMLog.RegisterOutputWriter(outw)
}

//
// SetPrefix set the default prefix for the subsequence writes.
//
func SetPrefix(prefix string) {
	defaultMLog.SetPrefix(prefix)
}

//
// SetTimeFormat set the default time format for the subsequence writes.
//
func SetTimeFormat(layout string) {
	defaultMLog.SetTimeFormat(layout)
}

//
// UnregisterErrorWriter remove the error writer by name.
//
func UnregisterErrorWriter(name string) {
	defaultMLog.UnregisterErrorWriter(name)
}

//
// UnregisterOutputWriter remove the output writer by name.
//
func UnregisterOutputWriter(name string) {
	defaultMLog.UnregisterOutputWriter(name)
}

//
// ErrorWriter return the internal default MultiLogger.
// A call to Write() on returned io.Writer will forward it to all registered
// error writers.
//
func ErrorWriter() io.Writer {
	return &defaultMLog
}
