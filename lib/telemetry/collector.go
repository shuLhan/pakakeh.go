// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

// Collector provides an interface to collect the metrics.
type Collector interface {
	// Collect the metrics with timestamp.
	Collect(timestamp int64) []Metric
}
