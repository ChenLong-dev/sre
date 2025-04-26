package runtime

import (
	"fmt"
	"strings"
)

func ExampleGetCaller() {
	result := GetCaller(1)
	fmt.Printf("%v\n", strings.Contains(result, "base/runtime/example_test.go"))

	// OutPut:
	// true
}

func ExampleGetFullCallers() {
	callers := GetFullCallers()
	fmt.Printf("%v\n", strings.Contains(callers[1], "base/runtime/caller.go"))
	fmt.Printf("%v\n", strings.Contains(callers[2], "base/runtime/example_test.go"))

	// OutPut:
	// true
	// true
}
