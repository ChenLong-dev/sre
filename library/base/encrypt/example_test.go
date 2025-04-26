package encrypt

import "fmt"

func ExampleEncodeID() {
	data, err := EncodeID("normal", 1)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", data)

	// OutPut:
	// DJ81kjR7
}

func ExampleDecodeID() {
	id, err := DecodeID("normal", "DJ81kjR7")
	if err != nil {
		return
	}
	fmt.Printf("%d\n", id)

	// OutPut:
	// 1
}
