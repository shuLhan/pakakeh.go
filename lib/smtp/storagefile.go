// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smtp

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

const (
	defDirBounce = "bounce"
	defDirSpool  = "/var/spool/smtpd"
)

//
// StorageFile implement the Storage interface where mail object is save and
// retrieved in file system inside a directory.
//
type StorageFile struct {
	dir  string
	buff bytes.Buffer
	enc  *gob.Encoder
	dec  *gob.Decoder
}

//
// NewStorageFile create and initialize new file storage.  If directory is
// empty, the default storage is located at "/var/spool/smtpd/".
//
func NewStorageFile(dir string) (fs *StorageFile, err error) {
	if len(dir) == 0 {
		dir = defDirSpool
	}

	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, err
	}

	dirBounce := filepath.Join(dir, defDirBounce)

	err = os.MkdirAll(dirBounce, 0700)
	if err != nil {
		return nil, err
	}

	fs = &StorageFile{
		dir: dir,
	}

	fs.enc = gob.NewEncoder(&fs.buff)
	fs.dec = gob.NewDecoder(&fs.buff)

	return fs, nil
}

//
// Delete the mail object on file system by ID.
//
func (fs *StorageFile) Delete(id string) (err error) {
	if len(id) == 0 {
		return
	}

	fpath := filepath.Join(fs.dir, id)
	return os.Remove(fpath)
}

//
// Load read the mail object from file system by ID.
//
func (fs *StorageFile) Load(id string) (mail *MailTx, err error) {
	if len(id) == 0 {
		return
	}

	fpath := filepath.Join(fs.dir, id)

	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	return fs.loadRaw(b)
}

//
// LoadAll mail objects from file system.
//
func (fs *StorageFile) LoadAll() (mails []*MailTx, err error) {
	d, err := os.Open(fs.dir)
	if err != nil {
		return nil, err
	}

	fis, err := d.Readdir(0)
	if err != nil {
		return nil, err
	}

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		mail, err := fs.Load(fi.Name())
		if err != nil {
			log.Printf("StorageFile.Load: %s\n", err)
			continue
		}

		mails = append(mails, mail)
	}

	return mails, nil
}

//
// Bounce move the incoming mail to bounced state.  In this storage
// service, the mail file is moved to "{dir}/bounce".
//
func (fs *StorageFile) Bounce(id string) error {
	oldp := filepath.Join(fs.dir, id)
	newp := filepath.Join(fs.dir, defDirBounce, id)

	return os.Rename(oldp, newp)
}

//
// Store the mail object into file system.
//
func (fs *StorageFile) Store(mail *MailTx) (err error) {
	if mail == nil {
		return
	}

	fs.buff.Reset()
	err = fs.enc.Encode(mail)
	if err != nil {
		return
	}

	fpath := filepath.Join(fs.dir, mail.ID)

	return ioutil.WriteFile(fpath, fs.buff.Bytes(), 0600)
}

func (fs *StorageFile) loadRaw(b []byte) (mail *MailTx, err error) {
	fs.buff.Reset()
	fs.buff.Write(b)

	mail = &MailTx{}
	err = fs.dec.Decode(mail)
	if err != nil {
		return nil, err
	}

	return mail, nil
}
