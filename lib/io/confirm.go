// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package io

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

// ConfirmYesNo display a question to standard output and read for answer
// from input Reader for simple yes "y" or no "n" answer.
// If input Reader is nil, it will set to standard input.
// If "defIsYes" is true and answer is empty (only new line), then it will
// return true.
func ConfirmYesNo(in io.Reader, msg string, defIsYes bool) bool {
	var (
		logp = `ConfirmYesNo`

		r         *bufio.Reader
		b, answer byte
		err       error
	)

	if in == nil {
		r = bufio.NewReader(os.Stdin)
	} else {
		r = bufio.NewReader(in)
	}

	yon := "[y/N]"

	if defIsYes {
		yon = "[Y/n]"
	}

	fmt.Printf("%s %s ", msg, yon)

	for {
		b, err = r.ReadByte()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf(`%s: %s`, logp, err)
			}
			break
		}
		if b == ' ' || b == '\t' {
			continue
		}
		if b == '\n' {
			break
		}
		if answer == 0 {
			// Capture only the first non-space character.
			answer = b
		}
	}

	if answer == 'y' || answer == 'Y' {
		return true
	}
	if answer == 0 {
		return defIsYes
	}

	return false
}
