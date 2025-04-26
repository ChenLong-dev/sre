package goroutine

import (
	"context"
	"errors"
	"fmt"
	"time"

	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func ExampleErrGroup_Go() {
	Init(&Config{
		Config: &render.Config{
			Stdout:        false,
			StdoutPattern: "[%T] [%t] [%U] [status: %s] %S  Group: %N:%I , Current: %n:%i , %E",
		},
	})

	ctx := context.Background()

	// 这种方式，当某个协程报错时，会调用cancel，取消其他协程
	// 常用
	eg := WithContext(ctx, "Test")

	// 这种方式，当某个协程报错时，不会取消其他协程
	// 不常用，容易浪费资源
	//eg := New("Test")

	eg.Go(ctx, "Test1", func(c context.Context) error {
		time.Sleep(2)

		return nil
	})
	eg.Go(ctx, "Test2", func(c context.Context) error {
		time.Sleep(5)

		return errors.New("this is a error")
	})

	err := eg.Wait()

	fmt.Printf("%s\n", err)

	// OutPut:
	// this is a error
}
