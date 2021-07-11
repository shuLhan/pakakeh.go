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
// For information, even if scp working normally on server, this package
// functionalities will not working if the server disable or does not support
// the "sftp" subsystem.
// For reference, on openssh, the following configuration
//
//	Subsystem sftp /usr/lib/sftp-server
//
// should be un-commented on /etc/ssh/sshd_config if its exist.
//
package sftp

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
)

const (
	subsystemNameSftp        = "sftp"
	defFxpVersion     uint32 = 3
	maxPacketRead     uint32 = 32000
)

// List of valid values for OpenFile flag.
const (
	// OpenFlagRead open remote file for read only.
	OpenFlagRead uint32 = 0x00000001

	// OpenFlagWrite open remote file for write only.
	OpenFlagWrite uint32 = 0x00000002

	// OpenFlagAppend any write to remote file handle opened with this
	// file will be appended at the end of the file.
	OpenFlagAppend uint32 = 0x00000004

	// OpenFlagCreate create new file on the server if it does not exist.
	OpenFlagCreate uint32 = 0x00000008

	// OpenFlagTruncate truncated the remote file.
	// OpenFlagCreate MUST also be specified.
	OpenFlagTruncate uint32 = 0x00000010

	// OpenFlagExcl passing this flag along OpenFlagCreate will fail the
	// remote file already exists.
	// OpenFlagCreate MUST also be specified.
	OpenFlagExcl uint32 = 0x00000020
)

// List of values for FXP_STATUS code.
const (
	statusCodeOK uint32 = iota
	statusCodeEOF
	statusCodeNoSuchFile
	statusCodePermissionDenied
	statusCodeFailure
	statusCodeBadMessage
	statusCodeNoConnection
	statusCodeConnectionLost
	statusCodeOpUnsupported
)

//
// Some response status code from server is mapped to existing errors on
// standard packages,
//
//	SSH_FX_EOF               (1) = io.EOF
//	SSH_FX_NO_SUCH_FILE      (2) = fs.ErrNotExist
//	SSH_FX_PERMISSION_DENIED (3) = fs.ErrPermission
//
// Other errors is defined below,
//
var (
	// ErrFailure or SSH_FX_FAILURE(4) is a generic catch-all error
	// message; it should be returned if an error occurs for which there
	// is no more specific error code defined.
	ErrFailure = errors.New("sftp: failure")

	// ErrBadMessage or SSH_FX_BAD_MESSAGE(5) may be returned if a badly
	// formatted packet or protocol incompatibility is detected.
	ErrBadMessage = errors.New("sftp: bad message")

	// ErrNoConnection or SSH_FX_NO_CONNECTION(6) indicates that the
	// client has no connection to the server.
	// This error returned by client not server.
	ErrNoConnection = errors.New("sftp: no connection")

	// ErrConnectionLost or SSH_FX_CONNECTION_LOST(7) indicated that the
	// connection to the server has been lost.
	// This error returned by client not server.
	ErrConnectionLost = errors.New("sftp: connection lost")

	// ErrOpUnsupported or SSH_FX_OP_UNSUPPORTED(8) indicates that an
	// attempt was made to perform an operation which is not supported for
	// the server or the server does not implement an operation.
	ErrOpUnsupported = errors.New("sftp: operation unsupported")

	// ErrSubsystem indicates that the server does not support or enable
	// the Subsystem for sftp.
	// For reference, on openssh, the following configuration
	//
	//	Subsystem sftp /usr/lib/sftp-server
	//
	// maybe commented on /etc/ssh/sshd_config.
	//
	ErrSubsystem = errors.New("sftp: unsupported subsystem")

	// ErrVersion indicates that this client does not support the version
	// on server.
	ErrVersion = errors.New("sftp: unsupported version")
)

func errBadMessage(msg string) error {
	return fmt.Errorf("%w: %s", ErrBadMessage, msg)
}

func errFailure(msg string) error {
	return fmt.Errorf("%w: %s", ErrFailure, msg)
}

func errUnexpectedResponse(exp, got byte) error {
	return fmt.Errorf("sftp: expecting packet type %d, got %d", exp, got)
}

func errVersion(version uint32) error {
	return fmt.Errorf("%w: %d", ErrVersion, version)
}

func handleStatusCode(code uint32, message string) error {
	switch code {
	case statusCodeEOF:
		return io.EOF
	case statusCodeNoSuchFile:
		return fs.ErrNotExist
	case statusCodePermissionDenied:
		return fs.ErrPermission
	case statusCodeFailure:
		return errFailure(message)
	case statusCodeBadMessage:
		return errBadMessage(message)
	case statusCodeNoConnection:
		return ErrNoConnection
	case statusCodeConnectionLost:
		return ErrConnectionLost
	case statusCodeOpUnsupported:
		return ErrOpUnsupported
	}
	return fmt.Errorf("sftp: %d %s", code, message)
}
