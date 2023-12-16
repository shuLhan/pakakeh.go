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
	attrSize        uint32 = 0x00000001
	attrUIDGID      uint32 = 0x00000002
	attrPermissions uint32 = 0x00000004
	attrAcModtime   uint32 = 0x00000008
	attrExtended    uint32 = 0x80000000
)

// FileAttrs define the attributes for opening or creating file on the remote.
type FileAttrs struct {
	exts extensions // attrExtended

	name string

	fsMode fs.FileMode

	size uint64 // attrSize

	flags       uint32
	uid         uint32 // attrUIDGID
	gid         uint32 // attrUIDGID
	permissions uint32 // attrPermissions
	atime       uint32 // attrAcModtime
	mtime       uint32 // attrAcModtime
}

// NewFileAttrs create and initialize [FileAttrs] from [fs.FileInfo].
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

	if fa.flags&attrSize != 0 {
		fa.size = binary.BigEndian.Uint64(payload[:8])
		payload = payload[8:]
		length += 8
	}
	if fa.flags&attrUIDGID != 0 {
		fa.uid = binary.BigEndian.Uint32(payload[:4])
		payload = payload[4:]
		length += 4
		fa.gid = binary.BigEndian.Uint32(payload[:4])
		payload = payload[4:]
		length += 4
	}
	if fa.flags&attrPermissions != 0 {
		fa.permissions = binary.BigEndian.Uint32(payload[:4])
		payload = payload[4:]
		length += 4
		fa.updateFsmode()
	}
	if fa.flags&attrAcModtime != 0 {
		fa.atime = binary.BigEndian.Uint32(payload[:4])
		payload = payload[4:]
		length += 4
		fa.mtime = binary.BigEndian.Uint32(payload[:4])
		payload = payload[4:]
		length += 4
	}
	if fa.flags&attrExtended != 0 {
		n := binary.BigEndian.Uint32(payload[:4])
		payload = payload[4:]
		length += 4

		fa.exts = make(extensions, n)
		for x := uint32(0); x < n; x++ {
			size := binary.BigEndian.Uint32(payload[:4])
			payload = payload[4:]
			length += 4

			name := string(payload[:size])
			payload = payload[size:]
			length += int(size)

			size = binary.BigEndian.Uint32(payload[:4])
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

	if fa.flags&attrSize != 0 {
		_ = binary.Write(w, binary.BigEndian, fa.size)
	}
	if fa.flags&attrUIDGID != 0 {
		_ = binary.Write(w, binary.BigEndian, fa.uid)
		_ = binary.Write(w, binary.BigEndian, fa.gid)
	}
	if fa.flags&attrPermissions != 0 {
		_ = binary.Write(w, binary.BigEndian, fa.permissions)
	}
	if fa.flags&attrAcModtime != 0 {
		_ = binary.Write(w, binary.BigEndian, fa.atime)
		_ = binary.Write(w, binary.BigEndian, fa.mtime)
	}
	if fa.flags&attrExtended != 0 {
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

// Mode return the file mode bits as standard [fs.FileMode] type.
func (fa *FileAttrs) Mode() fs.FileMode {
	return fa.fsMode
}

// Name return the name of file.
func (fa *FileAttrs) Name() string {
	return fa.name
}

// Permissions return the remote file mode and permissions.
func (fa *FileAttrs) Permissions() uint32 {
	return fa.permissions
}

// SetAccessTime set the file attribute access time.
func (fa *FileAttrs) SetAccessTime(v uint32) {
	fa.flags |= attrAcModtime
	fa.atime = v
}

// SetExtension set the file attribute extension.
func (fa *FileAttrs) SetExtension(name, data string) {
	if fa.exts == nil {
		fa.exts = extensions{}
	}
	fa.flags |= attrExtended
	fa.exts[name] = data
}

// SetGid set the file attribute group ID.
func (fa *FileAttrs) SetGid(gid uint32) {
	fa.flags |= attrUIDGID
	fa.gid = gid
}

// SetModifiedTime set the file attribute modified time.
func (fa *FileAttrs) SetModifiedTime(v uint32) {
	fa.flags |= attrAcModtime
	fa.mtime = v
}

// SetPermissions set the remote file permission.
func (fa *FileAttrs) SetPermissions(v uint32) {
	fa.flags |= attrPermissions
	fa.permissions = v
	fa.updateFsmode()
}

// SetSize set the remote file size.
func (fa *FileAttrs) SetSize(v uint64) {
	fa.flags |= attrSize
	fa.size = v
}

// SetUid set the file attribute user ID.
func (fa *FileAttrs) SetUid(uid uint32) { //revive:disable-line
	fa.flags |= attrUIDGID
	fa.uid = uid
}

// Size return the file size information.
func (fa *FileAttrs) Size() int64 {
	return int64(fa.size)
}

// Sys return the pointer to [FileAttrs] itself.
// This method is added to comply with [fs.FileInfo] interface.
func (fa *FileAttrs) Sys() interface{} {
	return fa
}

// Uid return the user ID of file.
func (fa *FileAttrs) Uid() uint32 { //revive:disable-line
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
