package http_test

import (
	"fmt"

	libhttp "github.com/shuLhan/share/lib/http"
)

func ExampleParseContentRange() {
	fmt.Println(libhttp.ParseContentRange(`bytes 10-/20`))   // OK
	fmt.Println(libhttp.ParseContentRange(`bytes 10-19/20`)) // OK
	fmt.Println(libhttp.ParseContentRange(`bytes -10/20`))   // OK
	fmt.Println(libhttp.ParseContentRange(`10-20/20`))       // Invalid, missing unit.
	fmt.Println(libhttp.ParseContentRange(`bytes 10-`))      // Invalid, missing "/size".
	fmt.Println(libhttp.ParseContentRange(`bytes -10/x`))    // Invalid, invalid "size".
	fmt.Println(libhttp.ParseContentRange(`bytes`))          // Invalid, missing position.
	// Output:
	// 10-
	// 10-19
	// -10
	// <nil>
	// <nil>
	// <nil>
	// <nil>
}

func ExampleRangePosition_ContentRange() {
	var (
		unit = libhttp.AcceptRangesBytes
		pos  = libhttp.RangePosition{
			Start: 10,
			End:   20,
		}
	)

	fmt.Println(pos.ContentRange(unit, 512))
	fmt.Println(pos.ContentRange(unit, 0))
	// Output:
	// bytes 10-20/512
	// bytes 10-20/*
}
