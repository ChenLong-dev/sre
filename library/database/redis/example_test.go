package redis

import (
	"context"
	"fmt"
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func ExampleNewPool() {
	c := Config{
		PoolConfig: &PoolConfig{
			Active:      10,
			Idle:        10,
			IdleTimeout: ctime.Duration(time.Hour * 2),
			CheckTime:   ctime.Duration(time.Second * 10),
			Wait:        true,
		},
		Proto: "tcp",
		DB:    1,
		Endpoint: &EndpointConfig{
			Address: "r-xxxxxxxxxxxxxxxx.redis.rds.aliyuncs.com",
			Port:    6379,
		},
		Auth:            "123456",
		MaxConnLifetime: ctime.Duration(time.Hour * 4),
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [duration: %D] %S  Func: %N  CMD: %n  Args: %J{a}  Reply: %J{r}",
		},
	}

	p := NewPool(&c)

	err := p.WrapDo(func(con *Conn) error {
		reply, err := con.Do(context.Background(), "sismember", "collection", "member")
		if err != nil {
			return err
		}

		fmt.Printf("%v\n", reply)
		return nil
	})
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	con := p.Get()
	defer con.Close()
	reply, err := con.Do(context.Background(), "sismember", "collection", "member")
	if err != nil {
		return
	}
	fmt.Printf("%v\n", reply)
}
