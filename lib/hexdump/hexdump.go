// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

// Package hexdump implements reading and writing bytes from and into
// hexadecimal number.
// It support parsing output from hexdump(1) tool.
package hexdump

import (
	"fmt"
	"io"
	"strconv"

	"git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

// Parse parse the default output of [hexdump(1)] utility from parameter in
// back into stream of byte.
//
// An example of default output of hexdump is
//
//	0000000 7865 5f70 6964 2f72 0000 0000 0000 0000
//	0000010 0000 0000 0000 0000 0000 0000 0000 0000
//	*
//	0000060 0000 0000 3030 3030 3537 0035 3030 3130
//
// The first column is the address and the rest of the column is the data.
//
// Each data column is 16-bit words in little-endian order, so in the above
// example, the first byte would be 65, second byte is 78 and so on.
// If parameter networkByteOrder (big-endian) is true, the first byte would be
// 78, second by is 65, and so on.
//
// The asterisk "*" means that the values from address 0000020 to 0000050 is
// equal to the previous line, 0000010.
//
// [hexdump(1)]: https://man.archlinux.org/man/hexdump.1
func Parse(in []byte, networkByteOrder bool) (out []byte, err error) {
	var (
		logp        = `ParseHexDump`
		parser      = bytes.NewParser(in, []byte(" \n"))
		d      byte = 255 // Just to make the first for-loop pass.

		token      []byte
		vint64     int64
		isAsterisk bool
	)
	for d != 0 {
		// Read the address.
		token, d = parser.Read()
		if len(token) == 0 {
			break
		}
		if len(token) == 1 {
			if token[0] != '*' {
				break
			}
			isAsterisk = true
			continue
		}

		vint64, err = strconv.ParseInt(string(token), 16, 64)
		if err != nil {
			return nil, fmt.Errorf(`%s: %w`, logp, err)
		}

		if isAsterisk {
			if len(out) > 0 {
				var start = len(out)
				if start < 16 {
					start = 0
				} else {
					start -= 16
				}
				var (
					prevRow      = out[start:]
					identicalRow = int((vint64 - int64(len(out))) / 16)
				)
				for range identicalRow {
					out = append(out, prevRow...)
				}
			}
		}

		// Read the two-hex, 16-bit words.
		for range 8 {
			token, d = parser.Read()
			if len(token) == 0 {
				break
			}

			vint64, err = strconv.ParseInt(string(token), 16, 64)
			if err != nil {
				return nil, fmt.Errorf(`%s: %w`, logp, err)
			}

			switch len(token) {
			case 2:
				out = append(out, byte(vint64))
			case 4:
				if networkByteOrder {
					out = append(out, byte(vint64>>8))
					out = append(out, byte(vint64))
				} else {
					out = append(out, byte(vint64))
					out = append(out, byte(vint64>>8))
				}
			}

			if d == '\n' {
				break
			}
		}
		// Ignore trailing characters.
		if d != '\n' {
			parser.SkipLine()
		}
	}
	return out, nil
}

// Print print each byte in slice as hexadecimal value into N column length.
func Print(title string, data []byte, col int) {
	var (
		start, x int
		c        byte
	)
	fmt.Print(title)
	for x, c = range data {
		if x%col == 0 {
			if x > 0 {
				fmt.Print(` ||`)
			}
			for y := start; y < x; y++ {
				if data[y] >= 33 && data[y] <= 126 {
					fmt.Printf(` %c`, data[y])
				} else {
					fmt.Print(` .`)
				}
			}
			fmt.Printf("\n%4d -", x)
			start = x
		}

		fmt.Printf(` %02X`, c)
	}
	rest := col - (x % col)
	if rest > 0 {
		for y := 1; y < rest; y++ {
			fmt.Print(`   `)
		}
		fmt.Print(` ||`)
	}
	for y := start; y <= x; y++ {
		if data[y] >= 33 && data[y] <= 126 {
			fmt.Printf(` %c`, data[y])
		} else {
			fmt.Print(` .`)
		}
	}

	fmt.Println()
}

// PrettyPrint write each byte in slice data as hexadecimal, ASCII
// character, and integer with 8 columns width.
func PrettyPrint(w io.Writer, title string, data []byte) {
	const ncol = 8

	fmt.Fprintf(w, "%s\n", title)
	fmt.Fprint(w, "          |  0  1  2  3  4  5  6  7 | 01234567 |   0   1   2   3   4   5   6   7 |\n")
	fmt.Fprint(w, "          |  8  9  A  B  C  D  E  F | 89ABCDEF |   8   9   A   B   C   D   E   F |\n")

	var (
		chunks = bytes.SplitEach(data, ncol)
		chunk  []byte
		x      int
		y      int
		c      byte
	)
	for x, chunk = range chunks {
		fmt.Fprintf(w, `%#08x|`, x*ncol)

		// Print as hex.
		for y, c = range chunk {
			fmt.Fprintf(w, ` %02x`, c)
		}
		for y++; y < ncol; y++ {
			fmt.Fprint(w, `   `)
		}

		// Print as char.
		fmt.Fprint(w, ` | `)
		for y, c = range chunk {
			if c >= 33 && c <= 126 {
				fmt.Fprintf(w, `%c`, c)
			} else {
				fmt.Fprint(w, `.`)
			}
		}
		for y++; y < ncol; y++ {
			fmt.Fprint(w, ` `)
		}

		// Print as integer.
		fmt.Fprint(w, ` |`)
		for y, c = range chunk {
			fmt.Fprintf(w, ` %3d`, c)
		}
		for y++; y < ncol; y++ {
			fmt.Fprint(w, `    `)
		}
		fmt.Fprintf(w, " |%d\n", x*ncol)
	}
}
