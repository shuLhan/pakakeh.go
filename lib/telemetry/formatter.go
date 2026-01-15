// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

// Formatter define the interface that responsible to convert single or bulk
// of Metric into its wire format.
type Formatter interface {
	// BulkFormat format list of Metric with metadata for transfer.
	BulkFormat(listm []Metric, md *Metadata) []byte

	// Format the Metric m and metadata for transfer.
	Format(m Metric, md *Metadata) []byte

	// Name return the name of formatter.
	Name() string
}
