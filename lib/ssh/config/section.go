// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// List of valid keys in Host or Match section.
const (
	keyAddKeysToAgent                  = "addkeystoagent"
	keyAddressFamily                   = "addressfamily"
	keyBatchMode                       = "batchmode"
	keyBindAddress                     = "bindaddress"
	keyBindInterface                   = "bindinterface"
	keyCASignatureAlgorithms           = "casignaturealgorithms"
	keyCanonicalDomains                = "canonicaldomains"
	keyCanonicalizeFallbackLocal       = "canonicalizefallbacklocal"
	keyCanonicalizeHostname            = "canonicalizehostname"
	keyCanonicalizeMaxDots             = "canonicalizemaxdots"
	keyCanonicalizePermittedCNAMEs     = "canonicalizepermittedcnames"
	keyCertificateFile                 = "certificatefile"
	keyChallengeResponseAuthentication = "challengeresponseauthentication"
	keyCheckHostIP                     = "checkhostip"
	keyClearAllForwardings             = "clearallforwardings"
	keyCompression                     = "compression"
	keyConnectTimeout                  = "connecttimeout"
	keyConnectionAttempts              = "connectionattempts"
	keyHostname                        = "hostname"
	keyIdentityAgent                   = "identityagent"
	keyIdentityFile                    = "identityfile"
	keyPort                            = "port"
	keySendEnv                         = "sendenv"
	keySetEnv                          = "setenv"
	keyUser                            = "user"
	keyVisualHostKey                   = "visualhostkey"
	keyXAuthLocation                   = "xauthlocation"
)

// TODO: list of keys that are not implemented yet due to hard or
// unknown how to test it.
// nolint: deadcode,varcheck
const (
	keyCiphers                          = "ciphers"
	keyControlMaster                    = "controlmaster"
	keyControlPath                      = "controlpath"
	keyControlPersist                   = "controlpersist"
	keyDynamicForward                   = "dynamicforward"
	keyEnableSSHKeysign                 = "enablesshkeysign"
	keyEscapeChar                       = "escapechar"
	keyExitOnForwardFailure             = "keyexitonforwardfailure"
	keyFingerprintHash                  = "fingerprinthash"
	keyForwardAgent                     = "forwardagent"
	keyForwardX11                       = "forwardx11"
	keyForwardX11Timeout                = "forwardx11timeout"
	keyForwardX11Trusted                = "forwardx11trusted"
	keyGatewayPorts                     = "gatewayports"
	keyGlobalKnownHostsFile             = "globalknownhostsfile"
	keyGSSAPIAuthentication             = "gssapiauthentication"
	keyGSSAPIDelegateCredentials        = "gssapidelegatecredentials"
	keyHashKnownHosts                   = "hashknownhosts"
	keyHostBasedAuthentication          = "hostbasedauthentication"
	keyHostBaseKeyTypes                 = "hostbasedkeytypes"
	keyHostKeyAlgorithms                = "hostkeyalgorithms"
	keyHostKeyAlias                     = "hostkeyalias"
	keyIdentitiesOnly                   = "identitiesonly"
	keyIgnoreUnknown                    = "ignoreunknown"
	keyInclude                          = "include"
	keyIPQoS                            = "ipqos"
	keyKbdInteractiveAuthentication     = "kbdinteractiveauthentication"
	keyKbdInteractiveDevices            = "kbdinteractivedevices"
	keyKexAlgorithms                    = "kexalgorithms"
	keyLocalCommand                     = "localcommand"
	keyLocalForward                     = "localforward"
	keyLogLevel                         = "loglevel"
	keyMACs                             = "macs"
	keyNoHostAuthenticationForLocalhost = "nohostauthenticationforlocalhost"
	keyNumberOfPasswordPrompts          = "numberofpasswordprompts"
	keyPasswordAuthentication           = "passwordauthentication"
	keyPermitLocalCommand               = "permitlocalcommand"
	keyPKCS11Provider                   = "pkcs11provider"
	keyPreferredAuthentications         = "preferredauthentications"
	keyProxyCommand                     = "proxycommand"
	keyProxyJump                        = "proxyjump"
	keyProxyUseFdpass                   = "proxyusefdpass"
	keyPubkeyAcceptedKeyTypes           = "pubkeyacceptedkeytypes"
	keyPubkeyAuthentication             = "pubkeyauthentication"
	keyRekeyLimit                       = "rekeylimit"
	keyRemoteCommand                    = "remotecommand"
	keyRemoteForward                    = "remoteforward"
	keyRequestTTY                       = "requesttty"
	keyRevokeHostKeys                   = "revokehostkeys"
	keyServerAliveCountMax              = "serveralivecountmax"
	keyServerAliveInterval              = "serveraliveinterval"
	keyStreamLocalBindMask              = "streamlocalbindmask"
	keyStreamLocalBindUnlink            = "streamlocalbindunlink"
	keyStrictHostKeyChecking            = "stricthostkeychecking"
	keySyslogFacility                   = "syslogfacility"
	keyTCPKeepAlive                     = "tcpkeepalive"
	keyTunnel                           = "tunnel"
	keyTunnelDevince                    = "tunneldevice"
	keyUpdatehostKeys                   = "updatehostkeys"
	keyUseKeychain                      = "usekeychain"
	keyUserKnownHostsFile               = "userknownhostsfile"
	keyVerifyHostKeyDNS                 = "verifyhostkeydns"
)

