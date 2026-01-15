// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import "time"

// List of default value or limit for AgentOptions.
const (
	defInterval        = 1 * time.Minute
	defIntervalMinimum = 10 * time.Second
)

// AgentOptions contains options to run the Agent.
type AgentOptions struct {
	// Name of the agent.
	Name string

	// Metadata provides static, additional information to be forwarded
	// along with the collected metrics.
	Metadata *Metadata

	// Timestamp define the function to be called to set the
	// [Metric.Timestamp].
	// Default to NanoTimestamp.
	Timestamp Timestamper

	// Collectors contains list of Collector that provide the metrics to
	// be forwarded.
	// An empty Collectors means no metrics will be collected and
	// forwarded.
	Collectors []Collector

	// Forwarders contains list of target where collected metrics will be
	// forwarded.
	Forwarders []Forwarder

	// Interval for collecting metrics.
	// Default value is one minutes with the minimium value is 10 seconds.
	Interval time.Duration
}

// init initialize the AgentOptions default values.
func (opts *AgentOptions) init() {
	if opts.Metadata == nil {
		opts.Metadata = NewMetadata()
	}
	if opts.Timestamp == nil {
		opts.Timestamp = NanoTimestamp()
	}
	if opts.Interval <= 0 {
		opts.Interval = defInterval
	} else if opts.Interval < defIntervalMinimum {
		opts.Interval = defIntervalMinimum
	}
}
