// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>
// Copyright 2015-2018, Shulhan <ms@kilabit.info>. All rights reserved.

package dsv

// Config for working with DSV configuration.
type Config struct {
	// ConfigPath path to configuration file.
	ConfigPath string
}

// GetConfigPath return the base path of configuration file.
func (cfg *Config) GetConfigPath() string {
	return cfg.ConfigPath
}

// SetConfigPath for reading input and writing rejected file.
func (cfg *Config) SetConfigPath(dir string) {
	cfg.ConfigPath = dir
}
