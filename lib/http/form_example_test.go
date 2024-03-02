package http

import (
	"fmt"
	"net/url"

	"git.sr.ht/~shulhan/pakakeh.go/lib/math/big"
)

func ExampleMarshalForm() {
	type T struct {
		Rat    *big.Rat `form:"big.Rat"`
		String string   `form:"string"`
		Bytes  []byte   `form:"bytes"`
		Int    int      `form:""` // With empty tag.
		F64    float64  `form:"f64"`
		F32    float32  `form:"f32"`
		NotSet int16    `form:"notset"`
		Uint8  uint8    `form:" uint8 "`
		Bool   bool     // Without tag.
	}
	var (
		in = T{
			Rat:    big.NewRat(`1.2345`),
			String: `a_string`,
			Bytes:  []byte(`bytes`),
			Int:    1,
			F64:    6.4,
			F32:    3.2,
			Uint8:  2,
			Bool:   true,
		}

		out url.Values
		err error
	)

	out, err = MarshalForm(in)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(out.Encode())

	// Output:
	// Bool=true&Int=1&big.Rat=1.2345&bytes=bytes&f32=3.2&f64=6.4&notset=0&string=a_string&uint8=2
}

func ExampleUnmarshalForm() {
	type T struct {
		Rat    *big.Rat `form:"big.Rat"`
		String string   `form:"string"`
		Bytes  []byte   `form:"bytes"`
		Int    int      `form:""` // With empty tag.
		F64    float64  `form:"f64"`
		F32    float32  `form:"f32"`
		NotSet int16    `form:"notset"`
		Uint8  uint8    `form:" uint8 "`
		Bool   bool     // Without tag.
	}
	var (
		in = url.Values{}

		out    T
		ptrOut *T
		err    error
	)

	in.Set("big.Rat", "1.2345")
	in.Set("string", "a_string")
	in.Set("bytes", "bytes")
	in.Set("int", "1")
	in.Set("f64", "6.4")
	in.Set("f32", "3.2")
	in.Set("uint8", "2")
	in.Set("bool", "true")

	err = UnmarshalForm(in, &out)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", out)
	}

	// Set the struct without initialized.
	err = UnmarshalForm(in, &ptrOut)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", ptrOut)
	}

	// Output:
	// {Rat:1.2345 String:a_string Bytes:[98 121 116 101 115] Int:1 F64:6.4 F32:3.2 NotSet:0 Uint8:2 Bool:true}
	// &{Rat:1.2345 String:a_string Bytes:[98 121 116 101 115] Int:1 F64:6.4 F32:3.2 NotSet:0 Uint8:2 Bool:true}
}

func ExampleUnmarshalForm_error() {
	type T struct {
		Int int
	}

	var (
		in = url.Values{}

		out    T
		ptrOut *T
		err    error
	)

	// Passing out as unsetable by function.
	err = UnmarshalForm(in, out)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(out)
	}

	// Passing out as un-initialized pointer.
	err = UnmarshalForm(in, ptrOut)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(out)
	}

	// Set the field with invalid type.
	in.Set("int", "a")
	err = UnmarshalForm(in, &out)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(out)
	}

	// Output:
	// UnmarshalForm: expecting *T got http.T
	// UnmarshalForm: *http.T is not initialized
	// {0}

}

func ExampleUnmarshalForm_slice() {
	type SliceT struct {
		NotSlice    string   `form:"multi_value"`
		SliceString []string `form:"slice_string"`
		SliceInt    []int    `form:"slice_int"`
	}

	var (
		in = url.Values{}

		sliceOut    SliceT
		ptrSliceOut *SliceT
		err         error
	)

	in.Add("multi_value", "first")
	in.Add("multi_value", "second")
	in.Add("slice_string", "multi")
	in.Add("slice_string", "value")
	in.Add("slice_int", "123")
	in.Add("slice_int", "456")

	err = UnmarshalForm(in, &sliceOut)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", sliceOut)
	}

	err = UnmarshalForm(in, &ptrSliceOut)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", ptrSliceOut)
	}

	// Output:
	// {NotSlice:first SliceString:[multi value] SliceInt:[123 456]}
	// &{NotSlice:first SliceString:[multi value] SliceInt:[123 456]}
}

func ExampleUnmarshalForm_zero() {
	type T struct {
		Rat    *big.Rat `form:"big.Rat"`
		String string   `form:"string"`
		Bytes  []byte   `form:"bytes"`
		Int    int      `form:""` // With empty tag.
		F64    float64  `form:"f64"`
		F32    float32  `form:"f32"`
		NotSet int16    `form:"notset"`
		Uint8  uint8    `form:" uint8 "`
		Bool   bool     // Without tag.
	}
	var (
		in = url.Values{}

		out T
		err error
	)

	in.Set("big.Rat", "1.2345")
	in.Set("string", "a_string")
	in.Set("bytes", "bytes")
	in.Set("int", "1")
	in.Set("f64", "6.4")
	in.Set("f32", "3.2")
	in.Set("uint8", "2")
	in.Set("bool", "true")

	err = UnmarshalForm(in, &out)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", out)
	}

	in.Set("bool", "")
	in.Set("int", "")
	in.Set("uint8", "")
	in.Set("f32", "")
	in.Set("f64", "")
	in.Set("string", "")
	in.Set("bytes", "")
	in.Set("big.Rat", "")

	err = UnmarshalForm(in, &out)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", out)
	}

	// Output:
	// {Rat:1.2345 String:a_string Bytes:[98 121 116 101 115] Int:1 F64:6.4 F32:3.2 NotSet:0 Uint8:2 Bool:true}
	// {Rat:0 String: Bytes:[] Int:0 F64:0 F32:0 NotSet:0 Uint8:0 Bool:false}
}
