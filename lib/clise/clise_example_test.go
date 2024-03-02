package clise

import "fmt"

func ExampleClise_Pop() {
	var (
		c    = New(5)
		item interface{}
	)

	c.Push(1, 2, 3, 4, 5, 6)
	item = c.Pop()
	for item != nil {
		fmt.Println(item)
		item = c.Pop()
	}
	// Output:
	// 6
	// 5
	// 4
	// 3
	// 2
}

func ExampleClise_RecentSlice() {
	var c = New(5)
	c.Push(1, 2, 3)
	fmt.Println(c.RecentSlice())
	c.Push(4, 5, 6, 7)
	fmt.Println(c.RecentSlice())
	// Output:
	// [1 2 3]
	// [6 7]
}

func ExampleClise_Reset() {
	var c = New(5)
	c.Push(1, 2, 3, 4, 5)
	fmt.Println(c.Slice())
	c.Reset()
	c.Push(1)
	fmt.Println(c.Slice())
	// Output:
	// [1 2 3 4 5]
	// [1]
}

func ExampleClise_Slice() {
	var c = New(5)
	c.Push(1, 2)
	fmt.Println(c.Slice())
	c.Push(3, 4, 5)
	fmt.Println(c.Slice())
	c.Push(6)
	fmt.Println(c.Slice())
	c.Push(7, 8, 9, 10)
	fmt.Println(c.Slice())
	// Output:
	// [1 2]
	// [1 2 3 4 5]
	// [2 3 4 5 6]
	// [6 7 8 9 10]
}
