// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry_test

import (
	"context"
	"fmt"
	"log"

	"git.sr.ht/~shulhan/pakakeh.go/lib/telemetry"
)

func ExampleBufferForwarder() {
	// Create the Formatter and Forwarder.
	var (
		dsvFmt = telemetry.NewDsvFormatter(';', telemetry.RuntimeMetricsAlias)
		bufFwd = telemetry.NewBufferForwarder(dsvFmt)
	)

	// Create metadata.
	var md = telemetry.NewMetadata()
	md.Set(`name`, `BufferForwarder`)
	md.Set(`version`, `0.1.0`)

	// Create the Agent.
	var (
		agentOpts = telemetry.AgentOptions{
			Metadata:   md,
			Forwarders: []telemetry.Forwarder{bufFwd},
			Timestamp:  telemetry.DummyTimestamp(),
		}
		agent = telemetry.NewAgent(agentOpts)
	)
	defer agent.Stop()

	// Forward single metric and print the result.
	var (
		m = telemetry.Metric{
			Name:  `usage`,
			Value: 0.5,
		}
		ctx = telemetry.ContextForwardWait(context.Background())

		err error
	)
	err = agent.Forward(ctx, m)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%s`, bufFwd.Bytes())

	// Forward list of Metric and print the result.
	var list = []telemetry.Metric{
		{Name: `usage`, Value: 0.4},
		{Name: `usage`, Value: 0.3},
	}
	err = agent.BulkForward(ctx, list)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`%s`, bufFwd.Bytes())
	// Output:
	// 1678606568;"usage";0.500000;"name=BufferForwarder,version=0.1.0"
	// 1678606568;"usage";0.400000;"name=BufferForwarder,version=0.1.0"
	// 1678606568;"usage";0.300000;"name=BufferForwarder,version=0.1.0"
}
