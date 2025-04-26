package etcd

import (
	"fmt"
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func ExampleNewClient() {
	c := NewClient(&Config{
		Endpoints: []*EndpointConfig{
			{
				Address: "localhost",
				Port:    2379,
			},
		},
		UserName:     "etcd",
		Password:     "123456",
		DataValueOut: true,
		DialTimeout:  ctime.Duration(time.Second * 10),
		Preload: []*PreloadPrefixConfig{{
			EnableWatch: true,
			Prefix:      "root/payment/staging/",
			ValueFilter: []string{"etcdv3_dir"},
		}},
		Tls: &TlsConfig{
			Enable:            true,
			CertFilePath:      "./cert.pem",
			KeyFilePath:       "./key.pem",
			TrustedCAFilePath: "./ca.pem",
		},
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] %S  Prefix: %P , Key: %K , Value: %V , Extra: %e",
		},
	})

	data, err := c.GetPrefix("root/payment/staging/")
	if err != nil {
		return
	}

	value, err := data.Get("single_rich_text")
	if err != nil {
		return
	}

	fmt.Println(value)
}
