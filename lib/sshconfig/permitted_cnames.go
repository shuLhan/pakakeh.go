// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2020 Shulhan <ms@kilabit.info>

package sshconfig

// PermittedCNAMEs contains list of canonical names (CNAME) for source and
// target.
type PermittedCNAMEs struct {
	sources []*pattern
	targets []*pattern
}
