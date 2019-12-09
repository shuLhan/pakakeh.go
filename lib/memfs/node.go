// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

var (
	errOffset = errors.New("Seek: invalid offset")
	errWhence = errors.New("Seek: invalid whence")
)

//
// Node represent a single file.
//
type Node struct {
	SysPath         string      // The original file path in system.
	Path            string      // Absolute file path in memory.
	Name            string      // File name.
	ContentType     string      // File type per MIME, for example "application/json".
	ContentEncoding string      // File type encoding, for example "gzip".
	ModTime         time.Time   // ModTime contains file modification time.
	Mode            os.FileMode // File mode.
	Size            int64       // Size of file.
	V               []byte      // Content of file.
	Parent          *Node       // Pointer to parent directory.
	Childs          []*Node     // List of files in directory.
	plainv          []byte      // Content of file in plain text.
	lowerv          []byte      // Content of file in lower cases.
	off             int64       // The cursor position when doing Read or Seek.
}

//
// NewNode create a new node based on file information "fi".
// If withContent is true, the file content and its type will be saved in
// node as V and ContentType.
//
func NewNode(parent *Node, fi os.FileInfo, withContent bool) (node *Node, err error) {
	if fi == nil {
		return nil, nil
	}

	var (
		sysPath string
		absPath string
	)

	if parent != nil {
		sysPath = filepath.Join(parent.SysPath, fi.Name())
		absPath = path.Join(parent.Path, fi.Name())
	} else {
		sysPath = fi.Name()
		absPath = fi.Name()
	}

	node = &Node{
		SysPath: sysPath,
		Path:    absPath,
		Name:    fi.Name(),
		ModTime: fi.ModTime(),
		Mode:    fi.Mode(),
		Size:    fi.Size(),
		V:       nil,
		Parent:  parent,
		Childs:  make([]*Node, 0),
	}

	if node.Mode.IsDir() || !withContent {
		node.Size = 0
		return node, nil
	}

	err = node.updateContent()
	if err != nil {
		return nil, err
	}

	err = node.updateContentType()
	if err != nil {
		return nil, err
	}

	return node, nil
}

//
// Close reset the offset position back to zero.
//
func (leaf *Node) Close() error {
	leaf.off = 0
	return nil
}

//
// Decode the contents of node (for example, uncompress with gzip) and return
// it.
//
func (leaf *Node) Decode() ([]byte, error) {
	if len(leaf.ContentEncoding) == 0 {
		leaf.plainv = leaf.V
		return leaf.plainv, nil
	}

	leaf.plainv = leaf.plainv[:0]

	if leaf.ContentEncoding == EncodingGzip {
		r, err := gzip.NewReader(bytes.NewReader(leaf.V))
		if err != nil {
			return nil, err
		}

		buf := make([]byte, 1024)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				leaf.plainv = append(leaf.plainv, buf[:n]...)
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			buf = buf[0:]
		}
	}

	return leaf.plainv, nil
}

//
// Read the content of node into p.
//
func (leaf *Node) Read(p []byte) (n int, err error) {
	// Implementations of Read are discouraged from returning a zero byte
	// count with a nil error, except when len(p) == 0.
	if len(p) == 0 {
		return 0, nil
	}
	if leaf.off >= leaf.Size {
		return 0, io.EOF
	}
	n = copy(p, leaf.V[leaf.off:])
	leaf.off += int64(n)
	return n, nil
}

//
// Seek sets the offset for the next Read offset, interpreted according to
// whence: SeekStart means relative to the start of the file, SeekCurrent
// means relative to the current offset, and SeekEnd means relative to the
// end. Seek returns the new offset relative to the start of the file and an
// error, if any.
//
func (leaf *Node) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += leaf.off
	case io.SeekEnd:
		offset += leaf.Size
	default:
		return 0, errWhence
	}
	if offset < 0 {
		return 0, errOffset
	}
	leaf.off = offset
	return leaf.off, nil
}

//
// removeChild remove a children node from list.  If child is not exist, it
// will return nil.
//
func (leaf *Node) removeChild(child *Node) *Node {
	for x := 0; x < len(leaf.Childs); x++ {
		if leaf.Childs[x] != child {
			continue
		}

		copy(leaf.Childs[x:], leaf.Childs[x+1:])
		n := len(leaf.Childs)
		leaf.Childs[n-1] = nil
		leaf.Childs = leaf.Childs[:n-1]

		child.Parent = nil
		child.Childs = nil

		return child
	}

	return nil
}

//
// update the node content and information based on new file information.
//
// If the newInfo is nil, it will read the file information based on node's
// SysPath.
//
// There are two possible changes that will happened: its either change on
// mode or change on content (size and modtime).
// Change on mode will not affect the content of node.
//
func (leaf *Node) update(newInfo os.FileInfo, withContent bool) (err error) {
	if newInfo == nil {
		newInfo, err = os.Stat(leaf.SysPath)
		if err != nil {
			return fmt.Errorf("lib/memfs: Node.update %q: %s",
				leaf.Path, err.Error())
		}
	}

	if leaf.Mode != newInfo.Mode() {
		leaf.Mode = newInfo.Mode()
		return nil
	}

	leaf.ModTime = newInfo.ModTime()
	leaf.Size = newInfo.Size()

	if !withContent || newInfo.IsDir() {
		return nil
	}

	return leaf.updateContent()
}

//
// updateContent read the content of file.
//
func (leaf *Node) updateContent() (err error) {
	if leaf.Size > MaxFileSize {
		return nil
	}

	leaf.V, err = ioutil.ReadFile(leaf.SysPath)
	if err != nil {
		return err
	}

	return nil
}

func (leaf *Node) updateContentType() error {
	leaf.ContentType = mime.TypeByExtension(path.Ext(leaf.Name))
	if len(leaf.ContentType) > 0 {
		return nil
	}

	if len(leaf.V) > 0 {
		leaf.ContentType = http.DetectContentType(leaf.V)
		return nil
	}

	data := make([]byte, 512)

	f, err := os.Open(leaf.SysPath)
	if err != nil {
		return err
	}

	_, err = f.Read(data)
	if err != nil {
		errc := f.Close()
		if errc != nil {
			panic(errc)
		}
		return err
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}

	leaf.ContentType = http.DetectContentType(data)

	return nil
}
