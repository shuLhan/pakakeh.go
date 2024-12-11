// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"fmt"
	"log"
	"regexp"
)

func ExampleFormat() {
	const codeIndentMissingImport = `
package main
func main() {
  fmt.Println("Hello, world")
}
`
	var req = Request{
		Body: codeIndentMissingImport,
	}
	var (
		out []byte
		err error
	)
	out, err = Format(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", out)

	//Output:
	//package main
	//
	//import "fmt"
	//
	//func main() {
	//	fmt.Println("Hello, world")
	//}
}

func ExampleRun() {
	const codeRun = `
package main
import "fmt"
func main() {
	fmt.Println("Hello, world")
}`

	var req = Request{
		Body: codeRun,
	}
	var (
		out []byte
		err error
	)
	out, err = Run(&req)
	if err != nil {
		fmt.Printf(`error: %s`, err)
	}
	fmt.Printf(`%s`, out)

	//Output:
	//Hello, world
}

func ExampleTest() {
	const codeTest = `
package test
import "testing"
func TestSum(t *testing.T) {
	var total = sum(1, 2, 3)
	if total != 6 {
		t.Fatalf("got %d, want 6", total)
	}
}`
	var req = Request{
		Body: codeTest,
		File: `testdata/test_test.go`,
	}
	var (
		rexDuration = regexp.MustCompile(`(?m)\s+(\d+\.\d+)s$`)
		out         []byte
		err         error
	)
	out, err = Test(&req)
	if err != nil {
		fmt.Printf(`error: %s`, err)
	}
	// Replace the test duration.
	out = rexDuration.ReplaceAll(out, []byte(" Xs"))
	fmt.Printf(`%s`, out)

	//Output:
	//ok  	git.sr.ht/~shulhan/pakakeh.go/lib/play/testdata Xs
}
