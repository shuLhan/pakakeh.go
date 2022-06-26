// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mlog

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

// MultiLogger represent a single log writer that write to multiple outputs.
// MultiLogger can have zero or more writers for standard output (normal log)
// and zero or more writers for standard error.
//
// Each call to write APIs (Errf, Fatalf, or Outf) will be prefixed with
// time format in UTC and optional prefix.
type MultiLogger struct {
	bufPool *sync.Pool

	qerr chan []byte
	qout chan []byte

	qerrFlush chan bool
	qoutFlush chan bool

	errs map[string]NamedWriter
	outs map[string]NamedWriter

	timeFormat string

	prefix []byte

	sync.Mutex
	isClosed bool
}

// NewMultiLogger create and initialize new MultiLogger.
func NewMultiLogger(timeFormat, prefix string, outs, errs []NamedWriter) *MultiLogger {
	return createMultiLogger(timeFormat, prefix, outs, errs)
}

func createMultiLogger(timeFormat, prefix string, outs, errs []NamedWriter) (mlog *MultiLogger) {
	var (
		w    NamedWriter
		name string
	)

	mlog = &MultiLogger{
		bufPool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
		timeFormat: timeFormat,
		prefix:     []byte(prefix),
		outs:       make(map[string]NamedWriter, len(outs)),
		errs:       make(map[string]NamedWriter, len(errs)),
		qout:       make(chan []byte, 512),
		qerr:       make(chan []byte, 512),
		qerrFlush:  make(chan bool, 1),
		qoutFlush:  make(chan bool, 1),
	}
	for _, w = range outs {
		name = w.Name()
		if len(name) == 0 {
			continue
		}
		mlog.outs[name] = w
	}
	for _, w = range errs {
		name = w.Name()
		if len(name) == 0 {
			continue
		}
		mlog.errs[name] = w
	}

	go mlog.processErrorQueue()
	go mlog.processOutputQueue()
	return mlog
}

func flush(qlog chan []byte, writers map[string]NamedWriter) {
	var (
		name string
		err  error
		nw   NamedWriter
		b    []byte
		x    int
	)

	for x = 0; x < len(qlog); x++ {
		b = <-qlog
		if len(b) == 0 {
			b = append(b, '\n')
		} else if b[len(b)-1] != '\n' {
			b = append(b, '\n')
		}
		for name, nw = range writers {
			_, err = nw.Write(b)
			if err != nil {
				log.Printf("MultiLogger: %s: %s", name, err)
			}
		}
	}
}

// Close flush and close all log forwarders.
// Any write to a closed MultiLogger will be ignored.
func (mlog *MultiLogger) Close() {
	mlog.Lock()
	mlog.isClosed = true
	close(mlog.qerr)
	close(mlog.qout)
	<-mlog.qerrFlush
	<-mlog.qoutFlush
	mlog.Unlock()
}

// Errf write the formatted string and its optional values to all error
// writers.
//
// If the generated string does not end with new line, it will be added.
func (mlog *MultiLogger) Errf(format string, v ...interface{}) {
	mlog.writeTo(mlog.qerr, format, v...)
}

// Fatalf is equal to Errf and os.Exit(1).
func (mlog *MultiLogger) Fatalf(format string, v ...interface{}) {
	mlog.Errf(format, v...)
	mlog.Flush()
	os.Exit(1)
}

// Flush all writes and wait until it finished.
func (mlog *MultiLogger) Flush() {
	mlog.Lock()
	if mlog.isClosed {
		mlog.Unlock()
		return
	}
	mlog.Unlock()

	mlog.qerrFlush <- true
	mlog.qoutFlush <- true
	<-mlog.qerrFlush
	<-mlog.qoutFlush
}

// Outf write the formatted string and its values to all output writers.
//
// If the generated string does not end with new line, it will be added.
func (mlog *MultiLogger) Outf(format string, v ...interface{}) {
	mlog.writeTo(mlog.qout, format, v...)
}

// Panicf is equal to Errf and followed by panic.
func (mlog *MultiLogger) Panicf(format string, v ...interface{}) {
	mlog.Errf(format, v...)
	mlog.Flush()
	var msg string = fmt.Sprintf(format, v...)
	panic(msg)
}

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
func (mlog *MultiLogger) PrintStack() {
	mlog.Errf("%s", debug.Stack())
	mlog.Flush()
}

