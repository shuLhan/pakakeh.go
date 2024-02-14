package http_test

import (
	"crypto/rand"
	"fmt"
	"log"
	"strings"

	libhttp "github.com/shuLhan/share/lib/http"
	"github.com/shuLhan/share/lib/test/mock"
)

func ExampleGenerateFormData() {
	// Mock the random reader for predictable output.
	// NOTE: do not do this on real code.
	rand.Reader = mock.NewRandReader([]byte(`randomseed`))

	var data = map[string][]byte{
		`name`: []byte(`test.txt`),
		`size`: []byte(`42`),
	}

	var (
		contentType string
		body        string
		err         error
	)
	contentType, body, err = libhttp.GenerateFormData(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(`contentType:`, contentType)
	fmt.Println(`body:`)
	fmt.Println(strings.ReplaceAll(body, "\r\n", "\n"))
	// Output:
	// contentType: multipart/form-data; boundary=72616e646f6d7365656472616e646f6d7365656472616e646f6d73656564
	// body:
	// --72616e646f6d7365656472616e646f6d7365656472616e646f6d73656564
	// Content-Disposition: form-data; name="name"
	//
	// test.txt
	// --72616e646f6d7365656472616e646f6d7365656472616e646f6d73656564
	// Content-Disposition: form-data; name="size"
	//
	// 42
	// --72616e646f6d7365656472616e646f6d7365656472616e646f6d73656564--
}
