// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

// Metric contains name, value, and timestamp.
type Metric struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
}
