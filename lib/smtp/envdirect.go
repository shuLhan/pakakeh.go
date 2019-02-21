// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"crypto/tls"
	"fmt"
)

//
// EnvDirect provide a direct environment setting.
//
type EnvDirect struct {
	hostname string
	domains  []string
	cert     *tls.Certificate
}

// Certificate return the server certificate.
func (denv *EnvDirect) Certificate() *tls.Certificate {
	return denv.cert
}

// Domains return list of domains handled by mail server.
func (denv *EnvDirect) Domains() []string {
	return denv.domains
}

// Hostname return the hostname of mail server.
func (denv *EnvDirect) Hostname() string {
	return denv.hostname
}

//
// LoadCertificate load TLS certificate and its private key from file.
//
func (denv *EnvDirect) LoadCertificate(certFile, keyFile string) (err error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf("smtp: error at loading certificate: " + err.Error())
	}

	denv.cert = &cert

	return nil
}
