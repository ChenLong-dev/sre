package slice

import "fmt"

func ExampleStringSliceContains() {
	strSli := []string{"name", "age", "gender"}
	name := StrSliceContains(strSli, "name")
	fmt.Printf("%v\n", name)

	// OutPut:
	// true
}
