package ini

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestOpen(t *testing.T) {
	cases := []struct {
		desc   string
		inFile string
		expErr string
	}{{
		desc:   "With invalid file",
		expErr: "open : no such file or directory",
	}, {
		desc:   "With valid file",
		inFile: "testdata/input.ini",
	}}

	for _, c := range cases {
		t.Log(c.desc)

		_, err := Open(c.inFile)
		if err != nil {
			test.Assert(t, c.expErr, err.Error(), true)
			continue
		}
	}
}
