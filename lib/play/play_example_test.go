// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package play

import (
	"fmt"
	"log"
	"os"
	"regexp"
)

func ExampleGo_Format() {
	const codeIndentMissingImport = `
package main
func main() {
  fmt.Println("Hello, world")
}
`
	var playgo *Go
	var err error
	playgo, err = NewGo(GoOptions{
		Root: os.TempDir(),
	})
	if err != nil {
		log.Fatal(err)
	}

	var req = Request{
		Body: codeIndentMissingImport,
	}
	var out []byte
	out, err = playgo.Format(req)
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

func ExampleGo_Run() {
	const codeRun = `
package main
import "fmt"
func main() {
	fmt.Println("Hello, world")
}`

	var playgo *Go
	var err error
	playgo, err = NewGo(GoOptions{
		Root: os.TempDir(),
	})
	if err != nil {
		log.Fatal(err)
	}

	var req = Request{
		Body: codeRun,
	}
	var out []byte
	out, err = playgo.Run(&req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf(`%s`, out)

	//Output:
	//Hello, world
}

func ExampleGo_Test() {
	const codeTest = `
package test
import "testing"
func TestSum(t *testing.T) {
	var total = sum(1, 2, 3)
	if total != 6 {
		t.Fatalf("got %d, want 6", total)
	}
}`
	var rexDuration = regexp.MustCompile(`(?m)\s+(\d+\.\d+)s$`)

	var playgo *Go
	var err error
	playgo, err = NewGo(GoOptions{
		Root: `testdata/`,
	})
	if err != nil {
		log.Fatal(err)
	}

	var req = Request{
		Body: codeTest,
		File: `/test_test.go`,
	}
	var out []byte
	out, err = playgo.Test(&req)
	if err != nil {
		log.Fatal(err)
	}
	// Replace the test duration.
	out = rexDuration.ReplaceAll(out, []byte(" Xs"))
	fmt.Printf(`%s`, out)

	//Output:
	//ok  	git.sr.ht/~shulhan/pakakeh.go/lib/play/testdata Xs
}
