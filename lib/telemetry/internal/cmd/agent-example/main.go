// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Program agent-example provide an example of how to create agent.
package main

import (
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/telemetry"
)

func main() {
	var (
		colGoMemstats = createGoMemStatsCollector()
		colGoMetrics  = createGoMetricsCollector()
		ilpFmt        = telemetry.NewIlpFormatter(`rescached`)
		stdoutFwd     = telemetry.NewStdoutForwarder(ilpFmt)
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
				colGoMemstats,
				colGoMetrics,
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
	var metricsFilter = regexp.MustCompile(`^go_(cpu|gc|memory|sched)_.*$`)
	col = telemetry.NewGoMetricsCollector(metricsFilter)
	return col
}

func createGoMemStatsCollector() (col *telemetry.GoMemStatsCollector) {
	var metricsFilter = regexp.MustCompile(`^.*$`)
	col = telemetry.NewGoMemStatsCollector(metricsFilter)
	return col
}
