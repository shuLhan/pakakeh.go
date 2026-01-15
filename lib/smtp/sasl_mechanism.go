// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2019 Shulhan <ms@kilabit.info>

package smtp

// SaslMechanism represent Simple Authentication and Security Layer (SASL)
// mechanism (RFC 4422).
type SaslMechanism int

// List of available SASL mechanism.
const (
	SaslMechanismPlain SaslMechanism = 1
)
