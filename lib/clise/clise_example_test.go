package clise

import "fmt"

func ExampleClise_RecentSlice() {
	c := New(5)
	c.Push(1, 2, 3)
	fmt.Printf("%v\n", c.RecentSlice())
	c.Push(4, 5, 6, 7)
	fmt.Printf("%v\n", c.RecentSlice())
	//Output:
	//[1 2 3]
	//[6 7]
}

func ExampleClise_Reset() {
	c := New(5)
	c.Push(1, 2, 3, 4, 5)
	fmt.Printf("%v\n", c.Slice())
	c.Reset()
	c.Push(1)
	fmt.Printf("%v\n", c.Slice())
	//Output:
	//[1 2 3 4 5]
	//[1]
}

func ExampleClise_Slice() {
	c := New(5)
	c.Push(1, 2)
	fmt.Printf("%v\n", c.Slice())
	c.Push(3, 4, 5)
	fmt.Printf("%v\n", c.Slice())
	c.Push(6)
	fmt.Printf("%v\n", c.Slice())
	c.Push(7, 8, 9, 10)
	fmt.Printf("%v\n", c.Slice())
	//Output:
	//[1 2]
	//[1 2 3 4 5]
	//[2 3 4 5 6]
	//[6 7 8 9 10]
}
