package time_test

import (
	"fmt"

	"git.sr.ht/~shulhan/pakakeh.go/lib/time"
)

func ExampleClock_After() {
	var (
		c = time.CreateClock(1, 2, 3)
	)

	fmt.Println(c.After(time.CreateClock(0, 2, 3)))
	fmt.Println(c.After(time.CreateClock(1, 1, 3)))
	fmt.Println(c.After(time.CreateClock(1, 2, 2)))

	fmt.Println(c.After(time.CreateClock(1, 2, 3))) // Equal Clock is not an After.
	fmt.Println(c.After(time.CreateClock(1, 2, 4)))
	fmt.Println(c.After(time.CreateClock(1, 3, 0)))
	fmt.Println(c.After(time.CreateClock(2, 0, 0)))
	// Output:
	// true
	// true
	// true
	// false
	// false
	// false
	// false
}

func ExampleClock_Before() {
	var (
		c = time.CreateClock(1, 2, 3)
	)

	fmt.Println(c.Before(time.CreateClock(0, 2, 3)))
	fmt.Println(c.Before(time.CreateClock(1, 1, 3)))
	fmt.Println(c.Before(time.CreateClock(1, 2, 2)))
	fmt.Println(c.Before(time.CreateClock(1, 2, 3))) // Equal Clock is not a Before.
	fmt.Println(c.Before(time.CreateClock(1, 2, 4)))
	fmt.Println(c.Before(time.CreateClock(1, 3, 0)))
	// Output:
	// false
	// false
	// false
	// false
	// true
	// true
}

func ExampleClock_Equal() {
	var (
		c = time.CreateClock(1, 2, 3)
	)

	fmt.Println(c.Equal(time.CreateClock(0, 2, 3)))
	fmt.Println(c.Equal(time.CreateClock(1, 2, 2)))
	fmt.Println(c.Equal(time.CreateClock(1, 2, 3)))
	fmt.Println(c.Equal(time.CreateClock(1, 2, 4)))
	fmt.Println(c.Equal(time.CreateClock(1, 3, 0)))
	// Output:
	// false
	// false
	// true
	// false
	// false
}

func ExampleClock_Hour() {
	var c = time.CreateClock(1, 2, 3)
	fmt.Println(c.Hour())
	// Output: 1
}

func ExampleClock_Minute() {
	var c = time.CreateClock(1, 2, 3)
	fmt.Println(c.Minute())
	// Output: 2
}

func ExampleClock_Second() {
	var c = time.CreateClock(1, 2, 3)
	fmt.Println(c.Second())
	// Output: 3
}

func ExampleClock_String() {
	var c = time.CreateClock(1, 2, 3)
	fmt.Println(c.String())
	// Output: 01:02:03
}

func ExampleCreateClock() {
	var c = time.CreateClock(-1, 2, 3) // The hour valid is invalid.
	fmt.Println(c)
	// Output:
	// 00:02:03
}

func ExampleParseClock() {
	var c = time.ParseClock(`01:23:60`) // The second value is invalid.
	fmt.Println(c)
	// Output:
	// 01:23:00
}

func ExampleSortClock() {
	var (
		list = []time.Clock{
			time.CreateClock(3, 2, 1),
			time.CreateClock(2, 3, 1),
			time.CreateClock(1, 2, 3),
		}
	)
	time.SortClock(list)
	fmt.Println(list)
	// Output:
	// [01:02:03 02:03:01 03:02:01]
}
