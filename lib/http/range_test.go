package http

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseMultipartRange(t *testing.T) {
	var (
		listTestData []*test.Data
		err          error
	)

	listTestData, err = test.LoadDataDir(`./testdata/range/`)
	if err != nil {
		t.Fatal(err)
	}

	var (
		boundary = `zxcv`

		tdata  *test.Data
		reader *bytes.Reader
		r      *Range
		pos    RangePosition
		vbyte  []byte
		got    strings.Builder
	)

	for _, tdata = range listTestData {
		t.Log(tdata.Name)

		vbyte = tdata.Input[`body`]
		vbyte = bytes.ReplaceAll(vbyte, []byte("\n"), []byte("\r\n"))

		reader = bytes.NewReader(vbyte)

		r, err = ParseMultipartRange(reader, boundary)
		if err != nil {
			vbyte = tdata.Output[`error`]
			test.Assert(t, `error`, string(vbyte), err.Error())
			continue
		}

		got.Reset()
		for _, pos = range r.Positions() {
			fmt.Fprintf(&got, "%s: %s\n", pos.String(), pos.Content())
		}

		vbyte = tdata.Output[`expected`]
		test.Assert(t, `content`, string(vbyte), got.String())
	}
}
