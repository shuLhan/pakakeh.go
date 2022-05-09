// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sftp

import (
	"encoding/binary"
	"io"
	"io/fs"
	"time"
)

const (
	fileModeSticky      uint32 = 0001000
	fileModeSetgid      uint32 = 0002000
	fileModeSetuid      uint32 = 0004000
	fileTypeFifo        uint32 = 0010000
	fileTypeCharDevice  uint32 = 0020000
	fileTypeDirectory   uint32 = 0040000
	fileTypeBlockDevice uint32 = 0060000
	fileTypeRegular     uint32 = 0100000
	fileTypeSymlink     uint32 = 0120000
	fileTypeSocket      uint32 = 0140000
	fileTypeMask        uint32 = 0170000
)

// List of valid values for FileAttrs.flags.
const (
	attr_SIZE        uint32 = 0x00000001
	attr_UIDGID      uint32 = 0x00000002
	attr_PERMISSIONS uint32 = 0x00000004
	attr_ACMODTIME   uint32 = 0x00000008
	attr_EXTENDED    uint32 = 0x80000000
)

// FileAttrs define the attributes for opening or creating file on the remote.
type FileAttrs struct {
	name        string
	flags       uint32
	size        uint64     // attr_SIZE
	uid         uint32     // attr_UIDGID
	gid         uint32     // attr_UIDGID
	permissions uint32     // attr_PERMISSIONS
	atime       uint32     // attr_ACMODTIME
	mtime       uint32     // attr_ACMODTIME
	exts        extensions // attr_EXTENDED
	fsMode      fs.FileMode
}

// NewFileAttrs create and initialize FileAttrs from FileInfo.
func NewFileAttrs(fi fs.FileInfo) (fa *FileAttrs) {
	fa = &FileAttrs{
		name: fi.Name(),
	}

	mode := fi.Mode()
	mtime := fi.ModTime()

	fa.SetSize(uint64(fi.Size()))
	fa.SetPermissions(uint32(mode.Perm()))
	fa.SetModifiedTime(uint32(mtime.Unix()))

	return fa
}

func newFileAttrs() (fa *FileAttrs) {
	return &FileAttrs{}
}

func unpackFileAttrs(payload []byte) (fa *FileAttrs, length int) {
	fa = &FileAttrs{}

	fa.flags = binary.BigEndian.Uint32(payload)
	payload = payload[4:]
	length += 4

	if fa.flags&attr_SIZE != 0 {
		fa.size = binary.BigEndian.Uint64(payload)
		payload = payload[8:]
		length += 8
	}
	if fa.flags&attr_UIDGID != 0 {
		fa.uid = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		length += 4
		fa.gid = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		length += 4
	}
	if fa.flags&attr_PERMISSIONS != 0 {
		fa.permissions = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		length += 4
		fa.updateFsmode()
	}
	if fa.flags&attr_ACMODTIME != 0 {
		fa.atime = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		length += 4
		fa.mtime = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		length += 4
	}
	if fa.flags&attr_EXTENDED != 0 {
		n := binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		length += 4

		fa.exts = make(extensions, n)
		for x := uint32(0); x < n; x++ {
			size := binary.BigEndian.Uint32(payload)
			payload = payload[4:]
			length += 4

			name := string(payload[:size])
			payload = payload[size:]
			length += int(size)

			size = binary.BigEndian.Uint32(payload)
			payload = payload[4:]
			length += 4

			data := string(payload[:size])
			payload = payload[size:]
			length += int(size)

			fa.exts[name] = data
		}
	}
	return fa, length
}

func (fa *FileAttrs) pack(w io.Writer) {
	_ = binary.Write(w, binary.BigEndian, fa.flags)

	if fa.flags&attr_SIZE != 0 {
		_ = binary.Write(w, binary.BigEndian, fa.size)
	}
	if fa.flags&attr_UIDGID != 0 {
		_ = binary.Write(w, binary.BigEndian, fa.uid)
		_ = binary.Write(w, binary.BigEndian, fa.gid)
	}
	if fa.flags&attr_PERMISSIONS != 0 {
		_ = binary.Write(w, binary.BigEndian, fa.permissions)
	}
	if fa.flags&attr_ACMODTIME != 0 {
		_ = binary.Write(w, binary.BigEndian, fa.atime)
		_ = binary.Write(w, binary.BigEndian, fa.mtime)
	}
	if fa.flags&attr_EXTENDED != 0 {
		n := uint32(len(fa.exts))
		_ = binary.Write(w, binary.BigEndian, n)
		for k, v := range fa.exts {
			_ = binary.Write(w, binary.BigEndian, uint32(len(k)))
			_, _ = w.Write([]byte(k))
			_ = binary.Write(w, binary.BigEndian, uint32(len(v)))
			_, _ = w.Write([]byte(v))
		}
	}
}

