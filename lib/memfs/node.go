// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"bytes"
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
	modTime time.Time // ModTime contains file modification time.

	os.FileInfo
	http.File

	Parent *Node // Pointer to parent directory.

	SysPath     string // The original file path in system.
	Path        string // Absolute file path in memory.
	name        string // File name.
	ContentType string // File type per MIME, for example "application/json".
	GenFuncName string // The function name for embedded Go code.

	Childs []*Node // List of files in directory.

	Content []byte // Content of file.
	plainv  []byte // Content of file in plain text.
	lowerv  []byte // Content of file in lower cases.

	size int64 // Size of file.
	off  int64 // The cursor position when doing Read or Seek.

	mode os.FileMode // File mode.
}

//
// NewNode create a new node based on file information "fi".
//
// The parent parameter is required to allow valid system path generated for
// new node.
//
// If maxFileSize is greater than zero, the file content and its type will be
// saved in node as Content and ContentType.
//
func NewNode(parent *Node, fi os.FileInfo, maxFileSize int64) (node *Node, err error) {
	if fi == nil {
		return nil, nil
	}

	var (
		logp    = "NewNode"
		sysPath string
		relPath string
	)

	sysPath = filepath.Join(parent.SysPath, fi.Name())
	relPath = path.Join(parent.Path, fi.Name())

	node = &Node{
		SysPath: sysPath,
		Path:    relPath,
		name:    fi.Name(),
		modTime: fi.ModTime(),
		mode:    fi.Mode(),
		size:    fi.Size(),
		Parent:  parent,
	}
	node.generateFuncName(sysPath)

	if node.mode.IsDir() {
		node.size = 0
		return node, nil
	}

	// If the file is symbolic link, update the node size and mode based
	// on original.
	if fi.Mode()&os.ModeSymlink != 0 {
		sysPath, err = filepath.EvalSymlinks(sysPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}

		fi, err = os.Lstat(sysPath)
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
func (node *Node) AddChild(child *Node) {
	if child.modTime.IsZero() {
		child.modTime = time.Now()
	}
	node.Childs = append(node.Childs, child)
	child.Parent = node
}

//
// Close reset the offset position back to zero.
//
func (node *Node) Close() error {
	node.off = 0
	return nil
}

func (node *Node) IsDir() bool {
	return node.mode.IsDir()
}

//
// MarshalJSON encode the node into JSON format.
// If the node is a file it will return the content of file;
// otherwise it will return the node with list of childs, but not including
// childs of childs.
//
func (node *Node) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	node.packAsJson(&buf, 0)
	return buf.Bytes(), nil
}

func (node *Node) ModTime() time.Time {
	return node.modTime
}

func (node *Node) Mode() os.FileMode {
	return node.mode
}

func (node *Node) Name() string {
	return node.name
}

//
// Read the content of node into p.
//
func (node *Node) Read(p []byte) (n int, err error) {
	// Implementations of Read are discouraged from returning a zero byte
	// count with a nil error, except when len(p) == 0.
	if len(p) == 0 {
		return 0, nil
	}
	if node.off >= node.size {
		return 0, fmt.Errorf("Read: %w", io.EOF)
	}
	n = copy(p, node.Content[node.off:])
	node.off += int64(n)
	return n, nil
}

//
// Readdir reads the contents of the directory associated with file and
// returns a slice of up to n FileInfo values, as would be returned by Lstat,
// in directory order.
// Subsequent calls on the same file will yield further FileInfos.
//
func (node *Node) Readdir(count int) (fis []os.FileInfo, err error) {
	if !node.IsDir() {
		return nil, nil
	}
	if count <= 0 || count >= len(node.Childs) {
		fis = make([]os.FileInfo, len(node.Childs))
		for x := 0; x < len(node.Childs); x++ {
			fis[x] = node.Childs[x]
		}
		node.off = 0
		return fis, nil
	}
	if node.off >= int64(len(node.Childs)) {
		return nil, nil
	}

	count += int(node.off)
	if count >= len(node.Childs) {
		count = len(node.Childs)
	}

	fis = make([]os.FileInfo, 0, count-int(node.off))

	for _, child := range node.Childs[node.off:count] {
		fis = append(fis, child)
	}

	node.off = int64(count)

	return fis, nil
}

//
// Save the content to file system and update the content of Node.
//
func (node *Node) Save(content []byte) (err error) {
	var (
		logp = "Save"
		f    *os.File
	)
	f, err = os.OpenFile(node.SysPath, os.O_WRONLY|os.O_TRUNC, node.mode.Perm())
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

	node.Content = content
	node.modTime = time.Now()
	node.size = int64(len(content))
	return nil
}

//
// Seek sets the offset for the next Read offset, interpreted according to
// whence: SeekStart means relative to the start of the file, SeekCurrent
// means relative to the current offset, and SeekEnd means relative to the
// end. Seek returns the new offset relative to the start of the file and an
// error, if any.
//
func (node *Node) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
	case io.SeekCurrent:
		offset += node.off
	case io.SeekEnd:
		offset += node.size
	default:
		return 0, errWhence
	}
	if offset < 0 {
		return 0, errOffset
	}
	node.off = offset
	return node.off, nil
}

//
// SetModTime set the file modification time.
//
func (node *Node) SetModTime(modTime time.Time) {
	node.modTime = modTime
}

//
// SetModTimeUnix set the file modification time using seconds and nanoseconds
// since January 1, 1970 UTC.
//
func (node *Node) SetModTimeUnix(seconds, nanoSeconds int64) {
	node.modTime = time.Unix(seconds, nanoSeconds)
}

