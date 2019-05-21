package ini

import (
	"fmt"
	"log"
)

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
