package reflect

import "fmt"

func ExampleStructToMap() {
	s := struct {
		FieldA string
		FieldB int
	}{
		FieldA: "a",
		FieldB: 1,
	}

	m, err := StructToMap(s)
	if err != nil {
		return
	}
	fmt.Printf("%v %v\n", m["FieldA"], m["FieldB"])

	// OutPut:
	// a 1
}
