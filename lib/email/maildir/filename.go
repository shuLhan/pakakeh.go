// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maildir

import (
	"fmt"
	"syscall"
)

// filename contains information about the format in file name.
type filename struct {
	nameTmp string // The name of file for the "tmp" directory.
	nameNew string // The name of file for the "new" directory.

	hostname string

	epoch   int64
	counter int64
	usecond int
	pid     int
}

// createFilename create and initialize "tmp" file name.
func createFilename(pid int, counter int64, hostname string) (fname filename) {
	var now = timeNow()

	fname = filename{
		epoch:    now.Unix(),
		usecond:  now.Nanosecond() / 1000,
		pid:      pid,
		counter:  counter,
		hostname: hostname,
	}
	fname.nameTmp = fmt.Sprintf(`%d.M%d_P%d_Q%d.%s`, fname.epoch,
		fname.usecond, fname.pid, fname.counter,
		fname.hostname)
	return fname
}

// generateNameNew generate file name for "new" based on file stat of pathTmp.
// It will return empty string if it cannot call Stat on pathTmp.
func (fname *filename) generateNameNew(pathTmp string, size int64) (nameNew string, err error) {
	var (
		logp = `generateNameNew`
		stat syscall.Stat_t
	)

	err = syscallStat(pathTmp, &stat)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	if size == 0 {
		size = stat.Size
	}

	fname.nameNew = fmt.Sprintf(`%d.M%d_P%d_V%d_I%d_Q%d.%s,S=%d:2`,
		fname.epoch, fname.usecond, fname.pid, stat.Dev, stat.Ino,
		fname.counter, fname.hostname, size)

	return fname.nameNew, nil
}