// Valid values for AddKeysToAgent.
const (
	valueAsk     = "ask"
	valueConfirm = "confirm"
	valueNo      = "no"
	valueYes     = "yes"
	valueAlways  = "always"
)

// Valid values for AddressFamily.
const (
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
	// Environments contains system environment variables that will be
	// passed to Execute().
	// The key and value is derived from "SendEnv" and "SetEnv".
	Environments map[string]string

	// PrivateKeys contains IdentityFile that has been parsed.
	// This field will be set once the Signers has been called.
	PrivateKeys map[string]any

	// Field store the unpacked key and value of Section.
	Field map[string]string

	CanonicalizePermittedCNAMEs *PermittedCNAMEs

	// name contains the raw value after Host or Match.
	name string

	AddKeysToAgent       string
	AddressFamily        string
	BindAddress          string
	BindInterface        string
	CanonicalizeHostname string

	Hostname string
	Port     string

	// The first IdentityFile that exist and valid.
	PrivateKeyFile string

	User string

	// WorkingDir contains the directory where the SSH client started.
	// This value is required when client want to copy file from/to
	// remote.
	// This field is optional, default to current working directory from
	// os.Getwd() or user's home directory.
	WorkingDir string

	XAuthLocation string

	// User's home directory.
	homeDir string

	identityAgent string

	CanonicalDomains      []string
	CASignatureAlgorithms []string
	CertificateFile       []string
	IdentityFile          []string

	// Patterns for Host section.
	patterns []*pattern

	// Criteria for Match section.
	criteria []*matchCriteria

	CanonicalizeMaxDots int
	ConnectionAttempts  int
	ConnectTimeout      int

	IsBatchMode                       bool
	IsCanonicalizeFallbackLocal       bool
	IsChallengeResponseAuthentication bool
	IsCheckHostIP                     bool
	IsClearAllForwardings             bool
	UseCompression                    bool
	UseVisualHostKey                  bool

	useCriteria bool

	useDefaultIdentityFile bool // Flag for the IdentityFile.
}

// newSection create new Host or Match with default values.
func newSection(name string) *Section {
	return &Section{
		Environments: map[string]string{},
		Field:        map[string]string{},

		name:           name,
		AddKeysToAgent: valueNo,
		AddressFamily:  valueAny,
		Port:           defPort,
		XAuthLocation:  defXAuthLocation,

		CASignatureAlgorithms: []string{
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
			ssh.KeyAlgoRSA,
		},

		IdentityFile: []string{
			"~/.ssh/id_dsa",
			"~/.ssh/id_ecdsa",
			"~/.ssh/id_ed25519",
			"~/.ssh/id_rsa",
		},

		ConnectionAttempts: defConnectionAttempts,

		useDefaultIdentityFile:            true,
		IsChallengeResponseAuthentication: true,
		IsCheckHostIP:                     true,
	}
}

