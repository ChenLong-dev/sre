package errcode

import (
	"fmt"
	"github.com/pkg/errors"
)

func ErrorFunc(s string) error {
	if s == "empty" {
		return errors.Wrapf(UnknownError, "string is %s", s)
	} else {
		return nil
	}
}

func CustomErrorFunc(s string) error {
	if s == "empty" {
		err := UnknownError.WithFrontCode(500).WithMessage("this is empty")
		return errors.Wrapf(err, "string is %s", s)
	} else {
		return nil
	}
}

func ExampleGroup() {
	group := NewGroup(InvalidParams).
		AddChildren(
			errors.Wrap(MysqlError, "user id:1 not found"),
			errors.Wrap(MongoError, "user id:2 not found"),
		)

	fmt.Printf("%s\n", group)
	fmt.Printf("%v\n", EqualError(InvalidParams, group))
	fmt.Printf("%s\n", Cause(group))

	// Output:
	// 1060003:参数错误
	// true
	// 1060003:参数错误
}

func ExampleError() {
	err := ErrorFunc("empty")
	fmt.Printf("%s\n", err)
	fmt.Printf("%v\n", EqualError(UnknownError, err))
	fmt.Printf("%s\n", Cause(err))

	// Output:
	// string is empty: 1060000:未知错误
	// true
	// 1060000:未知错误
}

func ExampleCustomError() {
	err := CustomErrorFunc("empty")
	fmt.Printf("%s\n", err)
	fmt.Printf("%v\n", EqualError(UnknownError, err))
	fmt.Printf("%s\n", Cause(err))
	fmt.Printf("%d\n", Cause(err).FrontendCode())

	// Output:
	// string is empty: 1060000:this is empty
	// true
	// 1060000:this is empty
	// 500
}