// AccessTime return the remote file access time.
func (fa *FileAttrs) AccessTime() uint32 {
	return fa.atime
}

// Extensions return the remote file attribute extensions as map of type and
// data.
func (fa *FileAttrs) Extensions() map[string]string {
	return map[string]string(fa.exts)
}

// Gid return the group ID attribute of file.
func (fa *FileAttrs) Gid() uint32 {
	return fa.gid
}

// IsDir return true if the file is a directory.
func (fa *FileAttrs) IsDir() bool {
	return fa.fsMode.IsDir()
}

// ModTime return the remote file modified time.
func (fa *FileAttrs) ModTime() time.Time {
	return time.Unix(int64(fa.mtime), 0)
}

// Mode return the file mode bits as standard fs.FileMode type.
func (fa *FileAttrs) Mode() fs.FileMode {
	return fa.fsMode
}

func (fa *FileAttrs) Name() string {
	return fa.name
}

// Permissions return the remote file mode and permissions.
func (fa *FileAttrs) Permissions() uint32 {
	return fa.permissions
}

// SetAccessTime set the file attribute access time.
func (fa *FileAttrs) SetAccessTime(v uint32) {
	fa.flags |= attr_ACMODTIME
	fa.atime = v
}

// SetExtension set the file attribute extension.
func (fa *FileAttrs) SetExtension(name, data string) {
	if fa.exts == nil {
		fa.exts = extensions{}
	}
	fa.flags |= attr_EXTENDED
	fa.exts[name] = data
}

// SetGid set the file attribute group ID.
func (fa *FileAttrs) SetGid(gid uint32) {
	fa.flags |= attr_UIDGID
	fa.gid = gid
}

// SetModifiedTime set the file attribute modified time.
func (fa *FileAttrs) SetModifiedTime(v uint32) {
	fa.flags |= attr_ACMODTIME
	fa.mtime = v
}

// SetPermissions set the remote file permission.
func (fa *FileAttrs) SetPermissions(v uint32) {
	fa.flags |= attr_PERMISSIONS
	fa.permissions = v
	fa.updateFsmode()
}

// SetSize set the remote file size.
func (fa *FileAttrs) SetSize(v uint64) {
	fa.flags |= attr_SIZE
	fa.size = v
}

// SetUid set the file attribute user ID.
func (fa *FileAttrs) SetUid(uid uint32) {
	fa.flags |= attr_UIDGID
	fa.uid = uid
}

// Size return the file size information.
func (fa *FileAttrs) Size() int64 {
	return int64(fa.size)
}

// Sys return the pointer to FileAttrs itself.
func (fa *FileAttrs) Sys() interface{} {
	return fa
}

// Uid return the user ID of file.
func (fa *FileAttrs) Uid() uint32 {
	return fa.uid
}

func (fa *FileAttrs) updateFsmode() {
	fa.fsMode = fs.FileMode(fa.permissions & 0777)
	switch fa.permissions & fileTypeMask {
	case fileTypeFifo:
		fa.fsMode |= fs.ModeNamedPipe
	case fileTypeCharDevice:
		fa.fsMode |= fs.ModeDevice | fs.ModeCharDevice
	case fileTypeDirectory:
		fa.fsMode |= fs.ModeDir
	case fileTypeBlockDevice:
		fa.fsMode |= fs.ModeDevice
	case fileTypeRegular:
		// NOOP
	case fileTypeSymlink:
		fa.fsMode |= fs.ModeSymlink
	case fileTypeSocket:
		fa.fsMode |= fs.ModeSocket
	}
	if fa.permissions&fileModeSetgid != 0 {
		fa.fsMode |= fs.ModeSetgid
	}
	if fa.permissions&fileModeSetuid != 0 {
		fa.fsMode |= fs.ModeSetuid
	}
	if fa.permissions&fileModeSticky != 0 {
		fa.fsMode |= fs.ModeSticky
	}
}
