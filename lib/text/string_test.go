package text

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestStringJSONEscape(t *testing.T) {
	cases := []struct {
		in  string
		exp string
	}{{
		in:  "",
		exp: "",
	}, {
		in: `	this\ is
		//\"test"`,
		exp: `\tthis\\ is\n\t\t\/\/\\\"test\"`,
	}, {
		in: ` `,
		exp: `\u0002\b\f\u000E\u000F\u0010\u0014\u001E\u001F `,
	}}

	var got string

	for _, c := range cases {
		t.Log(c)

		got = StringJSONEscape(c.in)

		test.Assert(t, "", c.exp, got, true)
	}
}
