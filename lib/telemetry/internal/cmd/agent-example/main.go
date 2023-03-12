// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Program agent-example provide an example of how to create agent.
package main

import (
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/shuLhan/share/lib/telemetry"
)

func main() {
	var (
		col       = createGoMemStatsCollector()
		ilpFmt    = telemetry.NewIlpFormatter(`rescached`)
		stdoutFwd = telemetry.NewStdoutForwarder(ilpFmt)
	)

	// Create metadata.
	var md = telemetry.NewMetadata()
	md.Set(`name`, `agent-example`)
	md.Set(`version`, `0.1.0`)

	// Create the Agent.
	var (
		agentOpts = telemetry.AgentOptions{
			Metadata:  md,
			Timestamp: telemetry.NanoTimestamp(),
			Forwarders: []telemetry.Forwarder{
				stdoutFwd,
			},
			Collectors: []telemetry.Collector{
				col,
			},
			Interval: 10 * time.Second,
		}
		agent = telemetry.NewAgent(agentOpts)
	)
	defer agent.Stop()

	var qsignal = make(chan os.Signal, 1)
	signal.Notify(qsignal, syscall.SIGQUIT, syscall.SIGSEGV, syscall.SIGTERM, syscall.SIGINT)
	<-qsignal
}

func createGoMetricsCollector() (col *telemetry.GoMetricsCollector) {
	var (
		metricsFilter *regexp.Regexp
		err           error
	)
	metricsFilter, err = regexp.Compile(`^go_(cpu|gc|memory|sched)_.*$`)
	if err != nil {
		log.Fatal(err)
	}
	col = telemetry.NewGoMetricsCollector(metricsFilter)
	return col
}

func createGoMemStatsCollector() (col *telemetry.GoMemStatsCollector) {
	var (
		metricsFilter *regexp.Regexp
		err           error
	)
	metricsFilter, err = regexp.Compile(`^.*$`)
	if err != nil {
		log.Fatal(err)
	}
	col = telemetry.NewGoMemStatsCollector(metricsFilter)
	return col
}
