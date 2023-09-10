package text

import (
	"reflect"
	"testing"
)

func TestChunk_MarshalJSON(t *testing.T) {
	type testCase struct {
		exp   string
		chunk Chunk
	}

	var cases = []testCase{{
		chunk: Chunk{
			StartAt: 1,
			V:       []byte("<script>a\"\\ \b\f\n\r\tz"),
		},
		exp: `{"StartAt":1,"V":"<script>a\"\\ \b\f\n\r\tz"}`,
	}}

	var (
		c   testCase
		got []byte
		err error
	)

	for _, c = range cases {
		got, err = c.chunk.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(c.exp, string(got)) {
			t.Fatalf(`want %s, got %s`, c.exp, got)
		}
	}
}
