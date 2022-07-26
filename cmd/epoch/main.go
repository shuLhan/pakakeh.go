// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Program epoch print the current date and time (Unix seconds, milliseconds,
// nanoseconds, local time, and UTC time) or the date and time based on the
// epoch on first parameter.
package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/shuLhan/share"
)

const usage = `= epoch

epoch print the current date and time (Unix seconds, milliseconds,
nanoseconds, local time, and UTC time) or the date and time based on the epoch
on first parameter.

== SYNOPSIS

	epoch <"help" |"version"|unix-seconds|unix-milliseconds|unix-nanoseconds>

== ARGUMENTS

Without a parameter, it will print the current time.
With single parameter, it will print the time based on that epoch.

help
	Print the current command usage (this message).

version
	Print the command version.


== EXAMPLE

Print the current date and time,

	$ epoch
	     Second: 1639896843
	Millisecond: 1639896843382
	Microsecond: 1639896843382879
	 Nanosecond: 1639896843382879358
	 Local time: 2021-12-19 13:54:03.382879358 +0700 WIB m=+0.000041439
	   UTC time: 2021-12-19 06:54:03.382879358 +0000 UTC

Print the date and time at epoch 1639800000,

	$ epoch 1639800000
	     Second: 1639800000
	Millisecond: 1639800000000
	Microsecond: 1639800000000000
	 Nanosecond: 1639800000000000000
	 Local time: 2021-12-18 11:00:00 +0700 WIB
	   UTC time: 2021-12-18 04:00:00 +0000 UTC`

const (
	cmdHelp    = "help"
	cmdVersion = "version"
)

func main() {
	var (
		t     time.Time
		cmd   string
		epoch int64
		err   error
	)

	log.SetFlags(0)
	log.SetPrefix("epoch: ")
	flag.Parse()

	cmd = strings.ToLower(flag.Arg(0))

	if len(cmd) == 0 {
		echo(time.Now())
		return
	}

	switch cmd {
	case cmdHelp:
		fmt.Println(usage)
		return

	case cmdVersion:
		fmt.Println("epoch v" + share.Version)
		return

	default:
		epoch, err = strconv.ParseInt(cmd, 10, 64)
		if err != nil {
			log.Fatalf("invalid epoch %s: %s", cmd, err)
		}

		switch len(cmd) {
		case 10: // Epoch in seconds.
			t = time.Unix(epoch, 0)
		case 13:
			t = time.UnixMilli(epoch)
		case 16:
			t = time.UnixMicro(epoch)
		}

		echo(t)
	}
}

func echo(t time.Time) {
	fmt.Printf(`     Second: %d
Millisecond: %d
Microsecond: %d
 Nanosecond: %d
 Local time: %s
   UTC time: %s
`, t.Unix(), t.UnixMilli(), t.UnixMicro(), t.UnixNano(), t, t.UTC())

}
