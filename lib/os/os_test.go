package os

import (
	"os"
	"testing"
)

// TestStat test to see the difference between Stat and Lstat.
func TestStat(t *testing.T) {
	t.Skip()

	var (
		files = []string{
			`testdata/exp`,
			`testdata/symlink_file`,
			`testdata/symlink_symlink_file`,
			`testdata/exp_dir`,
			`testdata/symlink_dir`,
			`testdata/symlink_symlink_dir`,
		}

		fi   os.FileInfo
		file string
		err  error
	)

	for _, file = range files {
		fi, err = os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf(` Stat %s: mode:%s, size:%d, is_dir:%t, is_symlink:%t, modtime:%s`,
			file, fi.Mode(), fi.Size(), fi.IsDir(),
			fi.Mode()&os.ModeSymlink != 0, fi.ModTime())

		fi, err = os.Lstat(file)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf(`Lstat %s: mode:%s, size:%d, is_dir:%t, is_symlink:%t, modtime:%s`,
			file, fi.Mode(), fi.Size(), fi.IsDir(),
			fi.Mode()&os.ModeSymlink != 0, fi.ModTime())
	}
}
