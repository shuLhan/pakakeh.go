// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package io extends the standard io library.
//
// DEPRECATED: This package has been merged into package lib/os and will
// be removed in the next six release v0.51.0.
package io

import (
	"fmt"
	"io"
	"log"
	"os"
)

// Copy file from in to out.
// If the output file is already exist, it will be truncated.
// If the file is not exist, it will created with permission set to user's
// read-write only.
//
// DEPRECATED: moved to lib/os#Copy.
func Copy(out, in string) (err error) {
	fin, err := os.Open(in)
	if err != nil {
		return fmt.Errorf("Copy: failed to open input file: %s", err)
	}

	fout, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Copy: failed to open output file: %s", err)
	}

	defer func() {
		err := fout.Close()
		if err != nil {
			log.Printf("Copy: failed to close output file: %s", err)
		}
	}()
	defer func() {
		err := fin.Close()
		if err != nil {
			log.Printf("Copy: failed to close input file: %s", err)
		}
	}()

	buf := make([]byte, 1024)
	for {
		n, err := fin.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if n == 0 {
			break
		}
		_, err = fout.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	return nil
}
