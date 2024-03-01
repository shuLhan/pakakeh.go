package http_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
)

func ExampleParseMultipartRange() {
	var (
		boundary = `zxcv`
	)

	var body = `--zxcv
Content-Range: bytes 0-6/50

Part 1
--zxcv

Missing Content-Range header, skipped.
--zxcv
Content-Range: bytes 7-13

Invalid Content-Range, missing size, skipped.
--zxcv
Content-Range: bytes 14-19/50

Part 2
--zxcv--
`

	body = strings.ReplaceAll(body, "\n", "\r\n")

	var (
		reader = bytes.NewReader([]byte(body))

		r   *libhttp.Range
		err error
	)
	r, err = libhttp.ParseMultipartRange(reader, boundary)
	if err != nil {
		log.Fatal(err)
	}

	var pos *libhttp.RangePosition
	for _, pos = range r.Positions() {
		fmt.Printf("%s: %s\n", pos.String(), pos.Content())
	}
	// Output:
	// 0-6: Part 1
	// 14-19: Part 2
}

func ExampleParseRange() {
	var r libhttp.Range

	// Empty range due to missing "=".
	r = libhttp.ParseRange(`bytes`)
	fmt.Println(r.String())

	r = libhttp.ParseRange(`bytes=10-`)
	fmt.Println(r.String())

	// The "20-30" is overlap with "10-".
	r = libhttp.ParseRange(`bytes=10-,20-30`)
	fmt.Println(r.String())

	// The "10-" is ignored since its overlap with the first range
	// "20-30".
	r = libhttp.ParseRange(`bytes=20 - 30 , 10 -`)
	fmt.Println(r.String())

	r = libhttp.ParseRange(`bytes=10-20`)
	fmt.Println(r.String())

	r = libhttp.ParseRange(`bytes=-20`)
	fmt.Println(r.String())

	r = libhttp.ParseRange(`bytes=0-9,10-19,-20`)
	fmt.Println(r.String())

	r = libhttp.ParseRange(`bytes=0-`)
	fmt.Println(r.String())

	// The only valid position here is 0-9, 10-19, and -20.
	// The x, -x, x-9, 0-x, 0-9-, and -0-9 is not valid position.
	// The -10 is overlap with -20.
	r = libhttp.ParseRange(`bytes=,x,-x,x-9,0-x,0-9-,-0-9,0-9,10-19,-20,-10,`)
	fmt.Println(r.String())

	// Output:
	//
	// bytes=10-
	// bytes=10-
	// bytes=20-30
	// bytes=10-20
	// bytes=-20
	// bytes=0-9,10-19,-20
	// bytes=0-
	// bytes=0-9,10-19,-20
}

func ptrInt64(v int64) *int64 { return &v }

func ExampleRange_Add() {
	var listpos = []struct {
		start *int64
		end   *int64
	}{
		{ptrInt64(0), ptrInt64(9)},  // OK.
		{ptrInt64(0), ptrInt64(5)},  // Overlap with [0,9].
		{ptrInt64(9), ptrInt64(19)}, // Overlap with [0,9].

		{ptrInt64(10), ptrInt64(19)}, // OK.
		{ptrInt64(19), ptrInt64(20)}, // Overlap with [10,19].
		{ptrInt64(20), ptrInt64(19)}, // End less than start.

		{nil, ptrInt64(10)}, // OK.
		{nil, ptrInt64(20)}, // Overlap with [nil,10].

		{ptrInt64(20), nil},          // Overlap with [nil,10].
		{ptrInt64(30), ptrInt64(40)}, // Overlap with [20,nil].
		{ptrInt64(30), nil},          // Overlap with [20,nil].
	}

	var r = libhttp.NewRange(``)

	for _, pos := range listpos {
		fmt.Println(r.Add(pos.start, pos.end), r.String())
	}

	// Output:
	// true bytes=0-9
	// false bytes=0-9
	// false bytes=0-9
	// true bytes=0-9,10-19
	// false bytes=0-9,10-19
	// false bytes=0-9,10-19
	// true bytes=0-9,10-19,-10
	// false bytes=0-9,10-19,-10
	// false bytes=0-9,10-19,-10
	// true bytes=0-9,10-19,-10,30-40
	// false bytes=0-9,10-19,-10,30-40
}

func ExampleRange_Positions() {
	var r = libhttp.NewRange(``)
	fmt.Println(r.Positions()) // Empty positions.

	r.Add(ptrInt64(10), ptrInt64(20))
	fmt.Println(r.Positions())
	// Output:
	// []
	// [10-20]
}

func ExampleRange_String() {
	var r = libhttp.NewRange(`MyUnit`)

	fmt.Println(r.String()) // Empty range will return empty string.

	r.Add(ptrInt64(0), ptrInt64(9))
	fmt.Println(r.String())
	// Output:
	//
	// myunit=0-9
}
