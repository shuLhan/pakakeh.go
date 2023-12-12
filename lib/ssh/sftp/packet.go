// Copyright 2021, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sftp

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	packetKindFxpInit byte = 1 + iota
	packetKindFxpVersion
	packetKindFxpOpen
	packetKindFxpClose
	packetKindFxpRead // 5
	packetKindFxpWrite
	packetKindFxpLstat
	packetKindFxpFstat
	packetKindFxpSetstat
	packetKindFxpFsetstat // 10
	packetKindFxpOpendir
	packetKindFxpReaddir
	packetKindFxpRemove
	packetKindFxpMkdir
	packetKindFxpRmdir // 15
	packetKindFxpRealpath
	packetKindFxpStat
	packetKindFxpRename
	packetKindFxpReadlink
	packetKindFxpSymlink // 20
)
const (
	packetKindFxpStatus = 101 + iota
	packetKindFxpHandle
	packetKindFxpData
	packetKindFxpName
	packetKindFxpAttrs
)

// TODO(ms): handle extended reply.
// const (
// packetKindFxpExtended = 200 + iota
// packetKindFxpExtendedReply
// )

type packet struct {
	// FxpHandle
	fh *FileHandle

	// FxpAttrs
	fa *FileAttrs

	exts extensions // from FxpVersion.

	message     string // from FxpStatus
	languageTag string // from FxpStatus

	// FxpData
	data []byte

	// FxpName
	nodes []*dirEntry

	version   uint32 // from FxpVersion.
	code      uint32 // from FxpStatus
	length    uint32
	requestID uint32

	kind byte
}

func unpackPacket(payload []byte) (pac *packet, err error) {
	logp := "packetUnpack"
	gotSize := uint32(len(payload))
	if gotSize < 9 {
		return nil, fmt.Errorf("%s: packet size too small %d", logp, gotSize)
	}

	pac = &packet{}

	pac.length = binary.BigEndian.Uint32(payload[:4])
	expSize := pac.length + 4
	if expSize != gotSize {
		return nil, fmt.Errorf("%s: expecting packet size %d, got %d", logp, expSize, gotSize)
	}
	pac.kind = payload[4]

	v := binary.BigEndian.Uint32(payload[5:])
	payload = payload[9:]
	if pac.kind == packetKindFxpVersion {
		pac.version = v
		pac.exts = unpackExtensions(payload)
		return pac, nil
	}

	pac.requestID = v

	switch pac.kind {
	case packetKindFxpStatus:
		pac.code = binary.BigEndian.Uint32(payload)
		payload = payload[4:]

		v = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		pac.message = string(payload[:v])
		payload = payload[v:]

		v = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		pac.languageTag = string(payload[:v])

	case packetKindFxpHandle:
		pac.fh = &FileHandle{}
		v = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		pac.fh.v = make([]byte, v)
		copy(pac.fh.v, payload[:v])

	case packetKindFxpData:
		v = binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		pac.data = payload[:v]

	case packetKindFxpName:
		var (
			length int
		)
		n := binary.BigEndian.Uint32(payload)
		payload = payload[4:]
		for x := uint32(0); x < n; x++ {
			node := &dirEntry{}

			v = binary.BigEndian.Uint32(payload)
			payload = payload[4:]
			node.fileName = string(payload[:v])
			payload = payload[v:]

			v = binary.BigEndian.Uint32(payload)
			payload = payload[4:]
			node.longName = string(payload[:v])
			payload = payload[v:]

			node.attrs, length = unpackFileAttrs(payload)
			node.attrs.name = node.fileName
			payload = payload[length:]

			pac.nodes = append(pac.nodes, node)
		}

	case packetKindFxpAttrs:
		pac.fa, _ = unpackFileAttrs(payload)
	}

	return pac, nil
}

func (pac *packet) fxpClose(fh *FileHandle) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpClose)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(fh.v)))
	_ = binary.Write(&buf, binary.BigEndian, fh.v)

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpFsetstat(fh *FileHandle, fa *FileAttrs) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpFsetstat)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(fh.v)))
	_ = binary.Write(&buf, binary.BigEndian, fh.v)
	fa.pack(&buf)

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpFstat(fh *FileHandle) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpFstat)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(fh.v)))
	_ = binary.Write(&buf, binary.BigEndian, fh.v)

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpInit(version uint32) []byte {
	var buf bytes.Buffer

	if version == 0 {
		version = defFxpVersion
	}
	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpInit)
	_ = binary.Write(&buf, binary.BigEndian, version)

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpLstat(remoteFile string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpLstat)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(remoteFile)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(remoteFile))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpMkdir(path string, fa *FileAttrs) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpMkdir)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(path)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(path))
	fa.pack(&buf)

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpOpen(filename string, pflags uint32, fa *FileAttrs) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpOpen)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(filename)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(filename))
	_ = binary.Write(&buf, binary.BigEndian, pflags)
	fa.pack(&buf)

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpOpendir(path string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpOpendir)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(path)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(path))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpRead(fh *FileHandle, offset uint64, length uint32) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpRead)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(fh.v)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(fh.v))
	_ = binary.Write(&buf, binary.BigEndian, offset)
	_ = binary.Write(&buf, binary.BigEndian, length)

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpReaddir(fh *FileHandle) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpReaddir)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(fh.v)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(fh.v))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpReadlink(linkPath string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpReadlink)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(linkPath)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(linkPath))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpRealpath(path string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpRealpath)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(path)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(path))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpRemove(remoteFile string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpRemove)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(remoteFile)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(remoteFile))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpRename(oldPath, newPath string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpRename)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(oldPath)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(oldPath))
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(newPath)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(newPath))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpRmdir(path string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpRmdir)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(path)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(path))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpSetstat(remoteFile string, fa *FileAttrs) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpSetstat)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(remoteFile)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(remoteFile))
	fa.pack(&buf)

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpStat(remoteFile string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpStat)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(remoteFile)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(remoteFile))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpSymlink(linkPath, targetPath string) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpSymlink)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(linkPath)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(linkPath))
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(targetPath)))
	_ = binary.Write(&buf, binary.BigEndian, []byte(targetPath))

	return sealPacket(buf.Bytes())
}

func (pac *packet) fxpWrite(fh *FileHandle, offset uint64, data []byte) []byte {
	var buf bytes.Buffer

	_ = binary.Write(&buf, binary.BigEndian, packetKindFxpWrite)
	_ = binary.Write(&buf, binary.BigEndian, pac.requestID)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(fh.v)))
	_ = binary.Write(&buf, binary.BigEndian, fh.v)
	_ = binary.Write(&buf, binary.BigEndian, offset)
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(data)))
	_ = binary.Write(&buf, binary.BigEndian, data)

	return sealPacket(buf.Bytes())
}

func sealPacket(in []byte) (out []byte) {
	lin := uint32(len(in))
	out = make([]byte, lin+4)
	binary.BigEndian.PutUint32(out, lin)
	copy(out[4:], in)
	return out
}
