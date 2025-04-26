package queue

import (
	"gitlab.shanhai.int/sre/library/internal/test"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"context"
	"sync"
	"testing"
	"time"
)

var testSessionPush = []test.SyncUnitCase{
	{
		Name: "invalid mode",
		Func: func(t *testing.T) {
			globalConfig.Session["SessionPush"] = &SessionConfig{
				QueueName:          "SessionPush",
				ExchangeName:       "boot",
				RoutingKey:         "",
				Durable:            true,
				DisableConfirmMode: true,
			}
			q := New(&globalConfig)
			s, err := q.NewSession("SessionPush")
			assert.NoError(t, err)
			defer s.Close()

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.EqualError(t, err, "confirm mode is disable, please use `UnsafePush` method")
		},
	},
	{
		Name: "disconnect",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			// 断网
			_ = toxiProxy.Disable()
			time.Sleep(time.Second)
			defer func() {
				_ = toxiProxy.Enable()
				time.Sleep(time.Second)
			}()

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.EqualError(t, err, ErrNotConnected.Error())
		},
	},
	{
		Name: "sudden disconnected",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			// 增加延迟
			tc, err := toxiProxy.AddToxic("limit resend",
				"latency", "",
				1, toxiproxy.Attributes{
					"latency": 5000,
				})
			assert.NoError(t, err)
			defer func() {
				_ = toxiProxy.RemoveToxic(tc.Name)
				_ = toxiProxy.Enable()
			}()
			go func() {
				time.Sleep(time.Second)
				_ = toxiProxy.Disable()
			}()

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.EqualError(t, err, "current confirm channel is closed")
		},
	},
	{
		Name: "multi",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			wg := new(sync.WaitGroup)
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					data := time.Now().Format(time.RFC3339Nano)
					err = s.Push(context.Background(), &amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(data),
					}, PushOption{})
					assert.NoError(t, err)
				}()
			}
			wg.Wait()

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)
		},
	},
	{
		Name: "invalid msg",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
				Headers: amqp.Table{
					"some": uint(1),
				},
			}, PushOption{})
			assert.EqualError(t, err, "table field \"some\" value uint not supported")
		},
	},
	{
		Name: "conn close",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			tc, err := toxiProxy.AddToxic("limit resend",
				"timeout", "",
				1, toxiproxy.Attributes{
					"timeout": 0,
				})
			assert.NoError(t, err)
			go func() {
				time.Sleep(time.Second)
				_ = toxiProxy.RemoveToxic(tc.Name)
			}()
			wg := new(sync.WaitGroup)
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					data := time.Now().Format(time.RFC3339Nano)
					err = s.Push(context.Background(), &amqp.Publishing{
						ContentType: "text/plain",
						Body:        []byte(data),
					}, PushOption{})
					assert.EqualError(t, err, "current confirm channel is closed")
				}()
			}
			wg.Wait()
		},
	},
}

var testSessionUnsafePush = []test.SyncUnitCase{
	{
		Name: "normal",
		Func: func(t *testing.T) {
			globalConfig.Session["SessionUnsafePush"] = &SessionConfig{
				QueueName:          "SessionUnsafePush",
				ExchangeName:       "boot",
				RoutingKey:         "",
				Durable:            true,
				DisableConfirmMode: true,
			}
			q := New(&globalConfig)
			s, err := q.NewSession("SessionUnsafePush")
			assert.NoError(t, err)
			defer s.Close()

			data := time.Now().Format(time.RFC3339Nano)
			err = s.UnsafePush(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			isConsume := make(chan bool)
			go func() {
				_ = s.SingleConsumeStream(context.Background(), func(d amqp.Delivery) error {
					b := string(d.Body)
					if b == data {
						isConsume <- true
					}
					return nil
				}, ConsumeOption{})
			}()

			select {
			case res := <-isConsume:
				assert.Equal(t, true, res)
			case tm := <-time.Tick(time.Second * 15):
				assert.Nil(t, tm)
			}
		},
	},
	{
		Name: "invalid mode",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			data := time.Now().Format(time.RFC3339Nano)
			err = s.UnsafePush(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.EqualError(t, err, "confirm mode is enable, please use `Push` method")
		},
	},
}
