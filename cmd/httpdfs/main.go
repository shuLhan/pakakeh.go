// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Program httpdfs implement [libhttp.Server] with [memfs.MemFS].
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	pakakeh "git.sr.ht/~shulhan/pakakeh.go"
	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
	"git.sr.ht/~shulhan/pakakeh.go/lib/memfs"
	"git.sr.ht/~shulhan/pakakeh.go/lib/systemd"
)

const (
	defAddress = `127.0.0.1:28194`
)

func main() {
	var serverOpts = libhttp.ServerOptions{
		EnableIndexHTML: true,
	}

	var (
		flagExclude string
		flagInclude string

		flagHelp    bool
		flagVersion bool
	)
	var shutdownIdleDuration string

	flag.StringVar(&serverOpts.Address, `address`, defAddress,
		`Set listen address.`)
	flag.StringVar(&serverOpts.BasePath, `base-path`, ``,
		`Set the base path (or prefix) for handling HTTP request.`)
	flag.StringVar(&shutdownIdleDuration, `shutdown-idle`, ``,
		`Set the duration when server will shutting down after idle.`)

	flag.StringVar(&flagExclude, `exclude`, ``, `Regex to exclude files in base directory`)
	flag.BoolVar(&flagHelp, `help`, false, `Print the command usage`)
	flag.StringVar(&flagInclude, `include`, ``, `Regex to include files in base directory`)
	flag.BoolVar(&flagVersion, `version`, false, `Print the program version`)

	flag.Parse()

	var err error
	if shutdownIdleDuration != `` {
		serverOpts.ShutdownIdleDuration, err = time.ParseDuration(shutdownIdleDuration)
		if err != nil {
			log.Fatalf(`invalid shutdown-idle %s: %s`, shutdownIdleDuration, err)
		}
	}

	var cmdName = os.Args[0]

	if flagHelp {
		usage(cmdName)
		os.Exit(0)
	}
	if flagVersion {
		fmt.Println(pakakeh.Version)
		os.Exit(0)
	}

	var dirBase = flag.Arg(0)
	if len(dirBase) == 0 {
		dirBase, err = os.Getwd()
		if err != nil {
			log.Fatalf(`%s: %s`, cmdName, err)
		}
	}

	var (
		mfsOpts = memfs.Options{
			Root:        dirBase,
			MaxFileSize: -1,
			TryDirect:   true,
		}
	)
	if len(flagInclude) != 0 {
		mfsOpts.Includes = []string{flagInclude}
	}
	if len(flagExclude) != 0 {
		mfsOpts.Excludes = []string{flagExclude}
	}

	serverOpts.Memfs, err = memfs.New(&mfsOpts)
	if err != nil {
		log.Fatalf(`%s: %s`, cmdName, err)
	}

	listeners, err := systemd.Listeners(true)
	if err != nil {
		log.Fatal(err)
	}
	if len(listeners) > 1 {
		log.Fatal(`too many listeners received for activation`)
	}
	if len(listeners) == 1 {
		serverOpts.Listener = listeners[0]
		gotAddr := serverOpts.Listener.Addr().String()
		if gotAddr != serverOpts.Address {
			log.Fatalf(`invalid Listener address, got %s, want %s`,
				gotAddr, serverOpts.Address)
		}
	}

	var httpd *libhttp.Server

	httpd, err = libhttp.NewServer(serverOpts)
	if err != nil {
		log.Fatalf(`%s: %s`, cmdName, err)
	}

	var signalq = make(chan os.Signal, 1)
	signal.Notify(signalq, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf(`%s: serving %q at http://%s`, cmdName, dirBase, serverOpts.Address)
		var errStart = httpd.Start()
		if errStart != nil {
			log.Printf(`%s: %s`, cmdName, errStart)
		}
	}()

	<-signalq

	err = httpd.Stop(0)
	if err != nil {
		log.Printf(`%s: %s`, cmdName, err)
	}
}

func usage(cmdName string) {
	fmt.Println(`= ` + cmdName + ` - a simple HTTP server

	` + cmdName + ` [options] <dir>

== Options

	-address <IP:PORT>
		Run the HTTP server on specific IP address and port.
		Default to ` + defAddress + `.

	-exclude <regex>
		Exclude the files matched by regex from being served.
		Default to empty, none of files is excluded.

	-help
		Print this usage.

	-include <regex>
		Serve only list of files matched with regex.
		Default to include CSS, HTML, JavaScript, ICO, JPG, PNG, and
		SVG files only.

	-version
		Print the program version.

== Parameter

	<dir>
		Directory to be served under HTTP.
		If not set default to current directory.`)
}
