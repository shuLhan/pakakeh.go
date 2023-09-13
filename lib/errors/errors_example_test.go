package errors_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	liberrors "github.com/shuLhan/share/lib/errors"
)

func ExampleE_Is() {
	var (
		errFileNotFound = &liberrors.E{
			Code:    400,
			Name:    `ERR_NOT_FOUND`,
			Message: `file not found`,
		}
		errResNotFound = &liberrors.E{
			Code:    404,
			Name:    `ERR_NOT_FOUND`,
			Message: `resource not found`,
		}

		rawJson = `{"code":400,"name":"ERR_NOT_FOUND","message":"file not found"}`

		e   *liberrors.E
		err error
	)

	err = json.Unmarshal([]byte(rawJson), &e)
	if err != nil {
		log.Fatal(err)
	}

	var gotErr error = e

	fmt.Println(errors.Is(gotErr, errFileNotFound))
	fmt.Println(errors.Is(gotErr, errResNotFound))

	// Output:
	// true
	// false
}
