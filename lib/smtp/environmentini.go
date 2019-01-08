// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/shuLhan/share/lib/ini"
	libnet "github.com/shuLhan/share/lib/net"
)

const (
	defFileConfig = "smtpd.conf"
	keyCertPath   = "certificate_key"
	keyPrivPath   = "private_key"
	keyHostname   = "hostname"
	keyDomains    = "domains"
	secSMTPD      = "smtpd"
)

//
// EnvironmentIni load the SMTP environment from configuration file, with INI
// format.
// By default, the file is named "smtpd.conf".
//
// Config Format
//
// The configuration file have one root section "smtpd", which contains the
// following keys: hostname,domains.
//
// "hostname" is the primary domain name of this mail server.  If its empty,
// it will use the system hostname.
//
// "domains" key contains list of domain name that will be handled for
// incoming message. Each domain is separated by comma.
//
// "certificate_key" contains path to TLS certificate key.
//
// "private_key" contains path to server private key.
//
// Example,
//
//	[smtpd]
//	hostname = mail.local
//	domains = local.localdomain,localhost,...
//	certificate_key = /path/to/cert.pem
//	private_key = /path/to/key.pem
//
type EnvironmentIni struct {
	hostname string
	domains  []string
	cert     *tls.Certificate
}

//
// NewEnvironmentIni create and initialize environment from ini file.
// If file is empty its default to "smtpd.conf" in current directory.
//
func NewEnvironmentIni(file string) (env *EnvironmentIni, err error) {
	env = &EnvironmentIni{}

	if len(file) == 0 {
		file = defFileConfig
	}

	cfg, err := ini.Open(file)
	if err != nil {
		return nil, errors.New("NewEnvironmentIni: " + err.Error())
	}

	err = env.init(cfg)
	if err != nil {
		return nil, err
	}

	return env, nil
}

//
// Domains return list of domain names.
//
func (env *EnvironmentIni) Domains() []string {
	return env.domains
}

//
// Hostname return the primary domain name for mail server.
//
func (env *EnvironmentIni) Hostname() string {
	return env.hostname
}

//
// Certificate return the server certificate for TLS.
//
func (env *EnvironmentIni) Certificate() *tls.Certificate {
	return env.cert
}

func (env *EnvironmentIni) init(cfg *ini.Ini) (err error) {
	var ok bool

	env.hostname, ok = cfg.Get(secSMTPD, "", keyHostname)
	if !ok {
		env.hostname, err = os.Hostname()
		if err != nil {
			log.Println("EnvironmentIni.init: ", err.Error())
			return err
		}
	}

	env.hostname = strings.ToLower(strings.TrimSpace(env.hostname))

	if !libnet.IsHostnameValid([]byte(env.hostname)) {
		return fmt.Errorf("EnvironmentIni: invalid hostname '%s'",
			env.hostname)
	}

	v, ok := cfg.Get(secSMTPD, "", keyDomains)
	if !ok || v == "true" {
		env.domains = append(env.domains, env.hostname)
		return nil
	}

	mapDomains := make(map[string]bool)
	mapDomains[env.hostname] = true

	domains := strings.Split(v, ",")
	for _, name := range domains {
		name = strings.ToLower(strings.TrimSpace(name))
		if !libnet.IsHostnameValid([]byte(name)) {
			return fmt.Errorf("EnvironmentIni: invalid domain '%s'",
				name)
		}
		mapDomains[name] = true
	}

	for k := range mapDomains {
		env.domains = append(env.domains, k)
	}

	sort.Strings(env.domains)

	err = env.loadCertificate(cfg)
	if err != nil {
		return err
	}

	return nil
}

func (env *EnvironmentIni) loadCertificate(cfg *ini.Ini) (err error) {
	certPEM, err := env.loadCertKey(cfg)
	if err != nil {
		return err
	}

	keyPEM, err := env.loadPrivateKey(cfg)
	if err != nil {
		return err
	}
	if certPEM == nil || keyPEM == nil {
		return nil
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return errors.New("loadCertificate: " + err.Error())
	}

	env.cert = &cert

	return nil
}

func (env *EnvironmentIni) loadCertKey(cfg *ini.Ini) (b []byte, err error) {
	v, ok := cfg.Get(secSMTPD, "", keyCertPath)
	if !ok {
		return nil, nil
	}
	if len(v) == 0 || v == "true" {
		log.Println("loadCertKey: certificate path is empty")
		return nil, nil
	}

	b, err = ioutil.ReadFile(v)
	if err != nil {
		return nil, errors.New("loadCertKey: " + err.Error())
	}

	if len(b) == 0 {
		return nil, errors.New("loadCertKey: empty certificate")
	}

	return b, nil
}

func (env *EnvironmentIni) loadPrivateKey(cfg *ini.Ini) (b []byte, err error) {
	v, ok := cfg.Get(secSMTPD, "", keyPrivPath)
	if !ok {
		return nil, nil
	}
	if len(v) == 0 || v == "true" {
		log.Println("loadPrivateKey: private key path is empty")
		return nil, nil
	}

	b, err = ioutil.ReadFile(v)
	if err != nil {
		return nil, errors.New("loadPrivateKey: " + err.Error())
	}
	if len(b) == 0 {
		return nil, errors.New("loadPrivateKey: empty private key")
	}

	return b, nil
}
