// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sftp

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client for SFTP.
type Client struct {
	sess *ssh.Session

	exts    extensions
	pipeIn  io.WriteCloser
	pipeOut io.Reader
	pipeErr io.Reader

	// The requestId is unique number that will be incremented by client,
	// to prevent the same ID generated on concurrent operations.
	requestId uint32

	version uint32

	mtxId sync.Mutex
}

// NewClient create and initialize new client for SSH file transfer protocol.
//
// On failure, it will return ErrSubsystem if the server does not support
// "sftp" subsystem, ErrVersion if the client does not support the
// server version, and other errors.
func NewClient(sshc *ssh.Client) (cl *Client, err error) {
	logp := "New"
	cl = &Client{}

	cl.sess, err = sshc.NewSession()
	if err != nil {
		return nil, fmt.Errorf("%s: NewSession: %w", logp, err)
	}

	cl.pipeIn, err = cl.sess.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("%s: StdinPipe: %w", logp, err)
	}
	cl.pipeOut, err = cl.sess.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("%s: StdoutPipe: %w", logp, err)
	}
	cl.pipeErr, err = cl.sess.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("%s: StderrPipe: %w", logp, err)
	}

	err = cl.sess.RequestSubsystem(subsystemNameSftp)
	if err != nil {
		if err.Error() == "ssh: subsystem request failed" {
			return nil, ErrSubsystem
		}
		return nil, fmt.Errorf("%s: RequestSubsystem: %w", logp, err)
	}

	cl.requestId = uint32(time.Now().Unix())

	err = cl.init()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	return cl, nil
}

// Close the client sftp session and release all resources.
func (cl *Client) Close() (err error) {
	err = cl.sess.Close()

	cl.requestId = 0
	cl.pipeErr = nil
	cl.pipeOut = nil
	cl.pipeIn = nil

	return fmt.Errorf(`Close: %w`, err)
}

// CloseFile close the remote file handle.
func (cl *Client) CloseFile(fh *FileHandle) (err error) {
	if fh == nil {
		return nil
	}
	var (
		logp    = "Close"
		req     = cl.generatePacket()
		payload = req.fxpClose(fh)
	)

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}

	return nil
}

// Create creates or truncates the named file.
// If the remote file does not exist, it will be created.
// If the remote file already exists, it will be truncated.
// On success, it will return the remote FileHandle ready for write only.
func (cl *Client) Create(remoteFile string, fa *FileAttrs) (*FileHandle, error) {
	pflags := OpenFlagWrite | OpenFlagCreate | OpenFlagTruncate
	return cl.OpenFile(remoteFile, pflags, fa)
}

// Fsetstat set the file attributes based on the opened remote file handle.
func (cl *Client) Fsetstat(fh *FileHandle, fa *FileAttrs) (err error) {
	var (
		logp = "Fsetstat"
		req  = cl.generatePacket()
	)
	if fh == nil || fa == nil {
		return nil
	}

	payload := req.fxpFsetstat(fh, fa)

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}
	return nil
}

// Fstat get the file attributes based on the opened remote file handle.
func (cl *Client) Fstat(fh *FileHandle) (fa *FileAttrs, err error) {
	var (
		logp    = "Fstat"
		req     = cl.generatePacket()
		payload = req.fxpFstat(fh)
	)
	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpAttrs {
		return nil, errUnexpectedResponse(packetKindFxpAttrs, res.kind)
	}
	fa = res.fa
	fa.name = fh.remotePath
	res.fa = nil
	return fa, nil
}

// Get copy remote file to local.
// The local file will be created if its not exist; otherwise it will
// truncated.
func (cl *Client) Get(remoteFile, localFile string) (err error) {
	var (
		logp   = "Get"
		offset uint64
	)

	fin, err := cl.Open(remoteFile)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	fout, err := os.Create(localFile)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	var data []byte
	for {
		data, err = cl.Read(fin, offset)
		if len(data) > 0 {
			_, err = fout.Write(data)
			if err != nil {
				break
			}
		}
		if err != nil {
			break
		}
		offset += uint64(len(data))
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = fout.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = cl.CloseFile(fin)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return err
}

// Lstat get the file attributes based on the remote file path.
// Unlike Stat(), the Lstat method does not follow symbolic links.
func (cl *Client) Lstat(remoteFile string) (fa *FileAttrs, err error) {
	var (
		logp    = "Lstat"
		req     = cl.generatePacket()
		payload = req.fxpLstat(remoteFile)
	)
	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpAttrs {
		return nil, errUnexpectedResponse(packetKindFxpAttrs, res.kind)
	}
	fa = res.fa
	fa.name = remoteFile
	res.fa = nil
	return fa, nil
}

// Mkdir create new directory on the server.
func (cl *Client) Mkdir(path string, fa *FileAttrs) (err error) {
	var (
		logp = "Mkdir"
		req  = cl.generatePacket()
	)
	if fa == nil {
		fa = newFileAttrs()
	}
	payload := req.fxpMkdir(path, fa)
	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}
	return nil
}

