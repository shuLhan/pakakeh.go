// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package maildir

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// List of maildir directories.
const (
	maildirCur = `cur`
	maildirNew = `new`
	maildirTmp = `tmp`
)

// Manager manage messages and folders in single file system.
// This is the main Maildir.
type Manager struct {
	folders  map[string]*Folder
	dir      string
	dirCur   string
	dirNew   string
	dirTmp   string
	hostname string
	counter  int64
	pid      int
}

// NewManager create new maildir Manager in directory and initialize the hostname,
// pid, and counter for generating unique name.
func NewManager(dir string) (mg *Manager, err error) {
	var logp = `NewManager`

	dir = strings.TrimSpace(dir)
	if len(dir) == 0 {
		return nil, fmt.Errorf(`%s: empty base directory`, logp)
	}

	mg = &Manager{
		folders: map[string]*Folder{},
		dir:     dir,
		pid:     osGetpid(),
	}

	err = mg.initDirs(dir)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	err = mg.scanFolders()
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	mg.hostname, err = osHostname()
	if err != nil || len(mg.hostname) == 0 {
		mg.hostname = os.Getenv(`HOST`)
		if len(mg.hostname) == 0 {
			mg.hostname = `localhost`
		}
	}
	return mg, nil
}

// initDirs initialize the maildir directories.
func (mg *Manager) initDirs(dir string) (err error) {
	var logp = `initDirs`

	mg.dirCur = filepath.Join(dir, maildirCur)
	err = os.MkdirAll(mg.dirCur, 0750)
	if err != nil {
		return fmt.Errorf(`%s: %s`, logp, err)
	}

	mg.dirNew = filepath.Join(dir, maildirNew)
	err = os.MkdirAll(mg.dirNew, 0750)
	if err != nil {
		return fmt.Errorf(`%s: %s`, logp, err)
	}

	mg.dirTmp = filepath.Join(dir, maildirTmp)
	err = os.MkdirAll(mg.dirTmp, 0700)
	if err != nil {
		return fmt.Errorf(`%s: %s`, logp, err)
	}

	return nil
}

// scanFolders scan folders inside the main maildir.
// A folder name begin with '.' and contains empty file named `maildirfolder`.
func (mg *Manager) scanFolders() (err error) {
	var (
		logp = `scanFolders`

		listEntry []os.DirEntry
	)

	listEntry, err = os.ReadDir(mg.dir)
	if err != nil {
		return fmt.Errorf(`%s: %w`, logp, err)
	}

	var (
		entry        os.DirEntry
		name         string
		fileMdfolder string
		folder       *Folder
	)
	for _, entry = range listEntry {
		if !entry.IsDir() {
			continue
		}

		name = entry.Name()
		if name == `.` || name == `..` {
			continue
		}
		if name[0] != '.' {
			continue
		}

		fileMdfolder = filepath.Join(mg.dir, name, fileMaildirFolder)
		_, err = os.Stat(fileMdfolder)
		if err != nil {
			continue
		}

		folder, err = NewFolder(mg.dir, name)
		if err != nil {
			continue
		}
		mg.folders[name] = folder
	}

	return nil
}

// Delete hard delete a message file in "cur".
// It will return no error if the file does not exist.
func (mg *Manager) Delete(fnCur string) (err error) {
	fnCur = strings.TrimSpace(fnCur)
	if len(fnCur) == 0 {
		// Prevent removing the cur directory.
		return nil
	}

	var (
		logp = `Delete`
		fdel = filepath.Join(mg.dirCur, fnCur)
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

// FetchNew fetch the message from "new" directory by its file name.
// This operation move email from "new" to "cur" with prefix ":2" added to
// file name.
// It will return nil without an error if string fnNew is zero or file does
// not exist.
func (mg *Manager) FetchNew(fnNew string) (fnCur string, msg []byte, err error) {
	fnNew = strings.TrimSpace(fnNew)
	if len(fnNew) == 0 {
		return ``, nil, nil
	}

	fnCur = fnNew

	var (
		logp    = `FetchNew`
		pathNew = filepath.Join(mg.dirNew, fnNew)
		pathCur = filepath.Join(mg.dirCur, fnCur)
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

	return fnCur, msg, nil
}

// Incoming save message received from external MTA in directory
// "${dir}/tmp/${unique}".
// Upon success, hard link it to "${dir}/new/${unique}" and delete the
// temporary file, and return the path of new file.
func (mg *Manager) Incoming(msg []byte) (fnNew string, err error) {
	var logp = `Incoming`

	if len(msg) == 0 {
		return ``, fmt.Errorf(`%s: empty message`, logp)
	}

	var (
		fname   = createFilename(mg.pid, mg.counter, mg.hostname)
		pathTmp = filepath.Join(mg.dirTmp, fname.nameTmp)
	)

	_, err = os.Stat(pathTmp)
	if err != nil {
		if !os.IsNotExist(err) {
			return ``, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	err = os.WriteFile(pathTmp, msg, 0660)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	fnNew, err = fname.generateNameNew(pathTmp, int64(len(msg)))
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	var pathNew = filepath.Join(mg.dirNew, fnNew)

	err = os.Link(pathTmp, pathNew)
	if err != nil {
		_ = os.Remove(pathTmp)
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	err = os.Remove(pathTmp)
	if err != nil {
		log.Printf(`%s: %s`, logp, err)
	}

	mg.counter++

	return fnNew, nil
}

// OutgoingQueue save the message in temporary queue directory before sending
// it to external MTA or processed.
//
// When mail is coming from MUA and received by server, the mail need
// to be successfully stored into disk by server, before replying with
// "250 OK" to client.
//
// On success it will return the file name.
func (mg *Manager) OutgoingQueue(msg []byte) (nameTmp string, err error) {
	var logp = `OutgoingQueue`

	if len(msg) == 0 {
		return ``, fmt.Errorf(`%s: empty message`, logp)
	}

	var (
		fname   = createFilename(mg.pid, mg.counter, mg.hostname)
		pathTmp = filepath.Join(mg.dirTmp, fname.nameTmp)
	)

	_, err = os.Stat(pathTmp)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return ``, fmt.Errorf(`%s: %w`, logp, err)
		}
	}

	err = os.WriteFile(pathTmp, msg, 0400)
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	mg.counter++
	nameTmp = fname.nameTmp

	return nameTmp, nil
}
