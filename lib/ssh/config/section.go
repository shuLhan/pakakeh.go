// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// List of valid keys in Host or Match section.
const (
	// List of key in Host or Match with single, string value.
	KeyAddKeysToAgent       = `addkeystoagent`
	KeyAddressFamily        = `addressfamily`
	KeyBindAddress          = `bindaddress`
	KeyBindInterface        = `bindinterface`
	KeyCanonicalizeHostname = `canonicalizehostname`
	KeySetEnv               = `setenv`
	KeyXAuthLocation        = `xauthlocation`

	// List of key in Host or Match with multiple, string values.
	KeyCASignatureAlgorithms = `casignaturealgorithms`
	KeyCanonicalDomains      = `canonicaldomains`
	KeyCertificateFile       = `certificatefile`
	KeyIdentityFile          = `identityfile`
	KeySendEnv               = `sendenv`
	KeyUserKnownHostsFile    = `userknownhostsfile`

	// List of key in Host or Match with integer value.
	KeyCanonicalizeMaxDots = `canonicalizemaxdots`
	KeyConnectTimeout      = `connecttimeout`
	KeyConnectionAttempts  = `connectionattempts`

	// List of key in Host or Match with boolean value.
	KeyBatchMode                       = `batchmode`
	KeyCanonicalizeFallbackLocal       = `canonicalizefallbacklocal`
	KeyChallengeResponseAuthentication = `challengeresponseauthentication`
	KeyCheckHostIP                     = `checkhostip`
	KeyClearAllForwardings             = `clearallforwardings`
	KeyCompression                     = `compression`
	KeyVisualHostKey                   = `visualhostkey`

	// List of key in Host or Match with value fetched using method.
	KeyCanonicalizePermittedCNames = `canonicalizepermittedcnames`
	KeyHostname                    = `hostname`
	KeyIdentityAgent               = `identityagent`
	KeyPort                        = `port`
	KeyUser                        = `user`
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
	keyVerifyHostKeyDNS                 = "verifyhostkeydns"
)

// Known values for key.
const (
	ValueAcceptNew = `accept-new`
	ValueAlways    = `always`
	ValueAsk       = `ask`
	ValueConfirm   = `confirm`
	ValueOff       = `off`
	ValueNo        = `no`
	ValueNone      = `none`
	ValueYes       = `yes`
)

// Valid values for key AddressFamily.
const (
	ValueAny   = `any`
	ValueInet  = `inet`
	ValueInet6 = `inet6`
)

// List of default key value.
const (
	DefConnectionAttempts = `1`
	DefPort               = `22`
	DefXAuthLocation      = `/usr/X11R6/bin/xauth`
)

// defaultCASignatureAlgorithms return list of default signature algorithms
// that client supported.
func defaultCASignatureAlgorithms() []string {
	return []string{
		ssh.KeyAlgoECDSA256,
		ssh.KeyAlgoECDSA384,
		ssh.KeyAlgoECDSA521,
		ssh.KeyAlgoED25519,
		ssh.KeyAlgoRSA,
	}
}

// defaultIdentityFile return list of default IdentityFile.
func defaultIdentityFile() []string {
	return []string{
		`~/.ssh/id_dsa`,
		`~/.ssh/id_ecdsa`,
		`~/.ssh/id_ed25519`,
		`~/.ssh/id_rsa`,
	}
}

// defaultUserKnownHostsFile return list of default KnownHostsFile.
func defaultUserKnownHostsFile() []string {
	return []string{
		`~/.ssh/known_hosts`,
		`~/.ssh/known_hosts2`,
	}
}

// Section is the type that represent SSH client Host and Match section in
// configuration.
type Section struct {
	// Field store the unpacked key and value of Section.
	// For section key that is not expecting string value, one can use
	// FieldBool or FieldInt64.
	Field map[string]string

	// env contains the key and value from SetEnv field.
	env map[string]string

	// name contains the raw value after Host or Match.
	name string

	// WorkingDir contains the directory where the SSH client started.
	// This value is required when client want to copy file from/to
	// remote.
	// This field is optional, default to current working directory from
	// os.Getwd() or user's home directory.
	WorkingDir string

	// User's home directory.
	homeDir string

	certificateFile []string
	IdentityFile    []string
	knownHostsFile  []string
	sendEnv         []string

	// Patterns for Host section.
	patterns []*pattern

	// Criteria for Match section.
	criteria []*matchCriteria

	// If true indicated that this is Match section.
	useCriteria bool
}