func newSectionHost(rawPattern string) (host *Section) {
	patterns := strings.Fields(rawPattern)

	host = newSection(rawPattern)
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

// mergeField merge the Field from other Section.
func (section *Section) mergeField(cfg *Config, other *Section) {
	var (
		key   string
		value string
	)
	for key, value = range other.Field {
		// The key and value in other should be valid, so no need to
		// check for error.
		_ = section.set(cfg, key, value)
	}
}

// set the section field by raw key and value.
func (section *Section) set(cfg *Config, key, value string) (err error) {
	switch key {
	case keyAddKeysToAgent:
		err = section.setAddKeysToAgent(value)
	case keyAddressFamily:
		err = section.setAddressFamily(value)
	case keyBatchMode:
		section.IsBatchMode, err = parseBool(key, value)
	case keyBindAddress:
		section.BindAddress = value
	case keyBindInterface:
		section.BindInterface = value
	case keyCanonicalDomains:
		section.CanonicalDomains = strings.Fields(value)
	case keyCanonicalizeFallbackLocal:
		section.IsCanonicalizeFallbackLocal, err = parseBool(key, value)
	case keyCanonicalizeHostname:
		err = section.setCanonicalizeHostname(value)
	case keyCanonicalizeMaxDots:
		section.CanonicalizeMaxDots, err = strconv.Atoi(value)

	case keyCanonicalizePermittedCNAMEs:
		err = section.setCanonicalizePermittedCNAMEs(value)

	case keyCASignatureAlgorithms:
		section.setCASignatureAlgorithms(value)

	case keyCertificateFile:
		section.CertificateFile = append(section.CertificateFile, value)

	case keyChallengeResponseAuthentication:
		section.IsChallengeResponseAuthentication, err = parseBool(key, value)

	case keyCheckHostIP:
		section.IsCheckHostIP, err = parseBool(key, value)

	case keyClearAllForwardings:
		section.IsClearAllForwardings, err = parseBool(key, value)
	case keyCompression:
		section.UseCompression, err = parseBool(key, value)
	case keyConnectionAttempts:
		section.ConnectionAttempts, err = strconv.Atoi(value)
	case keyConnectTimeout:
		section.ConnectTimeout, err = strconv.Atoi(value)

	case keyIdentityAgent:
		section.setIdentityAgent(value)

	case keyIdentityFile:
		if section.useDefaultIdentityFile {
			section.IdentityFile = []string{value}
			section.useDefaultIdentityFile = false
		} else {
			section.IdentityFile = append(section.IdentityFile, value)
		}

	case keyHostname:
		section.Hostname = value
	case keyPort:
		section.Port = value
	case keySendEnv:
		section.setSendEnv(cfg.envs, value)
	case keySetEnv:
		section.setEnv(value)
	case keyUser:
		section.User = value
	case keyVisualHostKey:
		section.UseVisualHostKey, err = parseBool(key, value)
	case keyXAuthLocation:
		section.XAuthLocation = value
	default:
		// Store the unknown key into Field.
	}
	if err != nil {
		return err
	}
	section.Field[key] = value
	return nil
}

func (section *Section) setAddKeysToAgent(val string) (err error) {
	switch val {
	case valueAsk, valueConfirm, valueNo, valueYes:
		section.AddKeysToAgent = val
	default:
		return fmt.Errorf("%s: invalid value %q", keyAddKeysToAgent, val)
	}
	return nil
}

func (section *Section) setAddressFamily(val string) (err error) {
	switch val {
	case valueAny, valueInet, valueInet6:
		section.AddressFamily = val
	default:
		return fmt.Errorf("%s: invalid value %q", keyAddressFamily, val)
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
		return fmt.Errorf("%s: invalid rule", keyCanonicalizePermittedCNAMEs)
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
