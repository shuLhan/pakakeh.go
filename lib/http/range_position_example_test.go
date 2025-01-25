package http_test

import (
	"fmt"

	libhttp "git.sr.ht/~shulhan/pakakeh.go/lib/http"
)

func ExampleParseContentRange() {
	fmt.Println(libhttp.ParseContentRange(`bytes 10-/20`))   // Invalid, missing end.
	fmt.Println(libhttp.ParseContentRange(`bytes 10-19/20`)) // OK
	fmt.Println(libhttp.ParseContentRange(`bytes -10/20`))   // Invalid, missing start.
	fmt.Println(libhttp.ParseContentRange(`10-20/20`))       // Invalid, missing unit.
	fmt.Println(libhttp.ParseContentRange(`bytes 10-`))      // Invalid, missing "/size".
	fmt.Println(libhttp.ParseContentRange(`bytes -10/x`))    // Invalid, invalid "size".
	fmt.Println(libhttp.ParseContentRange(`bytes`))          // Invalid, missing position.
	// Output:
	// <nil> ParseContentRange: invalid Content-Range "bytes 10-/20": strconv.ParseInt: parsing "": invalid syntax
	// 10-19 <nil>
	// <nil> ParseContentRange: invalid Content-Range "bytes -10/20"
	// <nil> ParseContentRange: invalid Content-Range "10-20/20"
	// <nil> ParseContentRange: invalid Content-Range "bytes 10-"
	// <nil> ParseContentRange: invalid Content-Range "bytes -10/x"
	// <nil> ParseContentRange: invalid Content-Range "bytes"
}
