// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maildir

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// An empty file name inside folder to indicate that the directory is a
// folder.
const fileMaildirFolder = `maildirfolder`

// Folder is a directory under maildir that store messages per file.
// A folder contains three directories: tmp, new, and cur; and an empty
// file "maildirfolder" to indicate a directory is a folder.
type Folder struct {
	name   string // The base name of folder.
	dir    string // The full path of folder inside maildir.
	dirCur string
	dirNew string
	dirTmp string
}

// CreateFolder create folder under directory maildir, populate all required
// sub directories and file.
// A folder must start with dot '.' and does not contains unicode control
// character.
func CreateFolder(maildir, name string) (folder *Folder, err error) {
	var logp = `CreateFolder`

	folder = &Folder{}

	folder.name, err = sanitizeFolderName(name)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	maildir = filepath.Clean(maildir)
	if maildir == `.` {
		return nil, fmt.Errorf(`%s: invalid maildir %q`, logp, maildir)
	}

	folder.dir = filepath.Join(maildir, name)

	err = os.Mkdir(folder.dir, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	err = folder.initDirs()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	var fileMdfolder = filepath.Join(folder.dir, fileMaildirFolder)

	err = os.WriteFile(fileMdfolder, nil, 0600)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return folder, nil
}

// NewFolder initialize folder from directory maildir.
// It will return an error if the one of the directory is not exist or does
// not have permission to write, or "maildirfolder" file is not exist.
func NewFolder(maildir, name string) (folder *Folder, err error) {
	var logp = `NewFolder`

	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return nil, fmt.Errorf(`%s: empty folder name`, logp)
	}

	folder = &Folder{
		name: name,
		dir:  filepath.Join(maildir, name),
	}

	folder.dirCur = filepath.Join(folder.dir, maildirCur)
	folder.dirNew = filepath.Join(folder.dir, maildirNew)
	folder.dirTmp = filepath.Join(folder.dir, maildirTmp)

	var fileMdfolder = filepath.Join(folder.dir, fileMaildirFolder)
	_, err = os.Stat(fileMdfolder)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	err = checkDir(folder.dirCur)
	if err != nil {
		return nil, fmt.Errorf(`%s: %s: %w`, logp, folder.dirCur, err)
	}
	err = checkDir(folder.dirNew)
	if err != nil {
		return nil, fmt.Errorf(`%s: %s: %w`, logp, folder.dirNew, err)
	}
	err = checkDir(folder.dirTmp)
	if err != nil {
		return nil, fmt.Errorf(`%s: %s: %w`, logp, folder.dirTmp, err)
	}

	return folder, nil
}

// Delete hard delete a message file in "cur".
// It will return no error if the file does not exist.
func (folder *Folder) Delete(file string) (err error) {
	file = strings.TrimSpace(file)
	if len(file) == 0 {
		// Prevent removing the cur directory.
		return nil
	}

	var (
		logp = `Delete`
		fdel = filepath.Join(folder.dirCur, file)
	)

	err = os.Remove(fdel)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf(`%s: %w`, logp, err)
	}
	return nil
}

// Fetch fetch the content of file from folder "new" directory.
// If the file exist, move it to the "cur" and add suffix ":2" to the file
// name.
// If the file does not exist, it will return nil without an error.
func (folder *Folder) Fetch(file string) (fileCur string, msg []byte, err error) {
	file = strings.TrimSpace(file)
	if len(file) == 0 {
		return ``, nil, nil
	}

	fileCur = file + `:2`

	var (
		logp    = `Fetch`
		pathNew = filepath.Join(folder.dirNew, file)
		pathCur = filepath.Join(folder.dirCur, fileCur)
	)

	msg, err = os.ReadFile(pathNew)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ``, nil, nil
		}
		return ``, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	err = os.Rename(pathNew, pathCur)
	if err != nil {
		return ``, nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return fileCur, msg, nil
}

// Get the content of file from "cur" directory.
// It will return nil without an error if file is not exist.
func (folder *Folder) Get(file string) (msg []byte, err error) {
	var (
		logp    = `Get`
		pathCur = filepath.Join(folder.dirCur, file)
	)
	msg, err = os.ReadFile(pathCur)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	return msg, nil
}

func (folder *Folder) initDirs() (err error) {
	var logp = `initDirs`

	folder.dirCur = filepath.Join(folder.dir, maildirCur)
	err = os.Mkdir(folder.dirCur, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	folder.dirNew = filepath.Join(folder.dir, maildirNew)
	err = os.Mkdir(folder.dirNew, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	folder.dirTmp = filepath.Join(folder.dir, maildirTmp)
	err = os.Mkdir(folder.dirTmp, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	return nil
}

// checkDir check if dir is exist and have permission to write.
func checkDir(dir string) (err error) {
	var fi os.FileInfo

	fi, err = os.Stat(dir)
	if err != nil {
		return err
	}

	var perm = fi.Mode().Perm()
	if perm&0700 != 0700 {
		// No permission to write and enter the directory.
		return os.ErrPermission
	}

	return nil
}

// sanitizeFolderName check if the folder name is not empty, begin with
// period '.' and does not contains unicode control characters.
func sanitizeFolderName(name string) (out string, err error) {
	out = strings.TrimSpace(name)
	if len(out) == 0 {
		return ``, errors.New(`folder name is empty`)
	}
	if len(out) == 1 && out[0] == '.' {
		return ``, errors.New(`folder name is empty`)
	}
	if out[0] != '.' {
		return ``, errors.New(`folder name must begin with period`)
	}
	if out[1] == '.' {
		return ``, errors.New(`folder name must not begin with ".."`)
	}
	var r rune
	for _, r = range out {
		if !unicode.IsPrint(r) {
			return ``, fmt.Errorf(`folder name contains unprintable character %q`, r)
		}
		if r == '/' {
			return ``, errors.New(`folder name must not contains slash '/'`)
		}
	}
	return out, nil
}
