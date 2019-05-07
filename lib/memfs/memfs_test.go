package memfs

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	_testWD string //nolint: gochecknoglobals
)

func TestGet(t *testing.T) {
	// Limit file size to allow testing Get from disk on file "index.js".
	MaxFileSize = 15

	cases := []struct {
		path           string
		expV           []byte
		expContentType string
		expErr         error
	}{{
		path: "/",
	}, {
		path: "/exclude",
	}, {
		path:   "/exclude/dir",
		expErr: os.ErrNotExist,
	}, {
		path:           "/exclude/index.css",
		expV:           []byte("body {\n}\n"),
		expContentType: "text/css; charset=utf-8",
	}, {
		path:           "/exclude/index.html",
		expV:           []byte("<html></html>\n"),
		expContentType: "text/html; charset=utf-8",
	}, {
		path:           "/exclude/index.js",
		expContentType: "application/javascript",
	}, {
		path: "/include",
	}, {
		path:   "/include/dir",
		expErr: os.ErrNotExist,
	}, {
		path:           "/include/index.css",
		expV:           []byte("body {\n}\n"),
		expContentType: "text/css; charset=utf-8",
	}, {
		path:           "/include/index.html",
		expV:           []byte("<html></html>\n"),
		expContentType: "text/html; charset=utf-8",
	}, {
		path:           "/include/index.js",
		expContentType: "application/javascript",
	}, {
		path:           "/index.css",
		expV:           []byte("body {\n}\n"),
		expContentType: "text/css; charset=utf-8",
	}, {
		path:           "/index.html",
		expV:           []byte("<html></html>\n"),
		expContentType: "text/html; charset=utf-8",
	}, {
		path:           "/index.js",
		expContentType: "application/javascript",
	}, {
		path:           "/plain",
		expContentType: "application/octet-stream",
	}}

	mfs, err := New(nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	err = mfs.Mount(filepath.Join(_testWD, "/testdata"))
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Logf("Get %s", c.path)

		got, err := mfs.Get(c.path)
		if err != nil {
			test.Assert(t, "error", c.expErr, err, true)
			continue
		}

		if got.Size <= MaxFileSize {
			test.Assert(t, "node.V", c.expV, got.V, true)
		}

		test.Assert(t, "node.ContentType", c.expContentType,
			got.ContentType, true)
	}
}

func TestMount(t *testing.T) {
	cases := []struct {
		desc       string
		incs       []string
		excs       []string
		dir        string
		expErr     string
		expMapKeys []string
	}{{
		desc:       "With empty dir",
		expErr:     "open : no such file or directory",
		expMapKeys: make([]string, 0, 0),
	}, {
		desc:   "With file",
		dir:    filepath.Join(_testWD, "testdata/index.html"),
		expErr: "mount must be a directory",
	}, {
		desc: "With directory",
		excs: []string{
			"memfs_generate.go$",
		},
		dir: filepath.Join(_testWD, "testdata"),
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/index.css",
			"/exclude/index.html",
			"/exclude/index.js",
			"/include",
			"/include/index.css",
			"/include/index.html",
			"/include/index.js",
			"/index.css",
			"/index.html",
			"/index.js",
			"/plain",
		},
	}, {
		desc: "With excludes",
		excs: []string{
			`.*\.js$`,
			"memfs_generate.go$",
		},
		dir: filepath.Join(_testWD, "testdata"),
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/index.css",
			"/exclude/index.html",
			"/include",
			"/include/index.css",
			"/include/index.html",
			"/index.css",
			"/index.html",
			"/plain",
		},
	}, {
		desc: "With includes",
		incs: []string{
			`.*\.js$`,
		},
		excs: []string{
			"memfs_generate.go$",
		},
		dir: filepath.Join(_testWD, "testdata"),
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/index.js",
			"/include",
			"/include/index.js",
			"/index.js",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		mfs, err := New(c.incs, c.excs, true)
		if err != nil {
			t.Fatal(err)
		}

		err = mfs.Mount(c.dir)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error(), true)
			continue
		}

		gotListNames := mfs.ListNames()
		test.Assert(t, "names", c.expMapKeys, gotListNames, true)
	}
}

func TestFilter(t *testing.T) {
	cases := []struct {
		desc    string
		inc     []string
		exc     []string
		sysPath []string
		exp     []bool
	}{{
		desc: "With empty includes and excludes",
		sysPath: []string{
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/index.html"),
		},
		exp: []bool{
			true,
			true,
		},
	}, {
		desc: "With excludes only",
		exc: []string{
			`.*/exclude`,
			`.*\.html$`,
		},
		sysPath: []string{
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/exclude"),
			filepath.Join(_testWD, "/testdata/exclude/dir"),
			filepath.Join(_testWD, "/testdata/include"),
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/index.html"),
			filepath.Join(_testWD, "/testdata/index.css"),
		},
		exp: []bool{
			true,
			false,
			false,
			true,
			true,
			false,
			true,
		},
	}, {
		desc: "With includes only",
		inc: []string{
			".*/include",
			`.*\.html$`,
		},
		sysPath: []string{
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/include"),
			filepath.Join(_testWD, "/testdata/include/dir"),
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/index.html"),
			filepath.Join(_testWD, "/testdata/index.css"),
		},
		exp: []bool{
			true,
			true,
			true,
			true,
			true,
			false,
		},
	}, {
		desc: "With excludes and includes",
		exc: []string{
			`.*/exclude`,
			`.*\.js$`,
		},
		inc: []string{
			`.*/include`,
			`.*\.(css|html)$`,
		},
		sysPath: []string{
			filepath.Join(_testWD, "/testdata"),
			filepath.Join(_testWD, "/testdata/index.html"),
			filepath.Join(_testWD, "/testdata/index.css"),

			filepath.Join(_testWD, "/testdata/exclude"),
			filepath.Join(_testWD, "/testdata/exclude/dir"),
			filepath.Join(_testWD, "/testdata/exclude/index.css"),
			filepath.Join(_testWD, "/testdata/exclude/index.html"),
			filepath.Join(_testWD, "/testdata/exclude/index.js"),

			filepath.Join(_testWD, "/testdata/include"),
			filepath.Join(_testWD, "/testdata/include/dir"),
			filepath.Join(_testWD, "/testdata/include/index.css"),
			filepath.Join(_testWD, "/testdata/include/index.html"),
			filepath.Join(_testWD, "/testdata/include/index.js"),
		},
		exp: []bool{
			true,
			true,
			true,

			false,
			false,
			false,
			false,
			false,

			true,
			true,
			true,
			true,
			false,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		mfs, err := New(c.inc, c.exc, true)
		if err != nil {
			t.Fatal(err)
		}

		for x, sysPath := range c.sysPath {
			fi, err := os.Stat(sysPath)
			if err != nil {
				t.Fatal(err)
			}

			got := mfs.isIncluded(sysPath, fi.Mode())

			test.Assert(t, sysPath, c.exp[x], got, true)
		}
	}
}

func TestMain(m *testing.M) {
	var err error
	_testWD, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(filepath.Join(_testWD, "testdata/exclude/dir"), 0700)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok {
			log.Fatal("!ok:", err)
		}
		if perr.Err != os.ErrExist {
			log.Fatalf("perr: %+v %+v\n", perr.Err, os.ErrExist)
		}
	}

	err = os.MkdirAll(filepath.Join(_testWD, "testdata/include/dir"), 0700)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok {
			log.Fatal(err)
		}
		if perr.Err != os.ErrExist {
			log.Fatal(err)
		}
	}

	os.Exit(m.Run())
}
