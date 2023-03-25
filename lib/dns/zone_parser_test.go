package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestZoneParserDecodeString(t *testing.T) {
	type testCase struct {
		in       []byte
		exp      string
		expError string
	}

	var cases = []testCase{{
		in:  []byte(`"a\\b \"c\."`),
		exp: `a\\b "c\.`,
	}, {
		in:       []byte(`a\12a`),
		expError: `decodeString: invalid digits: \12a`,
	}, {
		in:       []byte(`a\12`),
		expError: `decodeString: invalid digits length: \12`,
	}, {
		in:       []byte(`a\999`),
		expError: `decodeString: invalid octet: \999`,
	}, {
		in:  []byte(`a\032b c`),
		exp: `a b`,
	}, {
		in:  []byte(`a\032b\.c`),
		exp: `a b.c`,
	}}

	var (
		zp = &zoneParser{}

		c   testCase
		got []byte
		err error
	)
	for _, c = range cases {
		got, err = zp.decodeString(c.in)
		if err != nil {
			test.Assert(t, `error`, c.expError, err.Error())
			continue
		}
		test.Assert(t, string(c.in), c.exp, string(got))
	}
}