// Open the remote file for read only.
func (cl *Client) Open(remoteFile string) (fh *FileHandle, err error) {
	return cl.OpenFile(remoteFile, OpenFlagRead, nil)
}

// OpenFile open remote file with custom open flag (OpenFlagRead,
// OpenFlagWrite, and so on) and with specific file attributes.
func (cl *Client) OpenFile(remoteFile string, flags uint32, fa *FileAttrs) (fh *FileHandle, err error) {
	var (
		logp = "open"
		req  = cl.generatePacket()
	)
	if fa == nil {
		fa = newFileAttrs()
	}

	payload := req.fxpOpen(remoteFile, flags, fa)

	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpHandle {
		return nil, errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	fh = res.fh
	fh.remotePath = remoteFile
	res.fh = nil
	return fh, nil
}

// Opendir open the directory on the server.
func (cl *Client) Opendir(path string) (fh *FileHandle, err error) {
	var (
		logp = "Opendir"
		req  = cl.generatePacket()
	)
	payload := req.fxpOpendir(path)
	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpHandle {
		return nil, errUnexpectedResponse(packetKindFxpHandle, res.kind)
	}
	fh = res.fh
	fh.remotePath = path
	res.fh = nil
	return fh, nil
}

// Put local file to remote file.
func (cl *Client) Put(localFile, remoteFile string) (err error) {
	logp := "Put"

	fin, err := os.Open(localFile)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	finfo, err := fin.Stat()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	fa := NewFileAttrs(finfo)

	fout, err := cl.Create(remoteFile, fa)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	var (
		offset uint64
		data   = make([]byte, 32000)
		n      int
	)

	n, err = fin.Read(data)
	for n > 0 {
		if err != nil {
			break
		}
		err = cl.Write(fout, offset, data[:n])
		if err != nil {
			break
		}
		if n < len(data) {
			break
		}

		offset += uint64(n)
		n, err = fin.Read(data)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = cl.CloseFile(fout)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	return nil
}

// Read the remote file using handle on specific offset.
// On end-of-file it will return empty data with io.EOF.
func (cl *Client) Read(fh *FileHandle, offset uint64) (data []byte, err error) {
	var (
		logp    = "Read"
		req     = cl.generatePacket()
		payload = req.fxpRead(fh, offset, maxPacketRead)
	)

	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpData {
		return nil, errUnexpectedResponse(packetKindFxpData, res.kind)
	}
	data = res.data
	res.data = nil
	return data, nil
}

// Readdir list files and/or directories inside the handle.
func (cl *Client) Readdir(fh *FileHandle) (nodes []fs.DirEntry, err error) {
	var (
		logp    = "Readdir"
		req     = cl.generatePacket()
		payload = req.fxpReaddir(fh)
	)

	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpName {
		return nil, errUnexpectedResponse(packetKindFxpName, res.kind)
	}
	for _, node := range res.nodes {
		nodes = append(nodes, node)
	}
	res.nodes = nil
	return nodes, nil
}

// Readlink read the target of a symbolic link.
func (cl *Client) Readlink(linkPath string) (node fs.DirEntry, err error) {
	var (
		logp    = "Readlink"
		req     = cl.generatePacket()
		payload = req.fxpReadlink(linkPath)
	)

	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpName {
		return nil, errUnexpectedResponse(packetKindFxpName, res.kind)
	}
	node = res.nodes[0]
	res.nodes = nil
	return node, nil
}

// Realpath canonicalize any given path name to an absolute path.
// This is useful for converting path names containing ".." components or
// relative pathnames without a leading slash into absolute paths.
func (cl *Client) Realpath(path string) (node fs.DirEntry, err error) {
	var (
		logp    = "Realpath"
		req     = cl.generatePacket()
		payload = req.fxpRealpath(path)
	)

	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpName {
		return nil, errUnexpectedResponse(packetKindFxpName, res.kind)
	}
	node = res.nodes[0]
	res.nodes = nil
	return node, nil
}

// Remove the remote file.
func (cl *Client) Remove(remoteFile string) (err error) {
	var (
		logp    = "Remove"
		req     = cl.generatePacket()
		payload = req.fxpRemove(remoteFile)
	)

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}

	return nil
}

// Rename the file, or move the file, from old path to new path.
func (cl *Client) Rename(oldPath, newPath string) (err error) {
	var (
		logp    = "Rename"
		req     = cl.generatePacket()
		payload = req.fxpRename(oldPath, newPath)
	)

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}

	return nil
}