// NewSection create new Host or Match with default values.
func NewSection(name string) *Section {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		name = `*`
	}

	return &Section{
		Field: map[string]string{
			KeyChallengeResponseAuthentication: ValueYes,
			KeyCheckHostIP:                     ValueYes,
			KeyConnectionAttempts:              DefConnectionAttempts,
			KeyPort:                            DefPort,
			KeyXAuthLocation:                   DefXAuthLocation,
		},
		env:  map[string]string{},
		name: name,
	}
}

func newSectionHost(rawPattern string) (host *Section) {
	patterns := strings.Fields(rawPattern)

	host = NewSection(rawPattern)
	host.patterns = make([]*pattern, 0, len(patterns))

	for _, pattern := range patterns {
		pat := newPattern(pattern)
		host.patterns = append(host.patterns, pat)
	}
	return host
}

// CASignatureAlgorithms return list of signature algorithms set from
// KeyCASignatureAlgorithms.
// If not set it will return the default CA signature algorithms.
func (section *Section) CASignatureAlgorithms() []string {
	var value = section.Field[KeyCASignatureAlgorithms]
	if len(value) == 0 {
		return defaultCASignatureAlgorithms()
	}
	return strings.Split(value, `,`)
}

// CanonicalDomains return list CanonicalDomains set in Section.
func (section *Section) CanonicalDomains() []string {
	var value = section.Field[KeyCanonicalDomains]
	if len(value) == 0 {
		return nil
	}
	return strings.Fields(value)
}

// CanonicalizePermittedCNames return the permitted CNAMEs set in Section,
// from KeyCanonicalizePermittedCNames.
func (section *Section) CanonicalizePermittedCNames() (pcnames *PermittedCNAMEs, err error) {
	var value = section.Field[KeyCanonicalizePermittedCNames]
	if len(value) == 0 {
		return nil, nil
	}
	pcnames, err = parseCanonicalizePermittedCNames(value)
	if err != nil {
		return nil, err
	}
	return pcnames, nil
}

// CertificateFile return list of certificate file, if its set in Host or
// Match configuration.
func (section *Section) CertificateFile() []string {
	return section.certificateFile
}

// Environments return system and/or custom environment that will be passed
// to remote machine.
// The key and value is derived from "SendEnv" and "SetEnv".
func (section *Section) Environments(sysEnv map[string]string) (env map[string]string) {
	var (
		key         string
		val         string
		sendPattern string
		ok          bool
	)

	env = make(map[string]string, len(section.sendEnv)+len(section.env))

	for key, val = range sysEnv {
		for _, sendPattern = range section.sendEnv {
			ok, _ = filepath.Match(sendPattern, key)
			if ok {
				env[key] = val
			}
		}
	}
	for key, val = range section.env {
		env[key] = val
	}
	return env
}

// FieldBool get the Field value as boolean.
// It will return false if key is not exist or value is invalid.
func (section *Section) FieldBool(key string) (vbool bool) {
	var vstr = section.Field[key]
	if len(vstr) == 0 {
		return false
	}
	vbool, _ = parseBool(key, vstr)
	return vbool
}

// FieldInt64 get the Field value as int64.
// If the value is unparseable as int64 it will return 0.
func (section *Section) FieldInt64(key string) (val int64) {
	var vstr = section.Field[key]
	if len(vstr) == 0 {
		return 0
	}
	val, _ = strconv.ParseInt(vstr, 10, 64)
	return val
}

// Hostname return the hostname of this section.
func (section *Section) Hostname() string {
	return section.Field[KeyHostname]
}

