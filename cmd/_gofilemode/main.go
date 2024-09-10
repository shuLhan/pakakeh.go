// SPDX-FileCopyrightText: 2021 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Program gofilemode inspect the Go file mode [1].
// Usage,
//
//	gofilemode <mode>
//
// Example,
//
//	$ gofilemode 2147484159
//
// [1] http://127.0.0.1:6060/pkg/io/fs/#FileMode
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	var (
		cmdName = filepath.Base(os.Args[0])
	)

	flag.Parse()
	argMode := flag.Arg(0)

	if len(argMode) == 0 {
		usage(cmdName)
		os.Exit(2)
	}

	mode, err := strconv.ParseUint(argMode, 10, 32)
	if err != nil {
		log.Fatalf("%s: %s", cmdName, err)
	}

	fmode := fs.FileMode(mode)

	fmt.Printf(`
   Is directory: %t
 Is append-only: %t
   Is exclusive: %t
   Is temporary: %t
     Is symlink: %t
      Is device: %t
  Is named pipe: %t
      Is socket: %t
     Is set UID: %t
     Is set GID: %t
 Is char device: %t
      Is sticky: %t
   Is irregular: %t
     Permission: %o
      As string: %s
`,
		fmode.IsDir(),
		(fmode&fs.ModeAppend) == fs.ModeAppend,
		(fmode&fs.ModeExclusive) == fs.ModeExclusive,
		(fmode&fs.ModeTemporary) == fs.ModeTemporary,
		(fmode&fs.ModeSymlink) == fs.ModeSymlink,

		(fmode&fs.ModeDevice) == fs.ModeDevice,
		(fmode&fs.ModeNamedPipe) == fs.ModeNamedPipe,
		(fmode&fs.ModeSocket) == fs.ModeSocket,
		(fmode&fs.ModeSetuid) == fs.ModeSetuid,
		(fmode&fs.ModeSetgid) == fs.ModeSetgid,

		(fmode&fs.ModeCharDevice) == fs.ModeCharDevice,
		(fmode&fs.ModeSticky) == fs.ModeSticky,
		(fmode&fs.ModeIrregular) == fs.ModeIrregular,
		fmode.Perm(),
		fmode.String(),
	)
}

func usage(cmdName string) {
	fmt.Printf(`
= gofilemode

Decode the Go file mode bits from unsigned integer value as defined in
io/fs.FileMode.

== Usage

	%s <mode-bits>

== Example

Print the file mode information from unsigned integer,

	$ %s 2147484159
   Is directory: true
 Is append-only: false
   Is exclusive: false
   Is temporary: false
     Is symlink: false
      Is device: false
  Is named pipe: false
      Is socket: false
     Is set UID: false
     Is set GID: false
 Is char device: false
      Is sticky: false
   Is irregular: false
     Permission: 777
      As string: drwxrwxrwx
`, cmdName, cmdName)
}
