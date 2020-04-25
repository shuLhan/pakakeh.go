// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	// Valid values for AddKeysToAgent.
	valueAsk     = "ask"
	valueConfirm = "confirm"
	valueNo      = "no"
	valueYes     = "yes"
	valueAlways  = "always"

	// Valid values for AddressFamily.
	valueAny   = "any"
	valueInet  = "inet"
	valueInet6 = "inet6"
)

const (
	envTerm = "TERM"
)

// List of default values.
const (
	defConnectionAttempts = 1
	defPort22             = 22
	defXAuthLocation      = "/usr/X11R6/bin/xauth"
)

//
// ConfigSection is the type that represent SSH client Host and Match section
// in configuration.
//
type ConfigSection struct {
	AddKeysToAgent              string
	AddressFamily               string
	BindAddress                 string
	BindInterface               string
	CanonicalDomains            []string
	CanonicalizeHostname        string
	CanonicalizeMaxDots         int
	CanonicalizePermittedCNAMEs *PermittedCNAMEs
	CASignatureAlgorithms       []string
	CertificateFile             []string
	ConnectionAttempts          int
	ConnectTimeout              int

	// Environments contains system environment variables that will be
	// passed to Execute().
	// The key and value is derived from "SendEnv" and "SetEnv".
	Environments map[string]string

	Hostname                          string
	IdentityFile                      []string
	Port                              int
	User                              string
	XAuthLocation                     string
	IsBatchMode                       bool
	IsCanonicalizeFallbackLocal       bool
	IsChallengeResponseAuthentication bool
	IsCheckHostIP                     bool
	IsClearAllForwardings             bool
	UseCompression                    bool
	UseVisualHostKey                  bool

	// Patterns for Host section.
	patterns []*configPattern

	// Criterias for Match section.
	criterias    []*matchCriteria
	useCriterias bool

	useDefaultIdentityFile bool // Flag for the IdentityFile.
}

// newConfigSection create new Host or Match with default values.
func newConfigSection() *ConfigSection {
	return &ConfigSection{
		AddKeysToAgent: valueNo,
		AddressFamily:  valueAny,
		CASignatureAlgorithms: []string{
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
			ssh.KeyAlgoRSA,
		},
		ConnectionAttempts: defConnectionAttempts,
		Environments: map[string]string{
			envTerm: os.Getenv(envTerm),
		},
		IdentityFile: []string{
			"~/.ssh/id_dsa",
			"~/.ssh/id_ecdsa",
			"~/.ssh/id_ed25519",
			"~/.ssh/id_rsa",
		},
		Port:                              defPort22,
		XAuthLocation:                     defXAuthLocation,
		useDefaultIdentityFile:            true,
		IsChallengeResponseAuthentication: true,
		IsCheckHostIP:                     true,
	}
}

//
// isMatch will return true if the string "s" match with one of Host or Match
// section.
//
func (section *ConfigSection) isMatch(s string) bool {
	if section.useCriterias {
		for _, criteria := range section.criterias {
			if criteria.isMatch(s) {
				return true
			}
		}
	} else {
		for _, pat := range section.patterns {
			if pat.isMatch(s) {
				return true
			}
		}
	}
	return false
}

//
// postConfig check, parse, and expand all of the fields values.
//
func (section *ConfigSection) postConfig(parser *configParser) {
	for x, identFile := range section.IdentityFile {
		if identFile[0] == '~' {
			section.IdentityFile[x] = strings.Replace(identFile,
				"~", parser.homeDir, 1)
		}
	}
}

func (section *ConfigSection) setAddKeysToAgent(val string) (err error) {
	switch val {
	case valueAsk, valueConfirm, valueNo, valueYes:
		section.AddKeysToAgent = val
	default:
		return fmt.Errorf("%s: invalid value %q", keyAddKeysToAgent,
			val)
	}
	return nil
}

func (section *ConfigSection) setAddressFamily(val string) (err error) {
	switch val {
	case valueAny, valueInet, valueInet6:
		section.AddressFamily = val
	default:
		return fmt.Errorf("%s: invalid value %q", keyAddressFamily,
			val)
	}
	return nil
}

func (section *ConfigSection) setCanonicalizeHostname(val string) (err error) {
	switch val {
	case valueNo, valueAlways, valueYes:
		section.CanonicalizeHostname = val
	default:
		return fmt.Errorf("%s: invalid value %q", keyBatchMode, val)
	}
	return nil
}

func (section *ConfigSection) setCanonicalizePermittedCNAMEs(val string) (err error) {
	sourceTarget := strings.Split(val, ":")
	if len(sourceTarget) != 2 {
		return fmt.Errorf("%s: invalid rule",
			keyCanonicalizePermittedCNAMEs)
	}

	listSource := strings.Split(sourceTarget[0], ",")
	sources := make([]*configPattern, 0, len(listSource))
	for _, domain := range listSource {
		src, err := newConfigPattern(domain)
		if err != nil {
			return fmt.Errorf("%s: invalid syntax %s",
				keyCanonicalizePermittedCNAMEs, domain)
		}
		sources = append(sources, src)
	}

	listTarget := strings.Split(sourceTarget[1], ",")
	targets := make([]*configPattern, 0, len(listTarget))
	for _, domain := range listTarget {
		target, err := newConfigPattern(domain)
		if err != nil {
			return fmt.Errorf("%s: invalid syntax %s",
				keyCanonicalizePermittedCNAMEs, domain)
		}
		targets = append(targets, target)
	}

	section.CanonicalizePermittedCNAMEs = &PermittedCNAMEs{
		sources: sources,
		targets: targets,
	}
	return nil
}

func (section *ConfigSection) setCASignatureAlgorithms(val string) {
	section.CASignatureAlgorithms = strings.Split(val, ",")
}

//
// setEnv set the Environments with key and value of format "KEY=VALUE".
//
func (section *ConfigSection) setEnv(env string) {
	kv := strings.SplitN(env, "=", 2)
	if len(kv) == 2 {
		section.Environments[kv[0]] = kv[1]
	}
}

func (section *ConfigSection) setSendEnv(envs map[string]string, pattern string) {
	for k, v := range envs {
		ok, _ := filepath.Match(pattern, k)
		if ok {
			section.Environments[k] = v
		}
	}
}