// IdentityAgent get the identity agent either from section config variable
// "IdentityAgent" or from environment variable SSH_AUTH_SOCK.
//
// There are four possible value: SSH_AUTH_SOCK, <$STRING>, <PATH>, or
// "none".
// If SSH_AUTH_SOCK, the socket path is read from the environment variable
// SSH_AUTH_SOCK.
// If value start with "$", then the socket path is set based on value of
// that environment variable.
// Other string beside "none" will be considered as path to socket.
//
// It will return empty string if IdentityAgent set to "none" or
// SSH_AUTH_SOCK is empty.
func (section *Section) IdentityAgent() string {
	var value = section.Field[KeyIdentityAgent]
	if value == `none` {
		return ``
	}
	if len(value) == 0 || value == envSSHAuthSock {
		return os.Getenv(envSSHAuthSock)
	}
	if value[0] == '$' {
		// Read the socket from environment variable defined by
		// value.
		return os.Getenv(value[1:])
	}
	// IdentityAgent set to path to socket.
	return value
}

// Port return the remote machine port of this section.
func (section *Section) Port() string {
	return section.Field[KeyPort]
}

// Signers convert the IdentityFile to ssh.Signer for authentication using
// PublicKey.
//
// This method will ask for passphrase from terminal, if one of IdentityFile
// is protected.
// Unless the value of IdentityFile changes, this method should be called
// only once, otherwise it will ask passphrase on every call.
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

		signers = append(signers, signer)
	}
	return signers, nil
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

	if len(section.IdentityFile) == 0 {
		section.IdentityFile = defaultIdentityFile()
	}

	for x, identFile := range section.IdentityFile {
		if identFile[0] == '~' {
			section.IdentityFile[x] = strings.Replace(identFile, "~", section.homeDir, 1)
		}
	}
}

// mergeField merge the Field from other Section.
func (section *Section) mergeField(other *Section) {
	var (
		key   string
		value string
	)
	for key, value = range other.Field {
		// The key and value in other should be valid, so no need to
		// check for error.
		_ = section.Set(key, value)
	}
}

// Set the section field by raw key and value.
func (section *Section) Set(key, value string) (err error) {
	switch key {
	case KeyAddKeysToAgent:
		err = validateAddKeysToAgent(value)
	case KeyAddressFamily:
		err = validateAddressFamily(value)
	case KeyBatchMode:
		_, err = parseBool(key, value)
	case KeyBindAddress:
	case KeyBindInterface:
	case KeyCanonicalDomains:
	case KeyCanonicalizeFallbackLocal:
		_, err = parseBool(key, value)
	case KeyCanonicalizeHostname:
		err = validateCanonicalizeHostname(value)
	case KeyCanonicalizeMaxDots:
		_, err = strconv.Atoi(value)

	case KeyCanonicalizePermittedCNames:
		_, err = parseCanonicalizePermittedCNames(value)

	case KeyCASignatureAlgorithms:
		value = strings.ToLower(value)

	case KeyCertificateFile:
		section.certificateFile = append(section.certificateFile, value)

	case KeyChallengeResponseAuthentication:
		_, err = parseBool(key, value)

	case KeyCheckHostIP:
		_, err = parseBool(key, value)

	case KeyClearAllForwardings:
		_, err = parseBool(key, value)
	case KeyCompression:
		_, err = parseBool(key, value)
	case KeyConnectionAttempts:
		_, err = strconv.Atoi(value)
	case KeyConnectTimeout:
		_, err = strconv.Atoi(value)

	case KeyIdentityAgent:

	case KeyIdentityFile:
		section.IdentityFile = append(section.IdentityFile, value)

	case KeyHostname:
		value = strings.ToLower(value)
	case KeyPort:
		_, err = strconv.Atoi(value)

	case KeySendEnv:
		section.sendEnv = append(section.sendEnv, value)
	case KeySetEnv:
		section.setEnv(value)
	case KeyUser:
		// User name is case sensitive.
	case KeyUserKnownHostsFile:
		section.setUserKnownHostsFile(value)

	case KeyVisualHostKey:
		_, err = parseBool(key, value)
	case KeyXAuthLocation:
	default:
		// Store the unknown key into Field.
	}
	if err != nil {
		return err
	}
	section.Field[key] = value
	return nil
}

