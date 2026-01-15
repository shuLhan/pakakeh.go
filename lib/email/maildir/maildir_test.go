// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package maildir

import (
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var epoch int64 = 1684640949 // 2023-05-21 03:49:09 +0000 UTC

	osGetpid = func() int {
		return 1000
	}

	osHostname = func() (string, error) {
		return `localhost`, nil
	}

	syscallStat = func(path string, stat *syscall.Stat_t) error {
		path = strings.TrimSpace(path)
		if len(path) == 0 {
			return os.ErrNotExist
		}
		stat.Dev = 36
		stat.Ino = 170430
		stat.Size = 15
		return nil
	}

	timeNow = func() time.Time {
		return time.Unix(epoch, 875494837)
	}

	os.Exit(m.Run())
}
