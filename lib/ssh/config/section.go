// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
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

// List of default values.
const (
	defConnectionAttempts = 1
	defPort               = "22"
	defXAuthLocation      = "/usr/X11R6/bin/xauth"
)

// Section is the type that represent SSH client Host and Match section in
// configuration.
type Section struct {
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
	identityAgent                     string
	IdentityFile                      []string
	Port                              string
	User                              string
	XAuthLocation                     string
	IsBatchMode                       bool
	IsCanonicalizeFallbackLocal       bool
	IsChallengeResponseAuthentication bool
	IsCheckHostIP                     bool
	IsClearAllForwardings             bool
	UseCompression                    bool
	UseVisualHostKey                  bool

	// User's home directory.
	homeDir string

	// WorkingDir contains the directory where the SSH client started.
	// This value is required when client want to copy file from/to
	// remote.
	// This field is optional, default to current working directory from
	// os.Getwd() or user's home directory.
	WorkingDir string

	// The first IdentityFile that exist and valid.
	PrivateKeyFile string

	// PrivateKeys contains IdentityFile that has been parsed.
	// This field will be set once the Signers has been called.
	PrivateKeys map[string]any

	// Patterns for Host section.
	patterns []*pattern

	// Criteria for Match section.
	criteria    []*matchCriteria
	useCriteria bool

	useDefaultIdentityFile bool // Flag for the IdentityFile.
}

// newSection create new Host or Match with default values.
func newSection() *Section {
	return &Section{
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
		Environments:       map[string]string{},
		IdentityFile: []string{
			"~/.ssh/id_dsa",
			"~/.ssh/id_ecdsa",
			"~/.ssh/id_ed25519",
			"~/.ssh/id_rsa",
		},
		Port:                              defPort,
		XAuthLocation:                     defXAuthLocation,
		useDefaultIdentityFile:            true,
		IsChallengeResponseAuthentication: true,
		IsCheckHostIP:                     true,
	}
}

func newSectionHost(rawPattern string) (host *Section) {
	patterns := strings.Fields(rawPattern)

	host = newSection()
	host.patterns = make([]*pattern, 0, len(patterns))

	for _, pattern := range patterns {
		pat := newPattern(pattern)
		host.patterns = append(host.patterns, pat)
	}
	return host
}