// User return the user value of this section.
func (section *Section) User() string {
	return section.Field[KeyUser]
}

// UserKnownHostsFile return list of user known_hosts file set in this
// Section.
func (section *Section) UserKnownHostsFile() []string {
	if len(section.knownHostsFile) == 0 {
		return defaultUserKnownHostsFile()
	}
	return section.knownHostsFile
}

// MarshalText encode the Section back to ssh_config format.
// The key is indented by two spaces.
func (section *Section) MarshalText() (text []byte, err error) {
	var buf bytes.Buffer

	if section.useCriteria {
		buf.WriteString(`Match`)

		var criteria *matchCriteria
		for _, criteria = range section.criteria {
			buf.WriteByte(' ')
			criteria.WriteTo(&buf)
		}
	} else {
		buf.WriteString(`Host`)

		if len(section.patterns) == 0 {
			buf.WriteByte(' ')
			buf.WriteString(section.name)
		} else {
			var pat *pattern
			for _, pat = range section.patterns {
				buf.WriteByte(' ')
				pat.WriteTo(&buf)
			}
		}
	}
	buf.WriteByte('\n')

	var (
		listKey = make([]string, 0, len(section.Field))
		key     string
		val     string
	)
	for key = range section.Field {
		listKey = append(listKey, key)
	}
	sort.Strings(listKey)

	for _, key = range listKey {
		if key == KeyIdentityFile {
			for _, val = range section.IdentityFile {
				buf.WriteString(`  `)
				buf.WriteString(key)
				buf.WriteByte(' ')
				buf.WriteString(section.pathUnfold(val))
				buf.WriteByte('\n')
			}
			continue
		}

		buf.WriteString(`  `)
		buf.WriteString(key)
		buf.WriteByte(' ')
		buf.WriteString(section.Field[key])
		buf.WriteByte('\n')
	}

	return buf.Bytes(), nil
}

// WriteTo marshal the Section into text and write it to w.
func (section *Section) WriteTo(w io.Writer) (n int64, err error) {
	var text []byte
	text, _ = section.MarshalText()

	var c int
	c, err = w.Write(text)
	return int64(c), err
}

// pathUnfold replace the home directory prefix with '~'.
func (section *Section) pathUnfold(in string) (out string) {
	if !strings.HasPrefix(in, section.homeDir) {
		return in
	}
	out = `~` + in[len(section.homeDir):]
	return out
}

// setEnv set the Environments with key and value of format "KEY=VALUE".
func (section *Section) setEnv(env string) {
	kv := strings.SplitN(env, "=", 2)
	if len(kv) == 2 {
		section.env[kv[0]] = kv[1]
	}
}

func (section *Section) setUserKnownHostsFile(val string) {
	var list = strings.Fields(val)
	if len(list) > 0 {
		section.knownHostsFile = append(section.knownHostsFile, list...)
	}
}

func validateAddKeysToAgent(val string) (err error) {
	switch val {
	case ValueAlways, ValueAsk, ValueConfirm, ValueNo, ValueYes:
	default:
		return fmt.Errorf(`%s: invalid value %q`, KeyAddKeysToAgent, val)
	}
	return nil
}

func validateAddressFamily(val string) (err error) {
	switch val {
	case ValueAny, ValueInet, ValueInet6:
	default:
		return fmt.Errorf(`%s: invalid value %q`, KeyAddressFamily, val)
	}
	return nil
}

func validateCanonicalizeHostname(val string) (err error) {
	switch val {
	case ValueNo, ValueAlways, ValueYes:
	default:
		return fmt.Errorf(`%s: invalid value %q`, KeyCanonicalizeHostname, val)
	}
	return nil
}

func parseCanonicalizePermittedCNames(val string) (pcnames *PermittedCNAMEs, err error) {
	sourceTarget := strings.Split(val, ":")
	if len(sourceTarget) != 2 {
		return nil, fmt.Errorf(`%s: invalid rule`, KeyCanonicalizePermittedCNames)
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

	pcnames = &PermittedCNAMEs{
		sources: sources,
		targets: targets,
	}
	return pcnames, nil
}
