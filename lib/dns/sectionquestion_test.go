package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestQuestionMarshalName(t *testing.T) {
	cases := []struct {
		desc string
		in   *SectionQuestion
		exp  []byte
	}{{
		desc: "Empty name",
		in: &SectionQuestion{
			Type:  QueryTypeA,
			Class: QueryClassIN,
		},
		exp: []byte{
			0x00, 0x00, 0x01, 0x00, 0x01,
		},
	}, {
		desc: "Single domain name",
		in: &SectionQuestion{
			Name:  []byte("kilabit"),
			Type:  QueryTypeA,
			Class: QueryClassIN,
		},
		exp: []byte{
			0x07, 'k', 'i', 'l', 'a', 'b', 'i', 't', 0x00,
			0x00, 0x01, 0x00, 0x01,
		},
	}, {
		desc: "Two domain names",
		in: &SectionQuestion{
			Name:  []byte("kilabit.info"),
			Type:  QueryTypeA,
			Class: QueryClassIN,
		},
		exp: []byte{
			0x07, 'k', 'i', 'l', 'a', 'b', 'i', 't',
			0x04, 'i', 'n', 'f', 'o',
			0x00,
			0x00, 0x01, 0x00, 0x01,
		},
	}, {
		desc: "Three domain names",
		in: &SectionQuestion{
			Name:  []byte("mail.kilabit.info"),
			Type:  QueryTypeA,
			Class: QueryClassIN,
		},
		exp: []byte{
			0x04, 'm', 'a', 'i', 'l',
			0x07, 'k', 'i', 'l', 'a', 'b', 'i', 't',
			0x04, 'i', 'n', 'f', 'o',
			0x00,
			0x00, 0x01, 0x00, 0x01,
		},
	}}

	for _, c := range cases {
		t.Log(c.desc)

		got, err := c.in.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		test.Assert(t, "packet", c.exp, got, true)
	}
}
