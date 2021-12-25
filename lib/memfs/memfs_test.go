package memfs

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shuLhan/share/lib/test"
	"github.com/shuLhan/share/lib/text/diff"
)

var (
	_testWD string
)

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

func TestNew(t *testing.T) {
	afile := filepath.Join(_testWD, "testdata/index.html")

	cases := []struct {
		desc       string
		expErr     string
		expMapKeys []string
		opts       Options
	}{{
		desc:   "With empty dir",
		expErr: "open : no such file or directory",
	}, {
		desc: "With file",
		opts: Options{
			Root: afile,
		},
		expErr: fmt.Sprintf("New: Init: mount: createRoot: %s must be a directory", afile),
	}, {
		desc: "With directory",
		opts: Options{
			Root: filepath.Join(_testWD, "testdata"),
			Excludes: []string{
				"memfs_generate.go$",
				"direct$",
				"node_save$",
			},
		},
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/index-link.css",
			"/exclude/index-link.html",
			"/exclude/index-link.js",
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
				"node_save$",
			},
		},
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/index-link.css",
			"/exclude/index-link.html",
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
				"node_save$",
			},
		},
		expMapKeys: []string{
			"/",
			"/exclude",
			"/exclude/index-link.js",
			"/include",
			"/include/index.js",
			"/index.js",
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		mfs, err := New(&c.opts)
		if err != nil {
			test.Assert(t, "error", c.expErr, err.Error())
			continue
		}

		gotListNames := mfs.ListNames()
		test.Assert(t, "ListNames", c.expMapKeys, gotListNames)
	}
}

func TestMemFS_AddFile(t *testing.T) {
	cases := []struct {
		desc     string
		intPath  string
		extPath  string
		exp      *Node
		expError string
	}{{
		desc: "With empty internal path",
	}, {
		desc:     "With external path is not exist",
		intPath:  "internal/file",
		extPath:  "is/not/exist",
		expError: "AddFile: stat is/not/exist: no such file or directory",
	}, {
		desc:    "With file exist",
		intPath: "internal/file",
		extPath: "testdata/direct/add/file",
		exp: &Node{
			SysPath:     "testdata/direct/add/file",
			Path:        "internal/file",
			name:        "file",
			ContentType: "text/plain; charset=utf-8",
			size:        22,
			Content:     []byte("Test direct add file.\n"),
			GenFuncName: "generate_internal_file",
		},
	}, {
		desc:    "With directories exist",
		intPath: "internal/file2",
		extPath: "testdata/direct/add/file2",
		exp: &Node{
			SysPath:     "testdata/direct/add/file2",
			Path:        "internal/file2",
			name:        "file2",
			ContentType: "text/plain; charset=utf-8",
			size:        24,
			Content:     []byte("Test direct add file 2.\n"),
			GenFuncName: "generate_internal_file2",
		},
	}}

	opts := &Options{
		Root: "testdata",
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := mfs.AddFile(c.intPath, c.extPath)
		if err != nil {
			test.Assert(t, "error", c.expError, err.Error())
			continue
		}

		if got != nil {
			got.modTime = time.Time{}
			got.mode = 0
			got.Parent = nil
			got.Childs = nil
		}

		test.Assert(t, c.desc+": AddFile", c.exp, got)

		if c.exp == nil {
			continue
		}

		got, err = mfs.Get(c.intPath)
		if err != nil {
			t.Fatal(err)
		}

		if got != nil {
			got.modTime = time.Time{}
			got.mode = 0
			got.Parent = nil
			got.Childs = nil
		}

		test.Assert(t, c.desc+": Get", c.exp, got)
	}
}

func TestMemFS_Get(t *testing.T) {
	cases := []struct {
		expErr         error
		path           string
		expV           []byte
		expContentType []string
	}{{
		path: "/",
	}, {
		path: "/exclude",
	}, {
		path:   "/exclude/dir",
		expErr: os.ErrNotExist,
	}, {
		path:           "/exclude/index-link.css",
		expV:           []byte("body {\n}\n"),
		expContentType: []string{"text/css; charset=utf-8"},
	}, {
		path:           "/exclude/index-link.html",
		expV:           []byte("<html></html>\n"),
		expContentType: []string{"text/html; charset=utf-8"},
	}, {
		path: "/exclude/index-link.js",
		expV: []byte("function X() {}\n"),
		expContentType: []string{
			"text/javascript; charset=utf-8",
			"application/javascript",
		},
	}, {
		path: "/include",
	}, {
		path:   "/include/dir",
		expErr: os.ErrNotExist,
	}, {
		path:           "/include/index.css",
		expV:           []byte("body {\n}\n"),
		expContentType: []string{"text/css; charset=utf-8"},
	}, {
		path:           "/include/index.html",
		expV:           []byte("<html></html>\n"),
		expContentType: []string{"text/html; charset=utf-8"},
	}, {
		path: "/include/index.js",
		expV: []byte("function X() {}\n"),
		expContentType: []string{
			"text/javascript; charset=utf-8",
			"application/javascript",
		},
	}, {
		path:           "/index.css",
		expV:           []byte("body {\n}\n"),
		expContentType: []string{"text/css; charset=utf-8"},
	}, {
		path:           "/index.html",
		expV:           []byte("<html></html>\n"),
		expContentType: []string{"text/html; charset=utf-8"},
	}, {
		path: "/index.js",
		expContentType: []string{
			"text/javascript; charset=utf-8",
			"application/javascript",
		},
	}, {
		path:           "/plain",
		expContentType: []string{"application/octet-stream"},
	}}

	dir := filepath.Join(_testWD, "/testdata")

	opts := &Options{
		Root: dir,
		// Limit file size to allow testing Get from disk on file "index.js".
		MaxFileSize: 15,
	}

	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range cases {
		got, err := mfs.Get(c.path)
		if err != nil {
			test.Assert(t, c.path+": error", c.expErr, err)
			continue
		}

		if got.size <= opts.MaxFileSize {
			test.Assert(t, c.path+": node.Content", c.expV, got.Content)
		}

		if len(got.ContentType) == 0 && len(c.expContentType) == 0 {
			continue
		}

		found := false
		for _, expCT := range c.expContentType {
			if expCT == got.ContentType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expecting one of the Content-Type %v, got %s",
				c.expContentType, got.ContentType)
		}
	}
}

