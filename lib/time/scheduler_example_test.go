package time_test

import (
	"fmt"
	"log"
	"time"

	libtime "github.com/shuLhan/share/lib/time"
)

func ExampleScheduler_Next() {
	// Override Now for making this example works.
	libtime.Now = func() time.Time {
		return time.Date(2013, time.January, 20, 14, 26, 59, 0, time.UTC)
	}

	var (
		sch *libtime.Scheduler
		err error
	)

	sch, err = libtime.NewScheduler(`minutely`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sch.Next())
	// Output: 2013-01-20 14:27:00 +0000 UTC
}
