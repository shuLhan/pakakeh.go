// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package sql

// ClientOptions contains options to connect to database server, including the
// migration directory.
type ClientOptions struct {
	DriverName   string
	DSN          string
	MigrationDir string
}
