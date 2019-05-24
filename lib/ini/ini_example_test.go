package ini

import (
	"fmt"
	"log"
)

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
	//[value2 value4]
}

func ExampleIni_AsMap() {
	input := []byte(`
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
`)

	inis, err := Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	iniMap := inis.AsMap()

	for k, v := range iniMap {
		fmt.Println(k, "=", v)
	}
	// Unordered output:
	// section::key = [value1 value2]
	// section::key2 = [true false]
	// section:sub:key = [value1 value2 value3]
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

	for _, sec := range in.secs {
		fmt.Printf("%s", sec)
		for _, v := range sec.Vars {
			fmt.Printf("%s", v)
		}
	}
	// Output:
	// [section]
	// key = value1
	// key2 = true
	// key = value2
	// key2 = false
	// [section "sub"]
	// key = value2
	// key = value1
}
