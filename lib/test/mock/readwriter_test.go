package mock

import (
	"io"
	"testing"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
)

func TestReadWriter(t *testing.T) {
	var (
		mockrw               = ReadWriter{}
		iorw   io.ReadWriter = &mockrw

		exp = `content`
		buf = make([]byte, 7)
	)

	mockrw.BufRead.WriteString(`content of read buffer`)
	_, _ = iorw.Read(buf)
	test.Assert(t, `Read`, exp, string(buf))

	_, _ = iorw.Write(buf)
	test.Assert(t, `Write`, exp, mockrw.BufWrite.String())

	_, _ = mockrw.WriteString(` of write buffer`)
	exp = `content of write buffer`
	test.Assert(t, `Write`, exp, mockrw.BufWrite.String())

	mockrw.Reset()

	_, _ = iorw.Read(buf)
	test.Assert(t, `Read buffer after reset`, ``, mockrw.BufRead.String())
	test.Assert(t, `Write buffer after reset`, ``, mockrw.BufWrite.String())
}
