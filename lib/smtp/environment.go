// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

//
// Environment contains SMTP server environment.
//
type Environment struct {
	// PrimaryDomain of the SMTP server.
	// This field is required.
	PrimaryDomain *Domain

	// VirtualDomains contains list of virtual domain handled by server.
	// This field is optional.
	VirtualDomains map[string]*Domain
}
