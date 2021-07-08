// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//
// Package sftp implement SSH File Transfer Protocol v3 as defined in
// draft-ietf-secsh-filexfer-02.txt.
//
// The sftp package extend the golang.org/x/crypto/ssh package by
// implementing "sftp" subsystem using the ssh.Client connection.
//
package sftp

import "fmt"

const (
	subsystemNameSftp        = "sftp"
	defFxpVersion     uint32 = 3
	maxPacketRead     uint32 = 32000
)

// List of values for FXP_OPEN pflags.
const (
	ssh_FXF_READ   uint32 = 0x00000001
	ssh_FXF_WRITE  uint32 = 0x00000002
	ssh_FXF_APPEND uint32 = 0x00000004
	ssh_FXF_CREAT  uint32 = 0x00000008
	ssh_FXF_TRUNC  uint32 = 0x00000010
	ssh_FXF_EXCL   uint32 = 0x00000020
)

// List of values for FXP_STATUS code.
const (
	ssh_FX_OK uint32 = iota
	ssh_FX_EOF
	ssh_FX_NO_SUCH_FILE
	ssh_FX_PERMISSION_DENIED
	ssh_FX_FAILURE
	ssh_FX_BAD_MESSAGE
	ssh_FX_NO_CONNECTION
	ssh_FX_CONNECTION_LOST
	ssh_FX_OP_UNSUPPORTED
)

func ErrUnexpectedResponse(exp, got byte) error {
	return fmt.Errorf("sftp: expecting packet type %d, got %d", exp, got)
}
