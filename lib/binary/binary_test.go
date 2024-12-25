// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	var now = time.Date(2024, 12, 26, 2, 21, 0, 0, time.UTC)
	timeNow = func() time.Time {
		return now
	}
	os.Exit(m.Run())
}
