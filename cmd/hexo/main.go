// Program hexcho print the input hex as hex itself, int64, octal, bytes,
// string, and binary.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

func main() {
	var (
		optFile bool
	)

	flag.BoolVar(&optFile, `file`, false, `make each arguments as files`)
	flag.Parse()

	var (
		arg string
		hex int64
		err error
		b   byte
	)

	if optFile {
		dumpFiles()
		return
	}

	for _, arg = range flag.Args() {
		fmt.Printf("[0x%s]\n", arg)

		hex, err = strconv.ParseInt(arg, 16, 64)
		if err != nil {
			log.Println(err)
			continue
		}

		var bytes = make([]byte, 0, 8)
		for x := 56; x >= 0; x -= 8 {
			bytes = append(bytes, byte(hex>>x))
		}

		fmt.Printf("  int64: %d\n", hex)

		fmt.Printf("    hex:")
		for _, b = range bytes {
			fmt.Printf(" %8x", b)
		}
		fmt.Println()

		fmt.Printf("  bytes:")
		for _, b = range bytes {
			fmt.Printf(" %8d", b)
		}
		fmt.Println()

		fmt.Printf("   char:")
		for _, b = range bytes {
			fmt.Printf(" %8c", b)
		}
		fmt.Println()
		fmt.Printf(" binary:")
		for _, b = range bytes {
			fmt.Printf(" %08b", b)
		}
		fmt.Println()
	}
}

func dumpFiles() {
	var (
		arg     string
		content []byte
		err     error
	)
	for _, arg = range flag.Args() {
		content, err = os.ReadFile(arg)
		if err != nil {
			log.Fatalf(`hexo: file %s: %s`, arg, err)
		}

		libbytes.DumpPrettyTable(os.Stdout, arg, content)
	}
}
