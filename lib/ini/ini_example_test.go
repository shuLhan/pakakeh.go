package ini

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	testInput = `
[section]
key=value1
key2=

[section "sub"]
key=value1
key=value2

[section]
key=value2
key2=false

[section "sub"]
key=value2
key=value3
`
)

func ExampleIni_Add() {
	ini := new(Ini)

	ini.Add("", "", "k1", "v1")
	ini.Add("s1", "", "", "v2")

	ini.Add("s1", "", "k1", "")
	ini.Add("s1", "", "k1", "v1")
	ini.Add("s1", "", "k1", "v2")
	ini.Add("s1", "", "k1", "v1")

	ini.Add("s1", "sub", "k1", "v1")
	ini.Add("s1", "sub", "k1", "v1")

	ini.Add("s2", "sub", "k1", "v1")

	err := ini.Write(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// [s1]
	// k1 =
	// k1 = v1
	// k1 = v2
	//
	// [s1 "sub"]
	// k1 = v1
	//
	// [s2 "sub"]
	// k1 = v1
}

func ExampleIni_Gets() {
	input := []byte(`
[section]
key=value1

[section "sub"]
key=value2

[section]
key=value3

[section "sub"]
key=value4
key=value2
`)

	inis, _ := Parse(input)

	fmt.Println(inis.Gets("section", "", "key"))
	fmt.Println(inis.Gets("section", "sub", "key"))
	//Output:
	//[value1 value3]
	//[value2 value4 value2]
}

func ExampleIni_GetsUniq() {
	input := []byte(`
[section]
key=value1

[section "sub"]
key=value2

[section]
key=value3

[section "sub"]
key=value4
key=value2
`)

	inis, _ := Parse(input)

	fmt.Println(inis.GetsUniq("section", "", "key", true))
	fmt.Println(inis.GetsUniq("section", "sub", "key", true))
	//Output:
	//[value1 value3]
	//[value2 value4]
}

func ExampleIni_AsMap() {
	input := []byte(`
[section]
key=value1
key2=

[section "sub"]
key=value1
key2=

[section]
key=value2
key2=false

[section "sub"]
key=value2
key=value3
`)

	inis, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	iniMap := inis.AsMap("", "")

	for k, v := range iniMap {
		fmt.Println(k, "=", v)
	}

	iniMap = inis.AsMap("section", "sub")

	fmt.Println()
	for k, v := range iniMap {
		fmt.Println(k, "=", v)
	}

	// Unordered output:
	// section::key = [value1 value2]
	// section::key2 = [ false]
	// section:sub:key = [value1 value2 value3]
	// section:sub:key2 = []
	//
	// key = [value1 value2 value3]
	// key2 = []
}

func ExampleMarshal() {
	ptrString := "b"
	ptrInt := int(2)

	type U struct {
		String string `ini:"::string"`
		Int    int    `ini:"::int"`
	}

	t := struct {
		String      string            `ini:"section::string"`
		Int         int               `ini:"section::int"`
		Bool        bool              `ini:"section::bool"`
		Duration    time.Duration     `ini:"section::duration"`
		Time        time.Time         `ini:"section::time" layout:"2006-01-02 15:04:05"`
		SliceString []string          `ini:"section:slice:string"`
		SliceInt    []int             `ini:"section:slice:int"`
		SliceUint   []uint            `ini:"section:slice:uint"`
		SliceBool   []bool            `ini:"section:slice:bool"`
		SliceStruct []U               `ini:"slice:OfStruct"`
		MapString   map[string]string `ini:"section:mapstring"`
		MapInt      map[string]int    `ini:"section:mapint"`
		PtrString   *string           `ini:"section:pointer"`
		PtrInt      *int              `ini:"section:pointer"`
	}{
		String:      "a",
		Int:         1,
		Bool:        true,
		Duration:    time.Minute,
		Time:        time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		SliceString: []string{"c", "d"},
		SliceInt:    []int{2, 3},
		SliceUint:   []uint{4, 5},
		SliceBool:   []bool{true, false},
		SliceStruct: []U{{
			String: "U.string 1",
			Int:    1,
		}, {
			String: "U.string 2",
			Int:    2,
		}},
		MapString: map[string]string{
			"k": "v",
		},
		MapInt: map[string]int{
			"keyInt": 6,
		},
		PtrString: &ptrString,
		PtrInt:    &ptrInt,
	}

	iniText, err := Marshal(&t)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", iniText)
	// Output:
	// [section]
	// string = a
	// int = 1
	// bool = true
	// duration = 1m0s
	// time = 2006-01-02 15:04:05
	//
	// [section "slice"]
	// string = c
	// string = d
	// int = 2
	// int = 3
	// uint = 4
	// uint = 5
	// bool = true
	// bool = false
	//
	// [slice "OfStruct"]
	// string = U.string 1
	// int = 1
	//
	// [slice "OfStruct"]
	// string = U.string 2
	// int = 2
	//
	// [section "mapstring"]
	// k = v
	//
	// [section "mapint"]
	// keyint = 6
	//
	// [section "pointer"]
	// ptrstring = b
	// ptrint = 2
}

func ExampleUnmarshal() {
	iniText := `
[section]
string = a
int = 1
bool = true
duration = 1s
time = 2006-01-02 15:04:05

[section "slice"]
string = c
string = d
int = 2
int = 3
bool = true
bool = false
uint = 4
uint = 5

[section "mapstring"]
k = v

[section "mapint"]
k = 6

[section "pointer"]
ptrstring = b
ptrint = 2
`
	t := struct {
		String      string            `ini:"section::string"`
		Int         int               `ini:"section::int"`
		Bool        bool              `ini:"section::bool"`
		Duration    time.Duration     `ini:"section::duration"`
		Time        time.Time         `ini:"section::time" layout:"2006-01-02 15:04:05"`
		SliceString []string          `ini:"section:slice:string"`
		SliceInt    []int             `ini:"section:slice:int"`
		SliceUint   []uint            `ini:"section:slice:uint"`
		SliceBool   []bool            `ini:"section:slice:bool"`
		MapString   map[string]string `ini:"section:mapstring"`
		MapInt      map[string]int    `ini:"section:mapint"`
		PtrString   *string           `ini:"section:pointer"`
		PtrInt      *int              `ini:"section:pointer"`
	}{}

	err := Unmarshal([]byte(iniText), &t)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("String: %v\n", t.String)
	fmt.Printf("Int: %v\n", t.Int)
	fmt.Printf("Bool: %v\n", t.Bool)
	fmt.Printf("Duration: %v\n", t.Duration)
	fmt.Printf("Time: %v\n", t.Time)
	fmt.Printf("SliceString: %v\n", t.SliceString)
	fmt.Printf("SliceInt: %v\n", t.SliceInt)
	fmt.Printf("SliceUint: %v\n", t.SliceUint)
	fmt.Printf("SliceBool: %v\n", t.SliceBool)
	fmt.Printf("MapString: %v\n", t.MapString)
	fmt.Printf("MapInt: %v\n", t.MapInt)
	fmt.Printf("PtrString: %v\n", *t.PtrString)
	fmt.Printf("PtrInt: %v\n", *t.PtrInt)
	// Output:
	// String: a
	// Int: 1
	// Bool: true
	// Duration: 1s
	// Time: 2006-01-02 15:04:05 +0000 UTC
	// SliceString: [c d]
	// SliceInt: [2 3]
	// SliceUint: [4 5]
	// SliceBool: [true false]
	// MapString: map[k:v]
	// MapInt: map[k:6]
	// PtrString: b
	// PtrInt: 2
}

func ExampleIni_Prune() {
	input := []byte(`
[section]
key=value1 # comment
key2= ; another comment

[section "sub"]
key=value1

; here is comment on section
[section]
key=value2
key2=false

[section "sub"]
key=value2
key=value1
`)

	in, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	in.Prune()

	err = in.Write(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	// Output:
	// [section]
	// key = value1
	// key2 = true
	// key = value2
	// key2 = false
	//
	// [section "sub"]
	// key = value2
	// key = value1
}

func ExampleIni_Rebase() {
	input := []byte(`
		[section]
		key=value1
		key2=

		[section "sub"]
		key=value1
`)

	other := []byte(`
		[section]
		key=value2
		key2=false

		[section "sub"]
		key=value2
		key=value1
`)

	in, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	in2, err := Parse(other)
	if err != nil {
		log.Fatal(err)
	}

	in.Prune()
	in2.Prune()

	in.Rebase(in2)

	err = in.Write(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
	// Output:
	// [section]
	// key = value1
	// key2 = true
	// key = value2
	// key2 = false
	//
	// [section "sub"]
	// key = value2
	// key = value1
}

func ExampleIni_Section() {
	input := []byte(`
[section]
key=value1 # comment
key2= ; another comment

[section "sub"]
key=value1

[section] ; here is comment on section
key=value2
key2=false

[section "sub"]
key=value2
key=value1
`)

	ini, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	sec := ini.Section("section", "")
	for _, v := range sec.vars {
		fmt.Printf("%s=%s\n", v.key, v.value)
	}
	// Output:
	// key=value1
	// key2=
	// key=value2
	// key2=false
}

func ExampleIni_Set() {
	input := []byte(`
[section]
key=value1 # comment
key2= ; another comment

[section "sub"]
key=value1

[section] ; here is comment on section
key=value2
key2=false

[section "sub"]
key=value2
key=value1
`)

	ini, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	ini.Set("", "sub", "key", "value3")
	ini.Set("sectionnotexist", "sub", "key", "value3")
	ini.Set("section", "sub", "key", "value3")
	ini.Set("section", "", "key", "value4")
	ini.Set("section", "", "keynotexist", "value4")

	err = ini.Write(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	//Output:
	//[section]
	//key=value1 # comment
	//key2= ; another comment
	//
	//[section "sub"]
	//key=value1
	//
	//[section] ; here is comment on section
	//key=value4
	//key2=false
	//
	//keynotexist = value4
	//
	//[section "sub"]
	//key=value2
	//key=value3
	//
	//[sectionnotexist "sub"]
	//key = value3
}

func ExampleIni_Subs() {
	input := []byte(`
[section]
key=value1 # comment
key2= ; another comment

[section "sub"]
key=value1

[section] ; here is comment on section
key=value2
key2=false

[section "sub"]
key=value2
key=value1
`)

	ini, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	subs := ini.Subs("section")

	for _, sub := range subs {
		fmt.Println(sub.SubName(), sub.Vals("key"))
	}
	// Output:
	// sub [value2 value1]
}

func ExampleIni_Unset() {
	input := []byte(`
[section]
key=value1 # comment
key2= ; another comment

[section "sub"]
key=value1

; here is comment on section
[section]
key=value2
key2=false

[section "sub"]
key=value2
key=value1
`)

	ini, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	ini.Unset("", "sub", "keynotexist")
	ini.Unset("sectionnotexist", "sub", "keynotexist")
	ini.Unset("section", "sub", "keynotexist")
	ini.Unset("section", "sub", "key")
	ini.Unset("section", "", "keynotexist")
	ini.Unset("section", "", "key")

	err = ini.Write(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}

	//Output:
	//[section]
	//key=value1 # comment
	//key2= ; another comment
	//
	//[section "sub"]
	//key=value1
	//
	//; here is comment on section
	//[section]
	//key2=false
	//
	//[section "sub"]
	//key=value2
}

func ExampleIni_Val() {
	input := `
[section]
key=value1
key2=

[section "sub"]
key=value1

[section]
key=value2
key2=false

[section "sub"]
key=value2
key=value3
`

	ini, err := Parse([]byte(input))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ini.Val("section:sub:key"))
	fmt.Println(ini.Val("section:sub:key2"))
	fmt.Println(ini.Val("section::key"))
	fmt.Println(ini.Val("section:key"))
	// Output:
	// value3
	//
	// value2
	//
}

func ExampleIni_Vals() {
	ini, err := Parse([]byte(testInput))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ini.Vals("section:key"))
	fmt.Println(ini.Vals("section::key"))
	fmt.Println(ini.Vals("section:sub:key2"))
	fmt.Println(ini.Vals("section:sub:key"))
	// Output:
	// []
	// [value1 value2]
	// []
	// [value1 value2 value2 value3]
}

func ExampleIni_ValsUniq() {
	ini, err := Parse([]byte(testInput))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ini.ValsUniq("section:key", true))
	fmt.Println(ini.ValsUniq("section::key", true))
	fmt.Println(ini.ValsUniq("section:sub:key2", true))
	fmt.Println(ini.ValsUniq("section:sub:key", true))
	// Output:
	// []
	// [value1 value2]
	// []
	// [value1 value2 value3]
}

func ExampleIni_Vars() {
	ini, err := Parse([]byte(testInput))
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range ini.Vars("section:") {
		fmt.Println(k, "=", v)
	}

	fmt.Println()
	for k, v := range ini.Vars("section:sub") {
		fmt.Println(k, "=", v)
	}
	// Unordered output:
	// section::key = value2
	// section::key2 = false
	//
	// key = value3
}
