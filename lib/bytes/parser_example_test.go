package bytes_test

import (
	"fmt"

	libbytes "github.com/shuLhan/share/lib/bytes"
)

func ExampleParser_AddDelimiters() {
	var (
		content = []byte(` a = b ; c = d `)
		delims  = []byte(`=`)
		parser  = libbytes.NewParser(content, delims)
	)

	token, d := parser.ReadNoSpace()
	fmt.Printf("%s:%c\n", token, d)

	parser.AddDelimiters([]byte{';'})
	token, d = parser.ReadNoSpace()
	fmt.Printf("%s:%c\n", token, d)

	// Output:
	// a:=
	// b:;
}

func ExampleParser_Delimiters() {
	var (
		content = []byte(`a=b;c=d;`)
		delims  = []byte{'=', ';'}
		parser  = libbytes.NewParser(content, delims)
	)

	fmt.Printf("%s\n", parser.Delimiters())
	// Output:
	// =;
}

func ExampleParser_ReadN() {
	var (
		content = []byte(`a=b;c=d;`)
		delims  = []byte{'=', ';'}
		parser  = libbytes.NewParser(content, delims)
	)

	token, c := parser.ReadN(2)
	fmt.Printf("token:%s c:%d\n", token, c)

	token, c = parser.ReadN(0)
	fmt.Printf("token:%s c:%d\n", token, c)

	token, c = parser.ReadN(10)
	fmt.Printf("token:%s c:%d\n", token, c)
	// Output:
	// token:a= c:98
	// token: c:98
	// token:b;c=d; c:0
}

func ExampleParser_ReadNoSpace() {
	var (
		content = []byte(` a = b ;`)
		delims  = []byte(`=;`)
		parser  = libbytes.NewParser(content, delims)
	)

	for {
		token, d := parser.ReadNoSpace()
		fmt.Printf("%s:%d\n", token, d)
		if d == 0 {
			break
		}
	}
	// Output:
	// a:61
	// b:59
	// :0
}

func ExampleParser_RemoveDelimiters() {
	var (
		content = []byte(` a = b ; c = d `)
		delims  = []byte(`=;`)
		parser  = libbytes.NewParser(content, delims)
	)

	token, _ := parser.ReadNoSpace()
	fmt.Printf("%s\n", token)

	parser.RemoveDelimiters([]byte{';'})
	token, _ = parser.ReadNoSpace()
	fmt.Printf("%s\n", token)

	// Output:
	// a
	// b ; c
}

func ExampleParser_Reset() {
	var (
		content = []byte(`a.b.c;`)
		delims  = []byte(`.`)
		parser  = libbytes.NewParser(content, delims)
	)

	parser.Read()
	parser.Reset(content, delims)
	remain, pos := parser.Stop()
	fmt.Printf("remain:%s pos:%d\n", remain, pos)
	// Output:
	// remain:a.b.c; pos:0
}

func ExampleParser_SetDelimiters() {
	var (
		content = []byte(`a.b.c;`)
		delims  = []byte(`.`)
		parser  = libbytes.NewParser(content, delims)
		token   []byte
	)

	token, _ = parser.Read()
	fmt.Println(string(token))

	parser.SetDelimiters([]byte(`;`))

	token, _ = parser.Read()
	fmt.Println(string(token))

	// Output:
	// a
	// b.c
}

func ExampleParser_Skip() {
	var (
		content = []byte(`a = b; c = d;`)
		delims  = []byte{'=', ';'}
		parser  = libbytes.NewParser(content, delims)
		token   []byte
	)

	parser.Skip()
	token, _ = parser.ReadNoSpace()
	fmt.Println(string(token))

	parser.Skip()
	token, _ = parser.ReadNoSpace()
	fmt.Println(string(token))

	parser.Skip()
	token, _ = parser.ReadNoSpace()
	fmt.Println(string(token))

	// Output:
	// b
	// d
	//
}

func ExampleParser_SkipN() {
	var (
		content = []byte(`a=b;c=d;`)
		delims  = []byte{'=', ';'}
		parser  = libbytes.NewParser(content, delims)
		token   []byte
		c       byte
	)

	c = parser.SkipN(2)
	fmt.Printf("Skip: %c\n", c)
	token, _ = parser.ReadNoSpace()
	fmt.Println(string(token))

	c = parser.SkipN(2)
	fmt.Printf("Skip: %c\n", c)
	token, _ = parser.ReadNoSpace()
	fmt.Println(string(token))

	_ = parser.SkipN(2)
	token, _ = parser.ReadNoSpace()
	fmt.Println(string(token))

	// Output:
	// Skip: b
	// b
	// Skip: d
	// d
	//
}

func ExampleParser_SkipHorizontalSpaces() {
	var (
		content = []byte(" \t\r\fA. \nB.")
		delims  = []byte{'.'}
		parser  = libbytes.NewParser(content, delims)
	)

	parser.SkipHorizontalSpaces()
	token, d := parser.Read()
	fmt.Printf("token:%s delim:%c\n", token, d)

	parser.SkipHorizontalSpaces()
	token, d = parser.Read() // The token include \n.
	fmt.Printf("token:%s delim:%c\n", token, d)

	parser.SkipHorizontalSpaces()
	token, d = parser.Read() // The token include \n.
	fmt.Printf("token:%s delim:%d\n", token, d)

	// Output:
	// token:A delim:.
	// token:
	// B delim:.
	// token: delim:0
}

func ExampleParser_SkipSpaces() {
	var (
		content = []byte(" \t\r\fA. \nB.")
		delims  = []byte{'.'}
		parser  = libbytes.NewParser(content, delims)
	)

	parser.SkipSpaces()
	token, d := parser.Read()
	fmt.Printf("token:%s delim:%c\n", token, d)

	parser.SkipSpaces()
	token, d = parser.Read() // The token include \n.
	fmt.Printf("token:%s delim:%c\n", token, d)

	parser.SkipSpaces()
	token, d = parser.Read() // The token include \n.
	fmt.Printf("token:%s delim:%d\n", token, d)

	// Output:
	// token:A delim:.
	// token:B delim:.
	// token: delim:0
}

func ExampleParser_Stop() {
	var (
		content = []byte(`a.b.c;`)
		delims  = []byte(`.`)
		parser  = libbytes.NewParser(content, delims)

		remain []byte
		pos    int
	)

	parser.Read()
	remain, pos = parser.Stop()
	fmt.Printf("remain:%s pos:%d\n", remain, pos)

	parser.Reset(content, []byte(`;`))
	parser.Read()
	remain, pos = parser.Stop()
	fmt.Printf("remain:%s pos:%d\n", remain, pos)

	// Output:
	// remain:b.c; pos:2
	// remain: pos:6
}

func ExampleParser_UnreadN() {
	var (
		parser = libbytes.NewParser([]byte(`a,b.c/d`), []byte(`,./`))
		token  []byte
		c      byte
	)

	parser.Read()
	parser.Read()
	parser.Read()
	parser.Read() // All content should be readed now.

	c = parser.UnreadN(2) // Move the index to '/'.
	fmt.Printf("UnreadN(2): %c\n", c)

	token, c = parser.Read()
	fmt.Printf("Read: %s %c\n", token, c)

	// Position 99 greater than current index, this will reset index to 0.
	c = parser.UnreadN(99)
	fmt.Printf("UnreadN(99): %c\n", c)

	token, c = parser.Read()
	fmt.Printf("Read: %s %c\n", token, c)

	// Output:
	// UnreadN(2): /
	// Read:  /
	// UnreadN(99): a
	// Read: a ,
}