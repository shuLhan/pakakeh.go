// Program httpd run HTTP server that serve files in directory testdata for
// testing with external tools.
// This program should be run from directory lib/http.
package main

import (
	"fmt"
	"log"

	libhttp "github.com/shuLhan/share/lib/http"
	"github.com/shuLhan/share/lib/http/internal"
)

func main() {
	var (
		srv *libhttp.Server
		err error
	)

	srv, err = internal.NewTestServer()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting test server at http://%s\n", srv.Options.Address)

	err = srv.Start()
	if err != nil {
		log.Println(err)
	}
}
