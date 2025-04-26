package agollo

import (
	"fmt"

	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 实时更新配置
func ExampleNewClient_daemon() {
	client := NewClient(&Config{
		AppID:             "your_appID",
		Cluster:           "dev",
		ServerHost:        "http://localhost:8080",
		PreloadNamespaces: []string{"application", "namespace1", "namespace2"},
		// 设置false实时获取最新配置
		NotDaemon: false,
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] %S , appID: %A , cluster: %C , serveHost: %I , namespace: %N , Extra: %E , changes: %c",
		},
	})
	defer client.Close()

	// 方式一: 先获取namespace下所有配置，再根据key获取最新value
	res, err := client.GetNamespace("application")
	if err != nil {
		panic(err)
	}
	res1, err := res.Load("key1")
	res2, err := res.Load("key2")
	if err != nil {
		panic(err)
	}
	fmt.Println(res1.(string))
	fmt.Println(res2.(string))

	// 方式二:通过namespace和key获取最新配置
	res3, err := client.GetNewValue("namespace", "key")
	if err != nil {
		panic(err)
	}

	fmt.Println(res3.(string))
}

// watch配置
func ExampleNew_watch() {
	client := NewClient(&Config{
		AppID:      "your_appID",
		Cluster:    "dev",
		ServerHost: "http://localhost:8080",
		// 设置true需手动watch
		NotDaemon:         true,
		PreloadNamespaces: []string{"application", "namespace2", "namespace3"},
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] %S , appID: %A , cluster: %C , serveHost: %I , namespace: %N , Extra: %E , changes: %c",
		},
	})

	defer client.Close()

	res := client.Get("namespace1")
	fmt.Printf("%v\n", res)

	resp, err := client.Watch([]string{"application"})
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case r1 := <-resp:
				if r1.Type == EventUpdate {
					if r1.Namespace == "application" {
						conf := client.Get(r1.Namespace)
						fmt.Println(conf)
					}
				} else if r1.Type == EventError {
					fmt.Printf("Watch err: %s\n", r1.ErrInfo.Err.Error())
				}
			case <-client.Done():
				return
			}
		}
	}()
}
