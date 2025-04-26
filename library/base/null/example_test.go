package null

import (
	"encoding/json"
	"fmt"
)

func Example() {
	i := 0
	ni := IntFrom(i)
	v, _ := ni.Value()
	fmt.Println(ni.Valid, v)

	var i2 *int
	ni = IntFromPtr(i2)
	v, _ = ni.Value()
	fmt.Println(ni.Valid, v)

	s := struct {
		Field1 int
		Field2 Int
		Field3 Int
	}{
		Field1: 0,
		Field2: IntFrom(0),
		Field3: IntFromPtr(nil),
	}
	str, _ := json.Marshal(s)
	fmt.Println(string(str))

	// Output:
	// true 0
	// false <nil>
	// {"Field1":0,"Field2":0,"Field3":null}
}
