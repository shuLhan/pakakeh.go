// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/shuLhan/share/lib/ini"
	libnet "github.com/shuLhan/share/lib/net"
)

const (
	defFileConfig = "smtpd.conf"
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
// Example,
//
//	[smtpd]
//	hostname = mail.local
//	domains = local.localdomain,localhost,...
//
type EnvironmentIni struct {
	hostname string
	domains  []string
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

	return nil
}
