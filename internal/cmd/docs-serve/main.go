package main

import (
	"git.sr.ht/~shulhan/ciigo"
	"github.com/shuLhan/share/lib/debug"
)

func main() {
	debug.Value = 1
	ciigo.Serve(nil, "_doc", ":8080", "")
}