// Signers convert the IdentityFile to ssh.Signer for authentication
// using PublicKey and store the parsed-unsigned private key into PrivateKeys.
//
// This method will ask for passphrase from terminal, if one of IdentityFile
// is protected.
// Unless the value of IdentityFile changes, this method should be called only
// once, otherwise it will ask passphrase on every call.
func (section *Section) Signers() (signers []ssh.Signer, err error) {
	var (
		logp = `Signers`

		pkeyFile      string
		pkeyPem       []byte
		pass          []byte
		signer        ssh.Signer
		pkey          any
		isMissingPass bool
	)

	section.PrivateKeys = make(map[string]any, len(section.IdentityFile))

	for _, pkeyFile = range section.IdentityFile {
		pkeyPem, err = os.ReadFile(pkeyFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		pkey, err = ssh.ParseRawPrivateKey(pkeyPem)
		if err != nil {
			_, isMissingPass = err.(*ssh.PassphraseMissingError)
			if !isMissingPass {
				return nil, fmt.Errorf(`%s: %w`, logp, err)
			}

			fmt.Printf("Enter passphrase for %s:", pkeyFile)

			pass, err = term.ReadPassword(0)
			if err != nil {
				return nil, fmt.Errorf(`%s: %w`, logp, err)
			}

			pkey, err = ssh.ParseRawPrivateKeyWithPassphrase(pkeyPem, pass)
			if err != nil {
				return nil, fmt.Errorf(`%s: %w`, logp, err)
			}
		}

		signer, err = ssh.NewSignerFromKey(pkey)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		if len(section.PrivateKeyFile) == 0 {
			section.PrivateKeyFile = pkeyFile
		}
		signers = append(signers, signer)
		section.PrivateKeys[pkeyFile] = pkey
	}
	return signers, nil
}

// GetIdentityAgent get the identity agent either from section config variable
// IdentityAgent or from environment variable SSH_AUTH_SOCK.
// It will return empty string if IdentityAgent set to "none" or SSH_AUTH_SOCK
// is empty.
func (section *Section) GetIdentityAgent() string {
	if section.identityAgent == "none" {
		return ""
	}
	if len(section.identityAgent) > 0 {
		return section.identityAgent
	}
	return os.Getenv(envSshAuthSock)
}

// isMatch will return true if the string "s" match with one of Host or Match
// section.
func (section *Section) isMatch(s string) bool {
	if section.useCriteria {
		for _, criteria := range section.criteria {
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

// init check, parse, and expand all of the fields values.
func (section *Section) init(workDir, homeDir string) {
	section.homeDir = homeDir
	section.WorkingDir = workDir

	for x, identFile := range section.IdentityFile {
		if identFile[0] == '~' {
			section.IdentityFile[x] = strings.Replace(identFile, "~", section.homeDir, 1)
		}
	}
}

func (section *Section) setAddKeysToAgent(val string) (err error) {
	switch val {
	case valueAsk, valueConfirm, valueNo, valueYes:
		section.AddKeysToAgent = val
	default:
		return fmt.Errorf("%s: invalid value %q", keyAddKeysToAgent,
			val)
	}
	return nil
}

func (section *Section) setAddressFamily(val string) (err error) {
	switch val {
	case valueAny, valueInet, valueInet6:
		section.AddressFamily = val
	default:
		return fmt.Errorf("%s: invalid value %q", keyAddressFamily,
			val)
	}
	return nil
}

func (section *Section) setCanonicalizeHostname(val string) (err error) {
	switch val {
	case valueNo, valueAlways, valueYes:
		section.CanonicalizeHostname = val
	default:
		return fmt.Errorf("%s: invalid value %q", keyBatchMode, val)
	}
	return nil
}

func (section *Section) setCanonicalizePermittedCNAMEs(val string) (err error) {
	sourceTarget := strings.Split(val, ":")
	if len(sourceTarget) != 2 {
		return fmt.Errorf("%s: invalid rule",
			keyCanonicalizePermittedCNAMEs)
	}

	listSource := strings.Split(sourceTarget[0], ",")
	sources := make([]*pattern, 0, len(listSource))
	for _, domain := range listSource {
		src := newPattern(domain)
		sources = append(sources, src)
	}

	listTarget := strings.Split(sourceTarget[1], ",")
	targets := make([]*pattern, 0, len(listTarget))
	for _, domain := range listTarget {
		target := newPattern(domain)
		targets = append(targets, target)
	}

	section.CanonicalizePermittedCNAMEs = &PermittedCNAMEs{
		sources: sources,
		targets: targets,
	}
	return nil
}

func (section *Section) setCASignatureAlgorithms(val string) {
	section.CASignatureAlgorithms = strings.Split(val, ",")
}

// setEnv set the Environments with key and value of format "KEY=VALUE".
func (section *Section) setEnv(env string) {
	kv := strings.SplitN(env, "=", 2)
	if len(kv) == 2 {
		section.Environments[kv[0]] = kv[1]
	}
}

// setIdentityAgent set the UNIX-domain socket used to communicate with
// the authentication agent.
// There are four possible value: SSH_AUTH_SOCK, <$STRING>, <PATH>, or
// "none".
// If SSH_AUTH_SOCK, the socket path is read from the environment variable
// SSH_AUTH_SOCK.
// If value start with "$", then the socket path is set based on value of that
// environment variable.
// Other string beside "none" will be considered as path to socket.
func (section *Section) setIdentityAgent(val string) {
	if val == envSshAuthSock {
		section.identityAgent = os.Getenv(envSshAuthSock)
		return
	}
	if val[0] == '$' {
		// Read the socket from environment variable defined by value.
		section.identityAgent = os.Getenv(val[1:])
		return
	}
	section.identityAgent = val
}

func (section *Section) setSendEnv(envs map[string]string, pattern string) {
	for k, v := range envs {
		ok, _ := filepath.Match(pattern, k)
		if ok {
			section.Environments[k] = v
		}
	}
}
