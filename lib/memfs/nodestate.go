// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

//
// NodeState contains the information about the file and its state.
//
type NodeState struct {
	// Node represent the file information.
	Node *Node
	// State of file, its either created, modified, or deleted.
	State FileState
}
