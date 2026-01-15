// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 Shulhan <ms@kilabit.info>

package smtp

// Environment contains SMTP server environment.
type Environment struct {
	// PrimaryDomain of the SMTP server.
	// This field is required.
	PrimaryDomain *Domain

	// VirtualDomains contains list of virtual domain handled by server.
	// This field is optional.
	VirtualDomains map[string]*Domain
}
