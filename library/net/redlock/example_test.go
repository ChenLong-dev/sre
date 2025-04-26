package redlock

import (
	"context"
	"sync"
	"time"

	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/log"
)

func getRedisPool() *redis.Pool {
	redisConfig := redis.Config{
		PoolConfig: &redis.PoolConfig{
			Active:      10,
			Idle:        10,
			IdleTimeout: ctime.Duration(time.Hour * 2),
			CheckTime:   ctime.Duration(time.Second * 10),
			Wait:        true,
		},
		Proto: "tcp",
		DB:    1,
		Endpoint: &redis.EndpointConfig{
			Address: "localhost",
			Port:    6379,
		},
		Auth:            "123456",
		MaxConnLifetime: ctime.Duration(time.Second * 1),
	}
	redisPool := redis.NewPool(&redisConfig)

	return redisPool
}

func ExampleNew() {
	c := &Config{
		ExpiryTime:  ctime.Duration(10 * time.Second),
		Tries:       50,
		RetryDelay:  ctime.Duration(50 * time.Millisecond),
		DriftFactor: 0.01,
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] [duration: %D] %S  Mutex: %N , Command: %n , State: %s",
		},
	}
	r := New(c, getRedisPool())
	m := r.NewMutex("library_redlock_example")

	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		defer wg.Done()

		// 加锁
		err := m.Lock(context.Background())
		if err != nil {
			log.Error("%s", err)
			return
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Second)

		// 解锁
		_ = m.Unlock(context.Background())
	}()

	wg.Wait()
}
