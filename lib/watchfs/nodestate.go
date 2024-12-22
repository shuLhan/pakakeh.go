// SPDX-FileCopyrightText: 2019 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

// NodeState contains the information about the file and its state.
type NodeState struct {
	// Node represent the file information.
	Node Node
	// State of file, its either created, modified, or deleted.
	State FileState
}
