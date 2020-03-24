// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package debug

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
)

//
// CPUProfile provide a wrapper to starting and stopping CPU profiler from
// package "runtime/pprof".
//
type CPUProfile struct {
	f *os.File
}

//
// NewCPUProfile create and start the CPU profiler.
// On success, it will return the running profiler; otherwise it will return
// nil.
// Do not forget to call Stop() when finished.
//
func NewCPUProfile(prefix string) (prof *CPUProfile) {
	var err error

	path := fmt.Sprintf("/tmp/%s.%d.cpu.pprof", prefix, os.Getpid())
	prof = &CPUProfile{}
	prof.f, err = os.Create(path)
	if err != nil {
		log.Println(prefix, ": NewCPUProfile:", err)
		return nil
	}
	err = pprof.StartCPUProfile(prof.f)
	if err != nil {
		prof.Stop()
		log.Println(prefix, ": NewCPUProfile:", err)
		return nil
	}
	return prof
}

//
// Stop the CPU profiler.
//
func (prof *CPUProfile) Stop() {
	pprof.StopCPUProfile()

	if prof == nil {
		return
	}
	if prof.f != nil {
		err := prof.f.Close()
		if err != nil {
			log.Println("CPUProfile.Stop: ", err)
		}
	}
}
