// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
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

	"github.com/shuLhan/share/lib/ascii"
	libbytes "github.com/shuLhan/share/lib/bytes"
)

var (
	errOffset = errors.New("Seek: invalid offset")
	errWhence = errors.New("Seek: invalid whence")
)

//
// Node represent a single file.
//
type Node struct {
	os.FileInfo
	http.File

	SysPath         string      // The original file path in system.
	Path            string      // Absolute file path in memory.
	name            string      // File name.
	ContentType     string      // File type per MIME, for example "application/json".
	ContentEncoding string      // File type encoding, for example "gzip".
	modTime         time.Time   // ModTime contains file modification time.
	mode            os.FileMode // File mode.
	size            int64       // Size of file.
	V               []byte      // Content of file.
	Parent          *Node       // Pointer to parent directory.
	Childs          []*Node     // List of files in directory.
	plainv          []byte      // Content of file in plain text.
	lowerv          []byte      // Content of file in lower cases.
	off             int64       // The cursor position when doing Read or Seek.
	GenFuncName     string      // The function name for generated Go code.
}

//
// NewNode create a new node based on file information "fi".
// If maxFileSize is greater than zero, the file content and its type will be
// saved in node as V and ContentType.
//
func NewNode(parent *Node, fi os.FileInfo, maxFileSize int64) (node *Node, err error) {
	if fi == nil {
		return nil, nil
	}

	var (
		logp    = "NewNode"
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
		name:    fi.Name(),
		modTime: fi.ModTime(),
		mode:    fi.Mode(),
		size:    fi.Size(),
		V:       nil,
		Parent:  parent,
		Childs:  make([]*Node, 0),
	}
	node.generateFuncName(sysPath)

	if node.mode.IsDir() {
		node.size = 0
		return node, nil
	}

	// If the file is symbolic link, update the node size and mode based
	// on original.
	if fi.Mode()&os.ModeSymlink != 0 {
		absPath, err := filepath.EvalSymlinks(sysPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}

		fi, err = os.Lstat(absPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}

		node.mode = fi.Mode()
		if node.mode.IsDir() {
			node.size = 0
			return node, nil
		}
		node.size = fi.Size()
	}

	err = node.updateContent(maxFileSize)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	err = node.updateContentType()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	return node, nil
}

//
// AddChild add the other node as child of this node.
//
func (leaf *Node) AddChild(child *Node) {
	leaf.Childs = append(leaf.Childs, child)
	child.Parent = leaf
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
// Encode compress and set the content of Node.
//
func (leaf *Node) Encode(content []byte) (err error) {
	logp := "Node.Encode"

	leaf.plainv = content
	leaf.lowerv = bytes.ToLower(content)

	switch leaf.ContentEncoding {
	case EncodingGzip:
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)

		_, err = gz.Write(content)
		if err != nil {
			_ = gz.Close()
			return fmt.Errorf("%s: %w", logp, err)
		}

		err = gz.Close()
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}

		leaf.V = libbytes.Copy(buf.Bytes())

	default:
		leaf.V = content
	}
	return nil
}

func (leaf *Node) IsDir() bool {
	return leaf.mode.IsDir()
}

//
// MarshalJSON encode the node into JSON format.
// If the node is a file it will return the content of file;
// otherwise it will return the node with list of childs, but not including
// childs of childs.
//
func (leaf *Node) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	leaf.packAsJson(&buf, 0)
	return buf.Bytes(), nil
}

func (leaf *Node) ModTime() time.Time {
	return leaf.modTime
}

func (leaf *Node) Mode() os.FileMode {
	return leaf.mode
}

func (leaf *Node) Name() string {
	return leaf.name
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
	if leaf.off >= leaf.size {
		return 0, io.EOF
	}
	n = copy(p, leaf.V[leaf.off:])
	leaf.off += int64(n)
	return n, nil
}

//
// Readdir reads the contents of the directory associated with file and
// returns a slice of up to n FileInfo values, as would be returned by Lstat,
// in directory order.
// Subsequent calls on the same file will yield further FileInfos.
//
func (leaf *Node) Readdir(count int) (fis []os.FileInfo, err error) {
	if !leaf.IsDir() {
		return nil, nil
	}
	if count <= 0 || count >= len(leaf.Childs) {
		fis = make([]os.FileInfo, len(leaf.Childs))
		for x := 0; x < len(leaf.Childs); x++ {
			fis[x] = leaf.Childs[x]
		}
		leaf.off = 0
		return fis, nil
	}
	if leaf.off >= int64(len(leaf.Childs)) {
		return nil, nil
	}

	count += int(leaf.off)
	if count >= len(leaf.Childs) {
		count = len(leaf.Childs)
	}

	fis = make([]os.FileInfo, 0, count-int(leaf.off))

	for _, child := range leaf.Childs[leaf.off:count] {
		fis = append(fis, child)
	}

	leaf.off = int64(count)

	return fis, nil
}

