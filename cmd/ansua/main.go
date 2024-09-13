// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bufio"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

//go:embed README.md
var readme string

const (
	opHelp = `help`
)

const (
	stateCompleted = `completed`
	statePaused    = `paused`
	stateRunning   = `running`
)

const defTickerDuration = 20 * time.Second

func main() {
	flag.Parse()

	var param1 = flag.Arg(0)

	if len(param1) == 0 {
		fmt.Println(readme)
		os.Exit(1)
	}

	param1 = strings.ToLower(param1)
	if param1 == opHelp {
		fmt.Println(readme)
		os.Exit(0)
	}

	var (
		dur time.Duration
		err error
	)

	dur, err = time.ParseDuration(param1)
	if err != nil {
		log.Fatalf(`%s: %s`, os.Args[0], err)
	}

	var execArgs = getExecArg()

	fmt.Printf(`Running for %s`, dur)
	if len(execArgs) != 0 {
		fmt.Printf(` and then execute command %q`, execArgs)
	}
	fmt.Println(`.`)

	var (
		orgDur    = dur
		timeStart = time.Now().Round(time.Second)
		ticker    = time.NewTicker(defTickerDuration)
		timer     = time.NewTimer(dur)
		signalq   = make(chan os.Signal, 1)
		inputq    = make(chan byte, 1)
		state     = stateRunning

		pressed byte
	)

	signal.Notify(signalq, os.Interrupt, syscall.SIGTERM)

	go readKey(inputq)

	for state == stateRunning {
		select {
		case <-signalq:
			onStopped(`[Terminated]`, orgDur, timeStart)
			timer.Stop()
			os.Exit(0)

		case <-timer.C:
			state = stateCompleted
			dur = 0

		case <-ticker.C:
			dur -= defTickerDuration
			fmt.Printf("% 9s remaining...\n", dur)

		case pressed = <-inputq:
			if pressed == 'p' {
				onStopped(`[Paused]`, orgDur, timeStart)
				timer.Stop()
				ticker.Stop()
				state = statePaused
			}
		}
		for state == statePaused {
			select {
			case <-signalq:
				onStopped(`[Terminated]`, orgDur, timeStart)
				os.Exit(0)
			case pressed = <-inputq:
				if pressed == 'r' {
					fmt.Println(`[Resumed]`)
					timer = time.NewTimer(dur)
					ticker = time.NewTicker(defTickerDuration)
					state = stateRunning
				}
			}
		}
	}

	fmt.Println(`Time completed.`)

	if len(execArgs) == 0 {
		return
	}

	fmt.Println(`Executing command...`)
	run(signalq, execArgs)
}

func getExecArg() (execArgs string) {
	var args = flag.Args()[1:]
	if len(args) == 0 {
		// No command provided, exit immediately.
		return ``
	}

	return strings.Join(args, ` `)
}

func onStopped(cause string, orgDur time.Duration, timeStart time.Time) {
	var dur = orgDur - time.Since(timeStart).Round(time.Second)
	fmt.Printf("%s remaining duration is %s.\n", cause, dur)
}

func readKey(inputq chan byte) {
	var (
		in = bufio.NewReader(os.Stdin)

		err error
		c   byte
	)
	fmt.Println(`Press and enter [p] to pause, [r] to resume.`)
	for {
		c, err = in.ReadByte()
		if c == 0 {
			continue
		}
		if err != nil {
			log.Println(err)
		}
		if c == 'p' || c == 'r' {
			inputq <- c
		}
	}
}

func run(signalq chan os.Signal, execArgs string) {
	var (
		ctx       context.Context
		ctxCancel context.CancelFunc
	)

	ctx, ctxCancel = context.WithCancel(context.Background())

	var execCmd = exec.CommandContext(ctx, `/bin/sh`, `-c`, execArgs)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	var done = make(chan struct{}, 1)

	go func() {
		var err2 = execCmd.Run()
		if err2 != nil {
			log.Printf(`%s: %s`, os.Args[0], err2)
		}
		done <- struct{}{}
	}()

	select {
	case <-signalq:
		ctxCancel()
		os.Exit(0)
	case <-done:
	}
	ctxCancel()
}
