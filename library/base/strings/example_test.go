package strings

import "fmt"

func ExampleSnakeNameToBigCamelName() {
	name := SnakeNameToBigCamelName("user_center")
	fmt.Printf("%s\n", name)

	// OutPut:
	// UserCenter
}