func TestMemFS_MarshalJSON(t *testing.T) {
	logp := "MarshalJSON"
	modTime := time.Date(2021, 7, 30, 20, 04, 00, 0, time.UTC)

	opts := &Options{
		Root: "testdata/direct/",
	}
	mfs, err := New(opts)
	if err != nil {
		t.Fatal(err)
	}

	mfs.resetAllModTime(modTime)

	got, err := json.MarshalIndent(mfs, "", "\t")
	if err != nil {
		t.Fatal(err)
	}

	exp := `{
	"path": "/",
	"name": "/",
	"content_type": "",
	"mod_time": 1627675440,
	"mode_string": "drwxr-xr-x",
	"size": 0,
	"is_dir": true,
	"childs": [
		{
			"path": "/add",
			"name": "add",
			"content_type": "",
			"mod_time": 1627675440,
			"mode_string": "drwxr-xr-x",
			"size": 0,
			"is_dir": true,
			"childs": null
		}
	]
}`

	diffs := diff.Text([]byte(exp), got, diff.LevelLines)
	if len(diffs.Adds) != 0 {
		t.Fatalf("%s: adds: %v", logp, diffs.Adds)
	}
	if len(diffs.Dels) != 0 {
		t.Fatalf("%s: dels: %#v", logp, diffs.Dels)
	}
	if len(diffs.Changes) != 0 {
		t.Fatalf("%s: changes: %s", logp, diffs.Changes)
	}
}

func TestMemFS_isIncluded(t *testing.T) {
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
			filepath.Join(_testWD, "/testdata/exclude/index-link.css"),
			filepath.Join(_testWD, "/testdata/exclude/index-link.html"),
			filepath.Join(_testWD, "/testdata/exclude/index-link.js"),

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
			Includes: c.inc,
			Excludes: c.exc,
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

			test.Assert(t, sysPath, c.exp[x], got)
		}
	}
}

func TestMerge(t *testing.T) {
	optsDirect := &Options{
		Root: "testdata/direct",
	}
	mfsDirect, err := New(optsDirect)
	if err != nil {
		t.Fatal(err)
	}

	optsInclude := &Options{
		Root: "testdata/include",
	}
	mfsInclude, err := New(optsInclude)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		exp    *MemFS
		desc   string
		params []*MemFS
	}{{
		desc:   "with the same instance",
		params: []*MemFS{mfsDirect, mfsDirect},
		exp: &MemFS{
			PathNodes: &PathNode{
				v: map[string]*Node{
					"/": &Node{
						SysPath: "..",
						Path:    "/",
						Childs: []*Node{
							mfsDirect.MustGet("/add"),
						},
						mode: 2147484141,
					},
					"/add":       mfsDirect.MustGet("/add"),
					"/add/file":  mfsDirect.MustGet("/add/file"),
					"/add/file2": mfsDirect.MustGet("/add/file2"),
				},
			},
		},
	}, {
		desc:   "with different instances",
		params: []*MemFS{mfsDirect, mfsInclude},
		exp: &MemFS{
			PathNodes: &PathNode{
				v: map[string]*Node{
					"/": &Node{
						SysPath: "..",
						Path:    "/",
						Childs: []*Node{
							mfsDirect.MustGet("/add"),
							mfsInclude.MustGet("/index.css"),
							mfsInclude.MustGet("/index.html"),
							mfsInclude.MustGet("/index.js"),
						},
						mode: 2147484141,
					},
					"/add":        mfsDirect.MustGet("/add"),
					"/add/file":   mfsDirect.MustGet("/add/file"),
					"/add/file2":  mfsDirect.MustGet("/add/file2"),
					"/index.css":  mfsInclude.MustGet("/index.css"),
					"/index.html": mfsInclude.MustGet("/index.html"),
					"/index.js":   mfsInclude.MustGet("/index.js"),
				},
			},
		},
	}}

	for _, c := range cases {
		got := Merge(c.params...)

		test.Assert(t, c.desc, c.exp.PathNodes.v, got.PathNodes.v)
	}
}

func TestScanDir(t *testing.T) {
	opts := Options{
		Root: "testdata/",
	}
	_, err := New(&opts)
	if err != nil {
		t.Fatal(err)
	}
}
