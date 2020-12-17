// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sql

//
// ClientOptions contains options to connect to database server, including the
// migration directory.
//
type ClientOptions struct {
	DriverName   string
	DSN          string
	MigrationDir string
}
