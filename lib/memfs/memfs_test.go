package memfs

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

var (
	_testWD string
)

func TestGet(t *testing.T) {
	// Limit file size to allow testing Get from disk on file "index.js".
	MaxFileSize = 15

	cases := []struct {
		desc    string
		dir     string
		paths   []string
		exp     [][]byte
		expErrs []error
	}{{
		desc: "With '/'",
		dir:  filepath.Join(_testWD, "/testdata"),
		paths: []string{
			"/",

			"/exclude",
			"/exclude/dir",
			"/exclude/index.css",
			"/exclude/index.html",
			"/exclude/index.js",

			"/include",
			"/include/dir",
			"/include/index.css",
			"/include/index.html",
			"/include/index.js",

			"/index.css",
			"/index.html",
			"/index.js",
		},
		exp: [][]byte{
			nil,

			nil,
			nil,
			[]byte("body {\n}\n"),
			[]byte("<html></html>\n"),
			[]byte("function X() {}\n"),

			nil,
			nil,
			[]byte("body {\n}\n"),
			[]byte("<html></html>\n"),
			[]byte("function X() {}\n"),

			[]byte("body {\n}\n"),
			[]byte("<html></html>\n"),
			[]byte("function X() {}\n"),
		},
		expErrs: []error{
			nil,

			nil,
			os.ErrNotExist,
			nil,
			nil,
			nil,

			nil,
			os.ErrNotExist,
			nil,
			nil,
			nil,

			nil,
			nil,
			nil,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		mfs, err := New(nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		err = mfs.Mount(c.dir)
		if err != nil {
			t.Fatal(err)
		}

		for x, path := range c.paths {
			t.Log("Get:", path)

			got, err := mfs.Get(path)
			if err != nil {
				test.Assert(t, "error", c.expErrs[x], err, true)
				continue
			}

			test.Assert(t, "content", c.exp[x], got, true)
		}
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
		desc:   "With empty dir",
		expErr: "open : no such file or directory",
	}, {
		desc:   "With file",
		dir:    filepath.Join(_testWD, "testdata/index.html"),
		expErr: "Mount must be a directory.",
	}, {
		desc: "With directory",
		dir:  filepath.Join(_testWD, "testdata"),
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
		},
	}, {
		desc: "With excludes",
		excs: []string{
			`.*\.js$`,
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
		},
	}, {
		desc: "With includes",
		incs: []string{
			`.*\.js$`,
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

		mfs, err := New(c.incs, c.excs)
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

func TestFiter(t *testing.T) {
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
			`.*\.js`,
		},
		inc: []string{
			`.*/include`,
			`.*\.html`,
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
			false,

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

		mfs, err := New(c.inc, c.exc)
		if err != nil {
			t.Fatal(err)
		}

		for x, sysPath := range c.sysPath {
			node, err := newNode(sysPath)
			if err != nil {
				t.Fatal(err)
			}

			got := mfs.isIncluded(node)

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