//
// Save the content to file system and update the content of Node.
//
func (leaf *Node) Save(content []byte) (err error) {
	var (
		logp = "Node.Save"
		f    *os.File
	)
	f, err = os.OpenFile(leaf.SysPath, os.O_WRONLY|os.O_TRUNC, leaf.mode.Perm())
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	_, err = f.Write(content)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	err = leaf.Encode(content)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	leaf.modTime = time.Now()
	leaf.size = int64(len(content))
	return nil
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
		offset += leaf.size
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
// SetModTime set the file modification time.
//
func (leaf *Node) SetModTime(modTime time.Time) {
	leaf.modTime = modTime
}

//
// SetMode set the mode of file.
//
func (leaf *Node) SetMode(mode os.FileMode) {
	leaf.mode = mode
}

//
// SetName set the name of file.
//
func (leaf *Node) SetName(name string) {
	leaf.name = name
}

//
// SetSize set the file size.
//
func (leaf *Node) SetSize(size int64) {
	leaf.size = size
}

//
// Size return the file size information.
//
func (leaf *Node) Size() int64 {
	return leaf.size
}

//
// Stat return the file information.
//
func (leaf *Node) Stat() (os.FileInfo, error) {
	return leaf, nil
}

//
// Sys return the underlying data source (can return nil).
//
func (leaf *Node) Sys() interface{} {
	return leaf
}

//
// addChild add new node as sub-directory or file of this node.
//
func (leaf *Node) addChild(
	sysPath string, fi os.FileInfo, maxFileSize int64,
) (child *Node, err error) {
	child, err = NewNode(leaf, fi, maxFileSize)
	if err != nil {
		return nil, err
	}

	child.SysPath = sysPath

	leaf.Childs = append(leaf.Childs, child)

	return child, nil
}

func (leaf *Node) generateFuncName(in string) {
	syspath := string(libbytes.InReplace([]byte(in), []byte(ascii.LettersNumber), '_'))
	leaf.GenFuncName = "generate_" + syspath
}

func (leaf *Node) packAsJson(buf *bytes.Buffer, depth int) {
	isDir := leaf.IsDir()

	_ = buf.WriteByte('{')

	_, _ = fmt.Fprintf(buf, `%q:%q,`, "path", leaf.Path)
	_, _ = fmt.Fprintf(buf, `%q:%q,`, "name", leaf.name)
	_, _ = fmt.Fprintf(buf, `%q:%q,`, "content_type", leaf.ContentType)
	_, _ = fmt.Fprintf(buf, `%q:%d,`, "mod_time", leaf.modTime.Unix())
	_, _ = fmt.Fprintf(buf, `%q:%q,`, "mode_string", leaf.mode)
	_, _ = fmt.Fprintf(buf, `%q:%d,`, "size", leaf.size)
	_, _ = fmt.Fprintf(buf, `%q:%t,`, "is_dir", isDir)
	if !isDir && depth == 0 {
		content := base64.StdEncoding.EncodeToString(leaf.V)
		_, _ = fmt.Fprintf(buf, `%q:%q,`, "content", content)
	}

	_, _ = fmt.Fprintf(buf, `%q:`, "childs")
	if depth == 0 {
		_ = buf.WriteByte('[')
		for x, child := range leaf.Childs {
			if x > 0 {
				_ = buf.WriteByte(',')
			}
			child.packAsJson(buf, depth+1)
		}
		_ = buf.WriteByte(']')
	} else {
		_, _ = buf.WriteString("null")
	}
	_ = buf.WriteByte('}')
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
// resetAllModTime set the modTime of node and its child to the t.
// This method is only intended for testing.
//
func (leaf *Node) resetAllModTime(t time.Time) {
	leaf.modTime = t
	for _, c := range leaf.Childs {
		c.resetAllModTime(t)
	}
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
func (leaf *Node) update(newInfo os.FileInfo, maxFileSize int64) (err error) {
	if newInfo == nil {
		newInfo, err = os.Stat(leaf.SysPath)
		if err != nil {
			return fmt.Errorf("lib/memfs: Node.update %q: %w",
				leaf.Path, err)
		}
	}

	if leaf.mode != newInfo.Mode() {
		leaf.mode = newInfo.Mode()
		return nil
	}

	leaf.modTime = newInfo.ModTime()
	leaf.size = newInfo.Size()

	if newInfo.IsDir() {
		return nil
	}

	return leaf.updateContent(maxFileSize)
}

//
// updateContent read the content of file.
//
func (leaf *Node) updateContent(maxFileSize int64) (err error) {
	if maxFileSize < 0 {
		// Negative maxFileSize means content will not be read.
		return nil
	}
	if leaf.size > maxFileSize {
		return nil
	}
	if leaf.size == 0 {
		leaf.V = nil
		return nil
	}

	leaf.V, err = ioutil.ReadFile(leaf.SysPath)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}

	return nil
}

func (leaf *Node) updateContentType() error {
	leaf.ContentType = mime.TypeByExtension(path.Ext(leaf.name))
	if len(leaf.ContentType) > 0 {
		return nil
	}

	if len(leaf.V) > 0 {
		leaf.ContentType = http.DetectContentType(leaf.V)
		return nil
	}
	if leaf.size == 0 {
		// The actual file size is zero, we set the content type to
		// default.
		leaf.ContentType = defContentType
		return nil
	}

	data := make([]byte, 512)

	f, err := os.Open(leaf.SysPath)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// File is empty.
			leaf.ContentType = defContentType
			return nil
		}
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
