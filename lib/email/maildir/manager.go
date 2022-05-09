// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package maildir provide a library to manage email using maildir format.
//
// # References
//
// [1] http://www.qmail.org/qmail-manual-html/man5/maildir.html
//
// [2] https://cr.yp.to/proto/maildir.html
package maildir

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	libtime "github.com/shuLhan/share/lib/time"
)

// Manager manage email in a directory.
type Manager struct {
	dirCur   string
	dirNew   string
	dirOut   string
	dirTmp   string
	hostname string
	pid      int
	counter  int
}

// New create new maildir Manager in directory and initialize the hostname,
// pid, and counter for generating unique name.
func New(dir string) (mg *Manager, err error) {
	if len(dir) == 0 {
		return nil, fmt.Errorf("email/maildir: New: empty base directory")
	}

	mg = &Manager{
		pid: os.Getpid(),
	}

	err = mg.initDirs(dir)
	if err != nil {
		return nil, err
	}

	mg.hostname, err = os.Hostname()
	if len(mg.hostname) == 0 && err != nil {
		mg.hostname = os.Getenv("HOST")
		if len(mg.hostname) == 0 {
			mg.hostname = "localhost"
		}
	}

	return mg, nil
}

func (mg *Manager) initDirs(dir string) (err error) {
	mg.dirCur = filepath.Join(dir, "cur")
	err = os.MkdirAll(mg.dirCur, 0750)
	if err != nil {
		return fmt.Errorf("email/maildir: initDirs: %s", err.Error())
	}

	mg.dirNew = filepath.Join(dir, "new")
	err = os.MkdirAll(mg.dirNew, 0750)
	if err != nil {
		return fmt.Errorf("email/maildir: initDirs: %s", err.Error())
	}

	mg.dirOut = filepath.Join(dir, "out")
	err = os.MkdirAll(mg.dirOut, 0700)
	if err != nil {
		return fmt.Errorf("email/maildir: initDirs: %s", err.Error())
	}

	mg.dirTmp = filepath.Join(dir, "tmp")
	err = os.MkdirAll(mg.dirTmp, 0700)
	if err != nil {
		return fmt.Errorf("email/maildir: initDirs: %s", err.Error())
	}

	return nil
}

// Delete email file in "cur".
func (mg *Manager) Delete(fname string) (err error) {
	if len(fname) == 0 {
		return fmt.Errorf("email/maildir: Delete: empty file name")
	}

	fdel := filepath.Join(mg.dirCur, fname)

	err = os.Remove(fdel)
	if err != nil {
		return fmt.Errorf("email/maildir: Delete: %s", err.Error())
	}

	return nil
}

// DeleteOutQueue delete temporary file in send queue.
func (mg *Manager) DeleteOutQueue(fname string) (err error) {
	if len(fname) == 0 {
		return nil
	}

	fname = filepath.Join(mg.dirOut, fname)

	err = os.Remove(fname)
	if err != nil {
		return fmt.Errorf("email/maildir: DeleteOutQueue: %s", err.Error())
	}

	return nil
}

// OutQueue save the email in temporary queue directory before sending it to
// external MTA or processed.
//
// When mail is coming from MUA and received by server, the mail need
// to be successfully stored into disk by server, before replying with
// "250 OK" to client.
func (mg *Manager) OutQueue(email []byte) (err error) {
	if len(email) == 0 {
		return nil
	}

	fname, _, err := mg.generateUniqueName(mg.dirOut)
	if err != nil {
		return err
	}

	err = os.WriteFile(fname, email, 0400)
	if err != nil {
		err = fmt.Errorf("email/maildir: OutQueue: %s", err.Error())
		return err
	}

	mg.counter++

	return nil
}

// Get will move email from "new" to "cur".
func (mg *Manager) Get(fname string) (err error) {
	if len(fname) == 0 {
		return nil
	}

	src := filepath.Join(mg.dirNew, fname)
	dst := filepath.Join(mg.dirCur, fname)

	err = os.Rename(src, dst)
	if err != nil {
		return fmt.Errorf("email/maildir: Read: %s", err.Error())
	}

	return nil
}

// Incoming save incoming message, from external MTA, in directory
// "${dir}/tmp/${unique}".  Upon success, hard link it to
// "${dir}/new/${unique}" and delete the temporary file.
func (mg *Manager) Incoming(email []byte) (err error) {
	if len(email) == 0 {
		return nil
	}

	tmpFile, uniqueName, err := mg.generateUniqueName(mg.dirTmp)
	if err != nil {
		return err
	}

	err = os.WriteFile(tmpFile, email, 0660)
	if err != nil {
		err = fmt.Errorf("email/maildir: Incoming: %s", err.Error())
		return err
	}

	newFile := filepath.Join(mg.dirNew, uniqueName)

	err = os.Link(tmpFile, newFile)
	if err != nil {
		_ = os.Remove(tmpFile)
		err = fmt.Errorf("email/maildir: Incoming: %s", err.Error())
		return err
	}

	err = os.Remove(tmpFile)
	if err != nil {
		log.Printf("email/maildir: Incoming: %s", err.Error())
	}

	mg.counter++

	return nil
}

// RemoveAll remove all files inside a directory.
func (mg *Manager) RemoveAll(dir string) {
	d, err := os.Open(dir)
	if err != nil {
		log.Println("email/maildir: RemoveAll: " + err.Error())
		return
	}
	fis, err := d.Readdir(0)
	if err != nil {
		log.Println("email/maildir: RemoveAll: " + err.Error())
		return
	}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		file := filepath.Join(dir, fi.Name())
		err = os.Remove(file)
		if err != nil {
			log.Println("email/maildir: RemoveAll: " + err.Error())
		}
	}
}

// generateUniqueName try generate unique name until 5 attempts or return an
// error.
func (mg *Manager) generateUniqueName(dir string) (fname, uniqueName string, err error) {
	x := 0
	for x < 5 {
		uniqueName = mg.uniqueName()
		fname = filepath.Join(dir, uniqueName)
		_, err = os.Stat(fname)
		if err != nil {
			if os.IsNotExist(err) {
				return fname, uniqueName, nil
			}
		}
		time.Sleep(2 * time.Second)
		x++
	}

	err = fmt.Errorf("email/maildir: OutQueue: %s", err.Error())

	return "", "", err
}

// uniqueName generate a unique name using the following format,
//
//	UnixTimestamp "." "M"(microsecond) "P"(ProcessID) "Q"(Counter) "."
//	hostname
func (mg *Manager) uniqueName() string {
	now := time.Now()

	return fmt.Sprintf("%d.M%dP%dQ%d.%s", now.Unix(),
		libtime.Microsecond(&now), mg.pid, mg.counter, mg.hostname)
}
