// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package config provide the ssh_config(5) parser and getter.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	envSSHAuthSock = `SSH_AUTH_SOCK`
)

const (
	keyHost  = "host"
	keyMatch = "match"
)

var (
	errMultipleEqual = errors.New("multiple '=' character")
)

// Config contains mapping of host's patterns and its options from SSH
// configuration file.
type Config struct {
	envs map[string]string

	// dir store the path to the "config" directory.
	dir string

	// workDir store the current working directory.
	workDir string

	homeDir string

	sections []*Section
}

// newConfig create new SSH Config instance from file.
func newConfig(file string) (cfg *Config, err error) {
	cfg = &Config{}

	// If file is empty, the dir is set to ".".
	cfg.dir, err = filepath.Abs(filepath.Dir(file))
	if err != nil {
		return nil, err
	}

	cfg.workDir, err = os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg.homeDir, err = os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// Load SSH configuration from file.
func Load(file string) (cfg *Config, err error) {
	if len(file) == 0 {
		return nil, nil
	}

	var (
		logp    = "Load"
		section *Section
	)

	cfg, err = newConfig(file)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", logp, file, err)
	}

	cfg.loadEnvironments()

	var p *parser

	p, err = newParser(cfg)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", logp, file, err)
	}

	var lines []string

	lines, err = p.load("", file)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", logp, file, err)
	}

	var (
		line string
		x    int
	)
	for x, line = range lines {
		if line[0] == '#' {
			continue
		}
		key, value, err := parseKeyValue(line)
		if err != nil {
			return nil, fmt.Errorf("%s %s line %d: %w", logp, file, x, err)
		}

		switch key {
		case keyHost:
			if section != nil {
				cfg.sections = append(cfg.sections, section)
				section = nil
			}
			section = newSectionHost(cfg, value)
		case keyMatch:
			if section != nil {
				cfg.sections = append(cfg.sections, section)
				section = nil
			}
			section, err = newSectionMatch(cfg, value)
			if err != nil {
				return nil, fmt.Errorf("%s %s line %d: %w", logp, file, x+1, err)
			}
		default:
			if section == nil {
				// No "Host" or "Match" define yet.
				continue
			}
			err = section.Set(key, value)
			if err != nil {
				return nil, fmt.Errorf("%s %s line %d: %w", logp, file, x+1, err)
			}
		}
	}
	if section != nil {
		cfg.sections = append(cfg.sections, section)
		section = nil
	}

	return cfg, nil
}

// Get the Host or Match configuration that match with the host name "s".
// If no Host or Match found, it return non-nil Section with default values.
func (cfg *Config) Get(host string) (section *Section) {
	section = NewSection(cfg, host)

	var hostMatch *Section
	for _, hostMatch = range cfg.sections {
		if hostMatch.isMatch(host) {
			section.merge(hostMatch)
		}
	}
	section.setDefaults()

	if host != `` && section.Field[KeyHostname] == `` {
		section.Set(KeyHostname, host)
	}

	return section
}

// Prepend other Config's sections to this Config.
// The other's sections will be at the top of the list.
//
// This function can be useful if we want to load another SSH config file
// without using Include directive.
func (cfg *Config) Prepend(other *Config) {
	newSections := make([]*Section, 0, len(cfg.sections)+len(other.sections))
	newSections = append(newSections, other.sections...)
	newSections = append(newSections, cfg.sections...)
	cfg.sections = newSections
}

// loadEnvironments get all environments variables and store it in the map for
// future use by SendEnv.
func (cfg *Config) loadEnvironments() {
	envs := os.Environ()
	for _, env := range envs {
		kv := strings.SplitN(env, "=", 2)
		if len(kv) == 0 {
			cfg.envs[kv[0]] = kv[1]
		}
	}
}

func parseBool(key, val string) (out bool, err error) {
	switch val {
	case ValueNo:
		return false, nil
	case ValueYes:
		return true, nil
	}
	return false, fmt.Errorf("%s: invalid value %q", key, val)
}

// parseKeyValue from single line.
//
// ssh_config(5):
//
//	Configuration options may be separated by whitespace or optional
//	whitespace and exactly one `='; the latter format is useful to avoid
//	the need to quote whitespace ...
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

// patternToRegex convert the Host and Match pattern string into regex.
func patternToRegex(in string) (out string) {
	sr := make([]rune, 0, len(in))
	for _, r := range in {
		switch r {
		case '*', '?':
			sr = append(sr, '.')
		case '.':
			sr = append(sr, '\\')
		}
		sr = append(sr, r)
	}
	return string(sr)
}