// Rmdir remove the directory on the server.
func (cl *Client) Rmdir(path string) (err error) {
	var (
		logp    = "Rmdir"
		req     = cl.generatePacket()
		payload = req.fxpRmdir(path)
	)

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}

	return nil
}

// Setstat change the file attributes on remote file or directory.
// These request can be used for operations such as changing the ownership,
// permissions or access times, as well as for truncating a file.
func (cl *Client) Setstat(remoteFile string, fa *FileAttrs) (err error) {
	var (
		logp = "Setstat"
		req  = cl.generatePacket()
	)
	if len(remoteFile) == 0 || fa == nil {
		return nil
	}

	payload := req.fxpSetstat(remoteFile, fa)

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}

	return nil
}

// Stat get the file attributes based on the remote file path.
// This method follow symbolic links.
func (cl *Client) Stat(remoteFile string) (fa *FileAttrs, err error) {
	var (
		logp    = "Stat"
		req     = cl.generatePacket()
		payload = req.fxpStat(remoteFile)
	)
	res, err := cl.send(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind == packetKindFxpStatus {
		return nil, handleStatusCode(res.code, res.message)
	}
	if res.kind != packetKindFxpAttrs {
		return nil, errUnexpectedResponse(packetKindFxpAttrs, res.kind)
	}
	fa = res.fa
	fa.name = remoteFile
	res.fa = nil
	return fa, nil
}

// Symlink create a symbolic link on the server.
// The `linkpath' specifies the path name of the symlink to be created and
// `targetpath' specifies the target of the symlink.
func (cl *Client) Symlink(targetPath, linkPath string) (err error) {
	var (
		logp    = "Symlink"
		req     = cl.generatePacket()
		payload = req.fxpSymlink(targetPath, linkPath)
	)
	if len(targetPath) == 0 || len(linkPath) == 0 {
		return nil
	}

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}
	return nil
}

// Write write the data into remote file at specific offset.
func (cl *Client) Write(fh *FileHandle, offset uint64, data []byte) (err error) {
	var (
		logp    = "Write"
		req     = cl.generatePacket()
		payload = req.fxpWrite(fh, offset, data)
	)

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	if res.kind != packetKindFxpStatus {
		return errUnexpectedResponse(packetKindFxpStatus, res.kind)
	}
	if res.code != statusCodeOK {
		return handleStatusCode(res.code, res.message)
	}
	return nil
}

func (cl *Client) generatePacket() (pac *packet) {
	cl.mtxId.Lock()
	cl.requestId++
	pac = &packet{
		requestId: cl.requestId,
	}
	cl.mtxId.Unlock()
	return pac
}

func (cl *Client) init() (err error) {
	logp := "init"

	req := cl.generatePacket()
	payload := req.fxpInit(defFxpVersion)

	res, err := cl.send(payload)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	cl.version = res.version
	cl.exts = res.exts
	res.exts = nil

	if cl.version != defFxpVersion {
		return errVersion(cl.version)
	}

	return nil
}

func (cl *Client) read() (res []byte, err error) {
	var (
		logp  = "read"
		block = make([]byte, 1024)
	)

	n, err := cl.pipeOut.Read(block)
	for n > 0 {
		res = append(res, block[:n]...)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("%s: %w", logp, err)
		}
		if n < len(block) {
			break
		}
		n, err = cl.pipeOut.Read(block)
	}
	return res, nil
}

func (cl *Client) send(payload []byte) (res *packet, err error) {
	var (
		logp = "send"
	)

	_, err = cl.pipeIn.Write(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: Write: %w", logp, err)
	}

	payload, err = cl.read()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	// TODO: check the response ID.
	res, err = unpackPacket(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	return res, nil
}
