// Package internal contains helpers for testing http.
package internal

import (
	"fmt"
	"net/http"
	"strings"

	libhttp "github.com/shuLhan/share/lib/http"
	"github.com/shuLhan/share/lib/memfs"
)

// NewTestServer create new HTTP server for testing.
func NewTestServer() (srv *libhttp.Server, err error) {
	var (
		logp = `NewTestServer`
		opts = &libhttp.ServerOptions{
			Memfs: &memfs.MemFS{
				Opts: &memfs.Options{
					Root:        `./testdata`,
					MaxFileSize: 30,
					TryDirect:   true,
				},
			},
			HandleFS: handleFS,
			Address:  `127.0.0.1:14832`,
		}
	)

	srv, err = libhttp.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}

	return srv, nil
}

// handleFS authenticate the request to Memfs using cookie.
//
// If the node does not start with "/auth/" it will return true.
//
// If the node path is start with "/auth/" and cookie name "sid" exist
// with value "authz" it will return true;
// otherwise it will redirect to "/" and return false.
func handleFS(node *memfs.Node, res http.ResponseWriter, req *http.Request) bool {
	var (
		lowerPath = strings.ToLower(node.Path)

		cookieSid *http.Cookie
		err       error
	)
	if strings.HasPrefix(lowerPath, "/auth/") {
		cookieSid, err = req.Cookie("sid")
		if err != nil {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return false
		}
		if cookieSid.Value != "authz" {
			http.Redirect(res, req, "/", http.StatusSeeOther)
			return false
		}
	}
	return true
}
