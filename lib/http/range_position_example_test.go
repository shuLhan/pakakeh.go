package http_test

import (
	"fmt"

	libhttp "github.com/shuLhan/share/lib/http"
)

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
