package etcd

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func TestDB_LoadPrefixData(t *testing.T) {
	t.Skipf("should have etcd environment\n")

	c := NewClient(&Config{
		Endpoints: []*EndpointConfig{
			{
				Address: "localhost",
				Port:    2379,
			},
		},
		UserName:     "root",
		Password:     "123456",
		DataValueOut: true,
		DialTimeout:  ctime.Duration(time.Second * 10),
		Preload:      nil,
		Tls:          nil,
		Config: &render.Config{
			Stdout:        true,
			StdoutPattern: "[%T] [%t] [%U] %S  Prefix: %P , Key: %K , Value: %V , Extra: %e",
		},
	})

	t.Run("normal", func(t *testing.T) {
		c.kvClient.Put(context.Background(), "/root/payment/staging/a", "test a")

		m, err := c.LoadPrefixData(context.Background(), "/root/payment/staging/", []string{}, false)
		assert.Nil(t, err)

		v, err := m.Get("a")
		assert.Nil(t, err)
		assert.Equal(t, "test a", v)
	})

	t.Run("filter", func(t *testing.T) {
		c.kvClient.Put(context.Background(), "/root/payment/staging/c/a", "test ca")
		c.kvClient.Put(context.Background(), "/root/payment/staging/c/b", "test cb")

		m, err := c.LoadPrefixData(context.Background(), "/root/payment/staging/", []string{"test"}, false)
		assert.Nil(t, err)

		_, err = m.Get("c")
		assert.NotNil(t, err)
	})

	t.Run("reset", func(t *testing.T) {
		_, err := c.LoadPrefixData(context.Background(), "/root/payment/staging/", []string{}, true)
		assert.Nil(t, err)
		time.Sleep(time.Second * 3)
		preNum := runtime.NumGoroutine()

		_, err = c.ReloadPrefixData(context.Background(), "/root/payment/staging/", []string{}, true)
		assert.Nil(t, err)
		time.Sleep(time.Second * 3)
		assert.Equal(t, preNum, runtime.NumGoroutine())
	})

	t.Run("close", func(t *testing.T) {
		time.Sleep(time.Second * 3)
		preNum := runtime.NumGoroutine()

		time.Sleep(time.Second * 3)
		_, err := c.LoadPrefixData(context.Background(), "/root/payment/staging/", []string{}, true)
		assert.Nil(t, err)

		time.Sleep(time.Second * 3)
		c.Close()

		time.Sleep(time.Second * 3)
		assert.Equal(t, preNum, runtime.NumGoroutine())
	})
}
