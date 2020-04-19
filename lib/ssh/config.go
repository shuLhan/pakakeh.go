// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssh

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	keyHost  = "host"
	keyMatch = "match"

	// List of valid keys in Host or Match section.
	keyAddKeysToAgent                  = "addkeystoagent"
	keyAddressFamily                   = "addressfamily"
	keyBatchMode                       = "batchmode"
	keyBindAddress                     = "bindaddress"
	keyBindInterface                   = "bindinterface"
	keyCanonicalDomains                = "canonicaldomains"
	keyCanonicalizeFallbackLocal       = "canonicalizefallbacklocal"
	keyCanonicalizeHostname            = "canonicalizehostname"
	keyCanonicalizeMaxDots             = "canonicalizemaxdots"
	keyCanonicalizePermittedCNAMEs     = "canonicalizepermittedcnames"
	keyCASignatureAlgorithms           = "casignaturealgorithms"
	keyCertificateFile                 = "certificatefile"
	keyChallengeResponseAuthentication = "challengeresponseauthentication"
	keyCheckHostIP                     = "checkhostip"
	keyClearAllForwardings             = "clearallforwardings"
	keyCompression                     = "compression"
	keyConnectionAttempts              = "connectionattempts"
	keyConnectTimeout                  = "connecttimeout"
	keyIdentityFile                    = "identityfile"
	keyHostname                        = "hostname"
	keyPort                            = "port"
	keyUser                            = "user"
	keyVisualHostKey                   = "visualhostkey"
	keyXAuthLocation                   = "xauthlocation"
)

//TODO: list of keys that is skipped due to time issues.
//nolint:varcheck,deadcode
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
	keyIdentityAgent                    = "identityagent"
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
	keySendEnv                          = "sendenv"
	keyServerAliveCountMax              = "serveralivecountmax"
	keyServerAliveInterval              = "serveraliveinterval"
	keySetEnv                           = "setenv"
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

var (
	errMultipleEqual = errors.New("multiple '=' character")
)

//
// Config contains mapping of host's patterns and its options from SSH
// configuration file.
//
type Config struct {
	sections []*ConfigSection
}

//
// NewConfig load SSH configuration from file.
//
func NewConfig(file string) (cfg *Config, err error) {
	if len(file) == 0 {
		return nil, nil
	}

	var (
		section *ConfigSection
		key     string
		value   string
	)

	cfg = &Config{
		sections: make([]*ConfigSection, 0),
	}

	parser, err := newConfigParser()
	if err != nil {
		return nil, fmt.Errorf("NewConfig %s: %w", file, err)
	}

	lines, err := parser.load("", file)
	if err != nil {
		return nil, fmt.Errorf("NewConfig %s: %w", file, err)
	}

	for x, line := range lines {
		if line[0] == '#' {
			continue
		}
		key, value, err = parseKeyValue(line)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", file, x, err)
		}

		switch key {
		case keyHost:
			if section != nil {
				section.postConfig(parser)
				cfg.sections = append(cfg.sections, section)
				section = nil
			}
			section, err = newSectionHost(value)
		case keyMatch:
			if section != nil {
				section.postConfig(parser)
				cfg.sections = append(cfg.sections, section)
				section = nil
			}
			section, err = newSectionMatch(value)
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
			section.CertificateFile = append(
				section.CertificateFile,
				value)

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

		case keyIdentityFile:
			if section.useDefaultIdentityFile {
				section.IdentityFile = []string{value}
				section.useDefaultIdentityFile = false
			} else {
				section.IdentityFile = append(
					section.IdentityFile, value)
			}

		case keyHostname:
			section.Hostname = value
		case keyPort:
			section.Port, err = strconv.Atoi(value)
		case keyUser:
			section.User = value
		case keyVisualHostKey:
			section.UseVisualHostKey, err = parseBool(key, value)
		case keyXAuthLocation:
			section.XAuthLocation = value
		}
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %w", file, x+1, err)
		}
	}
	if section != nil {
		section.postConfig(parser)
		cfg.sections = append(cfg.sections, section)
		section = nil
	}

	return cfg, nil
}

//
// Get the Host or Match configuration that match with the pattern "s".
//
func (cfg *Config) Get(s string) (section *ConfigSection) {
	for _, section := range cfg.sections {
		if section.isMatch(s) {
			return section
		}
	}
	return nil
}

func parseBool(key, val string) (out bool, err error) {
	switch val {
	case valueNo:
		return false, nil
	case valueYes:
		return true, nil
	}
	return false, fmt.Errorf("%s: invalid value %q", key, val)
}

//
// parseKeyValue from single line.
//
// ssh_config(5):
//
//	Configuration options may be separated by whitespace or optional
//	whitespace and exactly one `='; the latter format is useful to avoid
//	the need to quote whitespace ...
//
func parseKeyValue(line string) (key, value string, err error) {
	var (
		hasSeparator bool
		nequal       int
	)
	for y, r := range line {
		if r == ' ' || r == '=' {
			if r == '=' {
				nequal++
				if nequal > 1 {
					return key, value, errMultipleEqual
				}
			}
			if hasSeparator {
				continue
			}
			key = line[:y]
			hasSeparator = true
			continue
		}
		if !hasSeparator {
			continue
		}
		value = line[y:]
		break
	}
	key = strings.ToLower(key)
	value = strings.Trim(value, `"`)
	return key, value, nil
}
