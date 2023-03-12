// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry

// Collector provides an interface to collect the metrics.
type Collector interface {
	// Collect the metrics with timestamp.
	Collect(timestamp int64) []Metric
}
