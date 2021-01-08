package memfs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
)

var (
	_testWD string
)

func TestAddFile(t *testing.T) {
	cases := []struct {
		desc     string
		path     string
		exp      *Node
		expError string
	}{{
		desc: "With empty path",
	}, {
		desc:     "With path is not exist",
		path:     "is/not/exist",
		expError: "memfs.AddFile: stat is: no such file or directory",
	}, {
		desc: "With file exist",
		path: "testdata/direct/add/file",
		exp: &Node{
			SysPath:     "testdata/direct/add/file",
			Path:        "testdata/direct/add/file",
			name:        "file",
			ContentType: "text/plain; charset=utf-8",
			size:        22,
			V:           []byte("Test direct add file.\n"),
			GenFuncName: "generate_testdata_direct_add_file",
		},
	}, {
		desc: "With directories exist",
		path: "testdata/direct/add/file2",
		exp: &Node{
			SysPath:     "testdata/direct/add/file2",
			Path:        "testdata/direct/add/file2",
			name:        "file2",
			ContentType: "text/plain; charset=utf-8",
			size:        24,
			V:           []byte("Test direct add file 2.\n"),
			GenFuncName: "generate_testdata_direct_add_file2",
		},
	}}

	opts := &Options{
		Root:        "testdata",
		WithContent: true,
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := mfs.AddFile(c.path)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error(), true)
			continue
		}

		if got != nil {
			got.modTime = time.Time{}
			got.mode = 0
			got.Parent = nil
			got.Childs = nil
		}

		test.Assert(t, "AddFile", c.exp, got, true)

		if c.exp != nil {
			got, err := mfs.Get(c.path)
			if err != nil {
				t.Fatal(err)
			}

			if got != nil {
				got.modTime = time.Time{}
				got.mode = 0
				got.Parent = nil
				got.Childs = nil
			}

			test.Assert(t, "Get", c.exp, got, true)
		}
	}
}

func TestGet(t *testing.T) {

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

	dir := filepath.Join(_testWD, "/testdata")

	opts := &Options{
		Root: dir,
		// Limit file size to allow testing Get from disk on file "index.js".
		MaxFileSize: 15,
		WithContent: true,
	}

	mfs, err := New(opts)
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

		if got.size <= opts.MaxFileSize {
			test.Assert(t, "node.V", c.expV, got.V, true)
		}

		test.Assert(t, "node.ContentType", c.expContentType,
			got.ContentType, true)
	}
}

func TestMemFS_mount(t *testing.T) {
	afile := filepath.Join(_testWD, "testdata/index.html")

	cases := []struct {
		desc       string
		opts       Options
		expErr     string
		expMapKeys []string
	}{{
		desc:       "With empty dir",
		expErr:     "open : no such file or directory",
		expMapKeys: make([]string, 0),
	}, {
		desc: "With file",
		opts: Options{
			Root:        afile,
			WithContent: true,
		},
		expErr: fmt.Sprintf("memfs.New: mount: %q must be a directory", afile),
	}, {
		desc: "With directory",
		opts: Options{
			Root: filepath.Join(_testWD, "testdata"),
			Excludes: []string{
				"memfs_generate.go$",
				"direct$",
			},
			WithContent: true,
		},
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
		opts: Options{
			Root: filepath.Join(_testWD, "testdata"),
			Excludes: []string{
				`.*\.js$`,
				"memfs_generate.go$",
				"direct$",
			},
			WithContent: true,
		},
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
		opts: Options{
			Root: filepath.Join(_testWD, "testdata"),
			Includes: []string{
				`.*\.js$`,
			},
			Excludes: []string{
				"memfs_generate.go$",
				"direct$",
			},
			WithContent: true,
		},
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

		mfs, err := New(&c.opts)
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

		opts := &Options{
			Includes:    c.inc,
			Excludes:    c.exc,
			WithContent: true,
		}
		mfs, err := New(opts)
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