//
// SetMode set the mode of file.
//
func (node *Node) SetMode(mode os.FileMode) {
	node.mode = mode
}

//
// SetName set the name of file.
//
func (node *Node) SetName(name string) {
	node.name = name
}

//
// SetSize set the file size.
//
func (node *Node) SetSize(size int64) {
	node.size = size
}

//
// Size return the file size information.
//
func (node *Node) Size() int64 {
	return node.size
}

//
// Stat return the file information.
//
func (node *Node) Stat() (os.FileInfo, error) {
	return node, nil
}

//
// Sys return the underlying data source (can return nil).
//
func (node *Node) Sys() interface{} {
	return node
}

//
// addChild add new node as sub-directory or file of this node.
//
func (node *Node) addChild(
	sysPath string, fi os.FileInfo, maxFileSize int64,
) (child *Node, err error) {
	child, err = NewNode(node, fi, maxFileSize)
	if err != nil {
		return nil, fmt.Errorf("addChild: %w", err)
	}

	child.SysPath = sysPath

	node.Childs = append(node.Childs, child)

	return child, nil
}

func (node *Node) generateFuncName(in string) {
	syspath := string(libbytes.InReplace([]byte(in), []byte(ascii.LettersNumber), '_'))
	node.GenFuncName = "generate_" + syspath
}

func (node *Node) packAsJson(buf *bytes.Buffer, depth int) {
	isDir := node.IsDir()

	_ = buf.WriteByte('{')

	_, _ = fmt.Fprintf(buf, `"path":%q,`, node.Path)
	_, _ = fmt.Fprintf(buf, `"name":%q,`, node.name)
	_, _ = fmt.Fprintf(buf, `"content_type":%q,`, node.ContentType)
	_, _ = fmt.Fprintf(buf, `"mod_time":%d,`, node.modTime.Unix())
	_, _ = fmt.Fprintf(buf, `"mode_string":%q,`, node.mode)
	_, _ = fmt.Fprintf(buf, `"size":%d,`, node.size)
	_, _ = fmt.Fprintf(buf, `"is_dir":%t,`, isDir)
	if !isDir && depth == 0 {
		content := base64.StdEncoding.EncodeToString(node.Content)
		_, _ = fmt.Fprintf(buf, `"content":%q,`, content)
	}

	_, _ = fmt.Fprintf(buf, `"childs":`)
	if depth == 0 {
		_ = buf.WriteByte('[')
		for x, child := range node.Childs {
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
func (node *Node) removeChild(child *Node) *Node {
	for x := 0; x < len(node.Childs); x++ {
		if node.Childs[x] != child {
			continue
		}

		copy(node.Childs[x:], node.Childs[x+1:])
		n := len(node.Childs)
		node.Childs[n-1] = nil
		node.Childs = node.Childs[:n-1]

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
func (node *Node) resetAllModTime(t time.Time) {
	node.modTime = t
	for _, c := range node.Childs {
		c.resetAllModTime(t)
	}
}

//
// Update the node metadata or content based on new file information.
//
// The newInfo parameter is optional, if its nil, it will read the file
// information based on node's SysPath.
//
// The maxFileSize parameter is also optional.
// If its negative, the node content will not be updated.
// If its zero, it will default to 5 MB.
//
// There are two possible changes that will happen: its either change on
// mode or change on content (size and modtime).
// Change on mode will not affect the content of node.
//
func (node *Node) Update(newInfo os.FileInfo, maxFileSize int64) (err error) {
	logp := "Update"

	if newInfo == nil {
		newInfo, err = os.Stat(node.SysPath)
		if err != nil {
			return fmt.Errorf("%s: %s: %w", logp, node.SysPath, err)
		}
	}

	if node.mode != newInfo.Mode() {
		node.mode = newInfo.Mode()
		return nil
	}

	node.modTime = newInfo.ModTime()
	node.size = newInfo.Size()

	if newInfo.IsDir() {
		return nil
	}

	err = node.updateContent(maxFileSize)
	if err != nil {
		return fmt.Errorf("%s: %s: %w", logp, node.SysPath, err)
	}
	return nil
}

//
// updateContent read the content of file.
//
func (node *Node) updateContent(maxFileSize int64) (err error) {
	if maxFileSize < 0 {
		// Negative maxFileSize means content will not be read.
		return nil
	} else if maxFileSize == 0 {
		maxFileSize = defaultMaxFileSize
	}
	if node.size > maxFileSize {
		return nil
	}
	if node.size == 0 {
		node.Content = nil
		return nil
	}

	node.Content, err = ioutil.ReadFile(node.SysPath)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("updateContent: %w", err)
	}

	return nil
}

func (node *Node) updateContentType() error {
	node.ContentType = mime.TypeByExtension(path.Ext(node.name))
	if len(node.ContentType) > 0 {
		return nil
	}

	if len(node.Content) > 0 {
		node.ContentType = http.DetectContentType(node.Content)
		return nil
	}
	if node.size == 0 {
		// The actual file size is zero, we set the content type to
		// default.
		node.ContentType = defContentType
		return nil
	}

	logp := "updateContentType"
	data := make([]byte, 512)

	f, err := os.Open(node.SysPath)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// File is empty.
			node.ContentType = defContentType
			return nil
		}
		return fmt.Errorf("%s: %w", logp, err)
	}

	_, err = f.Read(data)
	if err != nil {
		errc := f.Close()
		if errc != nil {
			panic(errc)
		}
		return fmt.Errorf("%s: %w", logp, err)
	}

	err = f.Close()
	if err != nil {
		err = fmt.Errorf("%s: %w", logp, err)
		panic(err)
	}

	node.ContentType = http.DetectContentType(data)

	return nil
}
