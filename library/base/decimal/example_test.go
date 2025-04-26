package decimal

import "fmt"

func ExampleToFixed() {
	n := float64(1.123)
	n, err := ToFixed(1, n)
	if err != nil {
		return
	}
	fmt.Printf("%v\n", n)

	// OutPut:
	// 1.1
}
