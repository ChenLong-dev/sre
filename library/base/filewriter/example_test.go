package filewriter

import "fmt"

func ExampleNewSingleFileWriter() {
	fw := NewSingleFileWriter(
		logdir+"/test-rotate-exists/info.log",
		1024*1024, 1024*1024, 5,
	)

	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i)
	}

	s, err := fw.Write(data)
	if err != nil {
		return
	}
	fmt.Printf("%d\n", s)

	// OutPut:
	// 1024
}
