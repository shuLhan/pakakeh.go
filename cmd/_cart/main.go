// SPDX-FileCopyrightText: 2016 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/dsv"
	"git.sr.ht/~shulhan/pakakeh.go/lib/mining/classifier/cart"
	"git.sr.ht/~shulhan/pakakeh.go/lib/tabula"
)

type options struct {
	nRandomFeature int
}

func usage() {
	cmd := os.Args[0]
	fmt.Fprintf(os.Stderr, "Usage of %s: [-n number] [config.dsv]\n", cmd)
	flag.PrintDefaults()
}

func initFlags() (opts options) {
	flagUsage := []string{
		"Number of random feature (default 0)",
	}

	flag.IntVar(&opts.nRandomFeature, "n", 0, flagUsage[0])

	flag.Parse()

	return opts
}

func trace(s string) (string, time.Time) {
	fmt.Println("[START]", s)
	return s, time.Now()
}

func un(s string, startTime time.Time) {
	endTime := time.Now()
	fmt.Println("[END]", s, "with elapsed time",
		endTime.Sub(startTime))
}

func createCart(fcfg string, opts *options) (*cart.Runtime, error) {
	cartrt := &cart.Runtime{}

	config, e := os.ReadFile(fcfg)
	if e != nil {
		return nil, e
	}

	e = json.Unmarshal(config, cartrt)
	if e != nil {
		return nil, e
	}

	if opts.nRandomFeature > 0 {
		cartrt.NRandomFeature = opts.nRandomFeature
	}

	return cartrt, nil
}

func main() {
	defer un(trace("cart"))

	opts := initFlags()

	if len(flag.Args()) == 0 {
		usage()
		os.Exit(1)
	}

	fcfg := flag.Arg(0)

	// Parsing config file and check command parameter values.
	cartrt, e := createCart(fcfg, &opts)
	if e != nil {
		panic(e)
	}

	// Get dataset
	dataset := tabula.Claset{}
	_, e = dsv.SimpleRead(fcfg, &dataset)
	if e != nil {
		panic(e)
	}

	e = cartrt.Build(&dataset)
	if e != nil {
		panic(e)
	}
}
