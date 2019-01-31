// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debug

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

var prof *profile // nolint: gochecknoglobals

type profile struct {
	data *pprof.Profile
	path string
	file *os.File
}

func newProfile(prefix string) *profile {
	if prof == nil {
		prof = &profile{}
	}
	if len(prof.path) == 0 {
		prof.path = fmt.Sprintf("/tmp/%s.%d.heap.pprof", prefix,
			os.Getpid())
	}
	if prof.file == nil {
		var err error
		prof.file, err = os.Create(prof.path)
		if err != nil {
			log.Println("lib/debug: could not create memory profile: ", err)
		}
	}

	return prof
}

func (prof *profile) reset() {
	if prof.file != nil {
		_ = prof.file.Close()
	}
	prof.file = nil
	prof.path = ""
}

func (prof *profile) writeHeap() (err error) {
	if prof.file == nil {
		return nil
	}

	err = prof.file.Truncate(0)
	if err != nil {
		log.Println("lib/debug: error truncating memory profile: ", err)
		return err
	}

	_, err = prof.file.Seek(0, 0)
	if err != nil {
		log.Println("lib/debug: error at seek memory profile: ", err)
		return err
	}

	runtime.GC() // get up-to-date statistics.

	prof.data = pprof.Lookup("heap")

	err = prof.data.WriteTo(prof.file, 0)
	if err != nil {
		log.Println("lib/debug: could not write memory profile: ", err)
		return err
	}

	prof.data = nil

	return nil
}

//
// WriteHeapProfile write memory profile into "/tmp/{prefix}.pid.heap.pprof".
// If keepAlive is true, the file will be keep opened until error happened, or
// caller send keepAlive=false, or when program end.
//
func WriteHeapProfile(prefix string, keepAlive bool) {
	prof = newProfile(prefix)

	err := prof.writeHeap()
	if err != nil {
		keepAlive = false
	}

	if !keepAlive {
		prof.reset()
	}
}
