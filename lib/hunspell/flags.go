// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"fmt"
	"strconv"
	"strings"
)

func unpackFlags(ftype, rawflags string) (flags []string, err error) {
	switch ftype {
	case FlagASCII:
		flags = unpackFlagASCII(rawflags)
	case FlagUTF8:
		flags = unpackFlagUTF8(rawflags)
	case FlagLong:
		flags, err = unpackFlagLong(rawflags)
		if err != nil {
			return nil, err
		}
	case FlagNum:
		flags, err = unpackFlagNum(rawflags)
		if err != nil {
			return nil, err
		}
	}

	return flags, nil
}

func unpackFlagASCII(f string) (flags []string) {
	for x := 0; x < len(f); x++ {
		flags = append(flags, string(f[x]))
	}
	return
}

func unpackFlagUTF8(f string) (flags []string) {
	for _, r := range f {
		flags = append(flags, string(r))
	}
	return
}

func unpackFlagLong(f string) (flags []string, err error) {
	if len(f)%2 != 0 {
		return nil, fmt.Errorf("invalid long flags: %q", f)
	}
	for x := 0; x < len(f); x += 2 {
		flags = append(flags, f[x:x+2])
	}
	return
}

func unpackFlagNum(f string) (flags []string, err error) {
	flags = strings.Split(f, ",")

	// Trim spaces and check if all the flags is valid number.
	for x := 0; x < len(flags); x++ {
		flags[x] = strings.TrimSpace(flags[x])

		_, err = strconv.Atoi(flags[x])
		if err != nil {
			return nil, fmt.Errorf("invalid num flags: %q", flags[x])
		}
	}

	return
}
