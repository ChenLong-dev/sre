package cm

import "fmt"

func ExampleClient_GetOriginFile() {
	data, err := DefaultClient().GetOriginFile(
		"projects/app_framework/stg/config.yaml",
		"b413878fbdf5d2d658b386f9b4de109ac5dc2f40",
	)
	if err != nil {
		return
	}
	fmt.Printf("%s\n", data)
}