// RegisterErrorWriter register the named writer to one of error writers.
// The writer Name() must not be empty or it will not registered.
func (mlog *MultiLogger) RegisterErrorWriter(errw NamedWriter) {
	var name string = errw.Name()
	if len(name) == 0 {
		return
	}
	mlog.errs[name] = errw
}

// RegisterOutputWriter register the named writer to one of output writers.
// The writer Name() must not be empty or it will not registered.
func (mlog *MultiLogger) RegisterOutputWriter(outw NamedWriter) {
	var name string = outw.Name()
	if len(name) == 0 {
		return
	}
	mlog.outs[name] = outw
}

// SetPrefix set the default prefix for the subsequence writes.
func (mlog *MultiLogger) SetPrefix(prefix string) {
	mlog.prefix = []byte(prefix)
}

// SetTimeFormat set the default time format for the subsequence writes.
func (mlog *MultiLogger) SetTimeFormat(layout string) {
	mlog.timeFormat = layout
}

// UnregisterErrorWriter remove the error writer by name.
func (mlog *MultiLogger) UnregisterErrorWriter(name string) {
	delete(mlog.errs, name)
}

// UnregisterOutputWriter remove the output writer by name.
func (mlog *MultiLogger) UnregisterOutputWriter(name string) {
	delete(mlog.outs, name)
}

// Write write the b to all error writers.
// It will always return the length of b without an error.
func (mlog *MultiLogger) Write(b []byte) (n int, err error) {
	mlog.qerr <- libbytes.Copy(b)
	return len(b), nil
}

func (mlog *MultiLogger) processErrorQueue() {
	var (
		name string
		b    []byte
		w    NamedWriter
		err  error
		ok   bool
	)
	for {
		select {
		case b, ok = <-mlog.qerr:
			if !ok {
				// A closed channel is already empty, no need
				// to flush it.
				for name = range mlog.errs {
					delete(mlog.errs, name)
				}
				mlog.qerrFlush <- true
				return
			}

			if len(b) == 0 {
				b = append(b, '\n')
			} else if b[len(b)-1] != '\n' {
				b = append(b, '\n')
			}
			for name, w = range mlog.errs {
				_, err = w.Write(b)
				if err != nil {
					log.Printf("MultiLogger: %s: %s", name, err)
				}
			}

		case <-mlog.qerrFlush:
			flush(mlog.qerr, mlog.errs)
			mlog.qerrFlush <- true
		}
	}
}

func (mlog *MultiLogger) processOutputQueue() {
	var (
		name string
		b    []byte
		w    NamedWriter
		err  error
		ok   bool
	)

	for {
		select {
		case b, ok = <-mlog.qout:
			if !ok {
				// A closed channel is already empty, no need
				// to flush it.
				for name = range mlog.outs {
					delete(mlog.outs, name)
				}
				mlog.qoutFlush <- true
				return
			}

			if len(b) == 0 {
				b = append(b, '\n')
			} else if b[len(b)-1] != '\n' {
				b = append(b, '\n')
			}
			for name, w = range mlog.outs {
				_, err = w.Write(b)
				if err != nil {
					log.Printf("MultiLogger: %s: %s", name, err)
				}
			}

		case <-mlog.qoutFlush:
			flush(mlog.qout, mlog.outs)
			mlog.qoutFlush <- true
		}
	}
}

func (mlog *MultiLogger) writeTo(q chan []byte, format string, v ...interface{}) {
	mlog.Lock()
	if mlog.isClosed {
		mlog.Unlock()
		return
	}
	mlog.Unlock()

	var (
		buf    = mlog.bufPool.Get().(*bytes.Buffer)
		bufFmt = mlog.bufPool.Get().(*bytes.Buffer)
		args   = make([]interface{}, 0, len(v)+2)

		b []byte
	)
	buf.Reset()
	bufFmt.Reset()

	if len(mlog.timeFormat) > 0 {
		args = append(args, time.Now().UTC().Format(mlog.timeFormat))
		bufFmt.WriteString("%s ")
	}
	if len(mlog.prefix) > 0 {
		args = append(args, mlog.prefix)
		bufFmt.WriteString("%s ")
	}
	bufFmt.WriteString(format)
	args = append(args, v...)
	fmt.Fprintf(buf, bufFmt.String(), args...)

	b = libbytes.Copy(buf.Bytes())
	select {
	case q <- b:
	default:
		// Queue is full or closed.
	}

	mlog.bufPool.Put(bufFmt)
	mlog.bufPool.Put(buf)
}
