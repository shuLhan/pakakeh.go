// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sshconfig

// PermittedCNAMEs contains list of canonical names (CNAME) for source and
// target.
type PermittedCNAMEs struct {
	sources []*pattern
	targets []*pattern
}
