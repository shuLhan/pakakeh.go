package ini

import "fmt"

func ExampleIsValidVarName() {
	fmt.Println(IsValidVarName("1abcd"))
	fmt.Println(IsValidVarName("-abcd"))
	fmt.Println(IsValidVarName("_abcd"))
	fmt.Println(IsValidVarName(".abcd"))
	fmt.Println(IsValidVarName("a@bcd"))
	fmt.Println(IsValidVarName("a-b_c.d"))
	//Output:
	//false
	//false
	//false
	//false
	//false
	//true
}
