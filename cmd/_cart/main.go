// Copyright 2016 Mhd Sulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/dsv"
	"github.com/shuLhan/share/lib/mining/classifier/cart"
	"github.com/shuLhan/share/lib/tabula"
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

	if debug.Value >= 1 {
		fmt.Printf("[cart] Class index: %v\n", dataset.GetClassIndex())
	}

	e = cartrt.Build(&dataset)
	if e != nil {
		panic(e)
	}

	if debug.Value >= 1 {
		fmt.Println("[cart] CART tree:\n", cartrt)
	}
}
