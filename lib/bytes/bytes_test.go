package bytes

import (
	"bytes"
	"testing"

	"github.com/shuLhan/share/lib/test"
	libtext "github.com/shuLhan/share/lib/text"
)

func TestToLower(t *testing.T) {
	cases := []struct {
		in  []byte
		exp []byte
	}{{
		in:  []byte("@ABCDEFG"),
		exp: []byte("@abcdefg"),
	}, {
		in:  []byte("@ABCDEFG12345678"),
		exp: []byte("@abcdefg12345678"),
	}, {
		in:  []byte("@ABCDEFGhijklmno12345678"),
		exp: []byte("@abcdefghijklmno12345678"),
	}, {
		in:  []byte("@ABCDEFGhijklmnoPQRSTUVW12345678"),
		exp: []byte("@abcdefghijklmnopqrstuvw12345678"),
	}, {
		in:  []byte("@ABCDEFGhijklmnoPQRSTUVWxyz{12345678"),
		exp: []byte("@abcdefghijklmnopqrstuvwxyz{12345678"),
	}}

	for _, c := range cases {
		ToLower(&c.in)
		test.Assert(t, "ToLower", c.exp, c.in, true)
	}
}

var randomInput256 = libtext.Random([]byte(libtext.HexaLetters), 256)

func BenchmarkToLowerStd(b *testing.B) {
	in := make([]byte, len(randomInput256))
	copy(in, randomInput256)

	b.ResetTimer()

	for x := 0; x < b.N; x++ {
		bytes.ToLower(in)
	}
}

func BenchmarkToLower(b *testing.B) {
	in := make([]byte, len(randomInput256))
	copy(in, randomInput256)

	b.ResetTimer()

	for x := 0; x < b.N; x++ {
		ToLower(&in)
		copy(in, randomInput256)
	}
}
