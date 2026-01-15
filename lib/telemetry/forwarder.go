// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 Shulhan <ms@kilabit.info>

package telemetry

import "io"

// Forwarder provide the interface to be implemented by forwarder in order to
// store the collected metrics.
type Forwarder interface {
	// Implement the Close and Write from package [io].
	// Calling Forward after Close may cause panic.
	io.WriteCloser

	// Formatter return the Formatter being used to format the metrics.
	Formatter() Formatter
}
