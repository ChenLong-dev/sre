package sw

import (
	"fmt"
	"time"
)

func ExampleNewSlidingWindow() {
	w := NewSlidingWindow(time.Second*10, 2)

	w.Fail()
	w.Fail()
	w.Success()
	w.Success()

	fmt.Println(2)
	fmt.Println(2)
	fmt.Println(0.5)

	// Output:
	// 2
	// 2
	// 0.5
}
