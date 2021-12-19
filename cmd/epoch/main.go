// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Program epoch print the current time (Unix seconds, milliseconds,
nanoseconds, local time, and UTC time) or the time based on the epoch on
first parameter.
Usage,

	epoch <unix-seconds|unix-milliseconds|unix-nanoseconds>

Without a parameter, it will print the current time.
With single parameter, it will print the time based on that epoch.

Example,

	$ epoch

	Unix seconds: 1639896843
	  Unix milli: 1639896843382
	  Unix micro: 1639896843382879
	   Unix nano: 1639896843382879358
	  Local time: 2021-12-19 13:54:03.382879358 +0700 WIB m=+0.000041439
	    UTC time: 2021-12-19 06:54:03.382879358 +0000 UTC


	$ epoch 1639800000

	Unix seconds: 1639800000
	  Unix milli: 1639800000000
	  Unix micro: 1639800000000000
	   Unix nano: 1639800000000000000
	  Local time: 2021-12-18 11:00:00 +0700 WIB
	    UTC time: 2021-12-18 04:00:00 +0000 UTC
*/
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"
)

func main() {
	flag.Parse()

	ts := flag.Arg(0)

	if len(ts) == 0 {
		echo(time.Now())
		return
	}

	var t time.Time

	epoch, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		log.Fatalf("invalid epoch %s: %s", ts, err)
	}

	switch len(ts) {
	case 10: // Epoch in seconds.
		t = time.Unix(epoch, 0)
	case 13:
		t = time.UnixMilli(epoch)
	case 16:
		t = time.UnixMicro(epoch)
	}

	echo(t)
}

func echo(t time.Time) {
	fmt.Printf(`
Unix seconds: %d
  Unix milli: %d
  Unix micro: %d
   Unix nano: %d
  Local time: %s
    UTC time: %s
`, t.Unix(), t.UnixMilli(), t.UnixMicro(), t.UnixNano(), t, t.UTC())

}
