// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package time

import (
	"fmt"
	"log"
	"time"
)

func ExampleScheduler_Next() {
	// Override timeNow to make this example works.
	timeNow = func() time.Time {
		return time.Date(2013, time.January, 20, 14, 26, 59, 0, time.UTC)
	}

	var (
		sch *Scheduler
		err error
	)

	sch, err = NewScheduler(`minutely`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sch.Next())
	// Output: 2013-01-20 14:27:00 +0000 UTC
}
