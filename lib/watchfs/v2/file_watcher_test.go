// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.sr.ht/~shulhan/pakakeh.go/lib/reflect"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
	"git.sr.ht/~shulhan/pakakeh.go/lib/watchfs/v2"
)

type directFileInfo struct {
	modTime time.Time
	reflect.Equaler
	name  string
	size  int64
	mode  os.FileMode
	isDir bool
}

func newDirectFileInfo(fi os.FileInfo) (directfi *directFileInfo) {
	if fi == nil {
		return nil
	}
	directfi = &directFileInfo{
		name:    fi.Name(),
		size:    fi.Size(),
		mode:    fi.Mode(),
		modTime: fi.ModTime().Truncate(time.Second),
		isDir:   fi.IsDir(),
	}
	return directfi
}

func (directfi directFileInfo) Equal(v any) (err error) {
	var (
		other *directFileInfo
		ok    bool
	)
	other, ok = v.(*directFileInfo)
	if !ok {
		return fmt.Errorf(`expecting type %T, got %T`, other, v)
	}

	if directfi.name != other.name {
		return fmt.Errorf(`name: got %s, want %s`,
			other.name, directfi.name)
	}
	if directfi.size != other.size {
		return fmt.Errorf(`size: got %d, want %d`,
			other.size, directfi.size)
	}
	if directfi.mode != other.mode {
		return fmt.Errorf(`filemode: got %d, want %d`,
			other.mode, directfi.mode)
	}
	if directfi.modTime.IsZero() {
		directfi.modTime = other.modTime
	} else if directfi.modTime.After(other.modTime) {
		return fmt.Errorf(`modTime: got %v, want %v`,
			other.modTime, directfi.modTime)
	}
	if directfi.isDir != other.isDir {
		return fmt.Errorf(`isDir: got %t, want %t`,
			other.isDir, directfi.isDir)
	}
	return nil
}

func TestWatchFile(t *testing.T) {
	var (
		name = `file.txt`
		opts = watchfs.FileWatcherOptions{
			File:     filepath.Join(t.TempDir(), name),
			Interval: 50 * time.Millisecond,
		}

		fwatch = watchfs.WatchFile(opts)
		expfi  = &directFileInfo{
			name: name,
			size: 0,
			mode: 420,
		}

		file *os.File
		err  error
	)

	t.Run(`On created`, func(t *testing.T) {
		file, err = os.Create(opts.File)
		if err != nil {
			t.Fatal(err)
		}

		// It should trigger an event.
		var fi os.FileInfo = <-fwatch.C
		var gotfi = newDirectFileInfo(fi)
		test.Assert(t, `fwatch.C`, expfi, gotfi)

		fi, err = file.Stat()
		if err != nil {
			t.Fatal(err)
		}
		var orgfi = newDirectFileInfo(fi)
		test.Assert(t, `created `+opts.File+` on FileInfo`,
			orgfi, gotfi)

		expfi.modTime = orgfi.modTime
	})

	t.Run(`On update content`, func(t *testing.T) {
		var expBody = `update`
		_, err = file.WriteString(expBody)
		if err != nil {
			t.Fatal(err)
		}
		err = file.Sync()
		if err != nil {
			t.Fatal(err)
		}

		var fi os.FileInfo = <-fwatch.C
		var gotfi = newDirectFileInfo(fi)
		expfi.size = 6
		test.Assert(t, `fwatch.C`, expfi, gotfi)

		var gotBody []byte
		gotBody, err = os.ReadFile(opts.File)
		if err != nil {
			t.Fatal(err)
		}
		test.Assert(t, `body`, expBody, string(gotBody))

		expfi.modTime = gotfi.modTime
	})

	t.Run(`On update mode`, func(t *testing.T) {
		var expMode = os.FileMode(0750)
		err = file.Chmod(expMode)
		if err != nil {
			t.Fatal(err)
		}

		var fi os.FileInfo = <-fwatch.C
		var gotfi = newDirectFileInfo(fi)
		expfi.mode = expMode
		test.Assert(t, `fwatch.C`, expfi, gotfi)

		expfi.modTime = gotfi.modTime
	})

	t.Run(`On deleted`, func(t *testing.T) {
		err = file.Close()
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove(opts.File)
		if err != nil {
			t.Fatal(err)
		}

		var fi os.FileInfo = <-fwatch.C
		var gotfi = newDirectFileInfo(fi)
		var nilfi *directFileInfo
		test.Assert(t, `fwatch.C`, nilfi, gotfi)

		fwatch.Stop()

		var expError = `no such file or directory`
		var gotError = fwatch.Err().Error()
		if !strings.Contains(gotError, expError) {
			t.Fatalf(`error does not contains %q`, expError)
		}
	})
}
