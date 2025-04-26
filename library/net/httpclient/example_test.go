package httpclient

import (
	"context"
	"fmt"
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func ExampleClient_GetJSON() {
	c := Config{
		RequestTimeout:  ctime.Duration(5 * time.Second),
		RequestBodyOut:  false,
		ResponseBodyOut: false,
		Config: &render.Config{
			Stdout:        false,
			StdoutPattern: "[%T] [%t] [%U] [status: %s] [duration: %D] %S  URL: %M::%u , Header: %h , RequestBody: %b , ResponseBody: %B",
		},
	}
	client := NewHttpClient(&c)

	result := make([]GithubRepo, 0)
	err := client.GetJSON(
		context.Background(),
		"https://api.github.com/users/cmonoceros/repos",
		NewUrlValue().Add("page", "1").Add("per_page", "2"),
		NewJsonHeader(),
	).DecodeJSON(&result)
	if err != nil {
		return
	}
	fmt.Printf("%d\n", len(result))

	// OutPut:
	// 2
}
