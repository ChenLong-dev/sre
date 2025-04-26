package queue

import (
	"gitlab.shanhai.int/sre/library/internal/test"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"context"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// 通过新建session检查当前队列
// 为避免旧session中拒绝消息，需新建session获取
func inspectQueueByNewSession(t *testing.T, q *Queue, name string) amqp.Queue {
	ns, err := q.NewSession(name)
	assert.NoError(t, err)
	assert.NotNil(t, ns)
	assert.Equal(t, true, ns.IsReady())
	defer ns.Close()

	var aq amqp.Queue
	err = ns.ChannelDo(func(ch *amqp.Channel) error {
		aq, err = ch.QueueInspect(ns.GetConfig().QueueName)
		if err != nil {
			return err
		}
		return nil
	})
	assert.NoError(t, err)

	return aq
}

var testSessionSingleConsumeStream = []test.SyncUnitCase{
	{
		Name: "normal",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			var beforeCount int
			err = s.ChannelDo(func(ch *amqp.Channel) error {
				bq, e := ch.QueueInspect(s.GetConfig().QueueName)
				if e != nil {
					return e
				}
				beforeCount = bq.Messages
				return nil
			})
			assert.NoError(t, err)

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
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
				err = s.Close()
				assert.NoError(t, err)

				after := inspectQueueByNewSession(t, q, "unit")
				if beforeCount != 0 {
					assert.Less(t, after.Messages, beforeCount)
				}
			case tm := <-time.Tick(time.Second * 15):
				assert.Nil(t, tm)
			}
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

			err = s.SingleConsumeStream(context.Background(), func(d amqp.Delivery) error {
				return nil
			}, ConsumeOption{})
			assert.EqualError(t, err, ErrNotConnected.Error())
		},
	},
	{
		Name: "invalid exchange",
		Func: func(t *testing.T) {
			globalConfig.Session["SessionSingleConsumeStream"] = &SessionConfig{
				QueueName:    "SessionSingleConsumeStream",
				ExchangeName: "invalid",
				RoutingKey:   "",
				Durable:      true,
			}
			q := New(&globalConfig)
			s, err := q.NewSession("SessionSingleConsumeStream")
			assert.NoError(t, err)
			defer s.Close()

			err = s.SingleConsumeStream(context.Background(), func(d amqp.Delivery) error {
				return nil
			}, ConsumeOption{})
			assert.Error(t, err)
		},
	},
	{
		Name: "sudden close",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			// 增加延迟
			tc, err := toxiProxy.AddToxic("invalid queue",
				"latency", "",
				1, toxiproxy.Attributes{
					"latency": 2000,
				})
			assert.NoError(t, err)
			defer func() {
				_ = toxiProxy.RemoveToxic(tc.Name)
			}()
			go func() {
				time.Sleep(time.Second)
				err = s.Close()
				assert.NoError(t, err)
			}()

			err = s.SingleConsumeStream(context.Background(), func(d amqp.Delivery) error {
				return nil
			}, ConsumeOption{})
			assert.NotNil(t, err)
			assert.NotEqual(t, err, ErrNotConnected)
		},
	},
	{
		Name: "reconnect after init consumer",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			var beforeCount int
			err = s.ChannelDo(func(ch *amqp.Channel) error {
				bq, e := ch.QueueInspect(s.GetConfig().QueueName)
				if e != nil {
					return e
				}
				beforeCount = bq.Messages
				return nil
			})
			assert.NoError(t, err)

			data := time.Now().Format(time.RFC3339Nano)

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
			// 创建消费者后等待消费者连接成功
			time.Sleep(time.Second * 3)

			// 断网
			_ = toxiProxy.Disable()
			time.Sleep(time.Second)

			// 恢复
			_ = toxiProxy.Enable()
			time.Sleep(time.Second * 3)

			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			select {
			case res := <-isConsume:
				assert.Equal(t, true, res)
				err = s.Close()
				assert.NoError(t, err)

				after := inspectQueueByNewSession(t, q, "unit")
				if beforeCount != 0 {
					assert.Less(t, after.Messages, beforeCount)
				}
			case tm := <-time.Tick(time.Second * 15):
				assert.Nil(t, tm)
			}
		},
	},
	{
		Name: "reconnect before init consumer",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			data := time.Now().Format(time.RFC3339Nano)

			wg := new(sync.WaitGroup)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = s.SingleConsumeStream(context.Background(), func(d amqp.Delivery) error {
					return nil
				}, ConsumeOption{})
				assert.Error(t, err)
			}()

			// 创建消费者后，直接断网
			// 不等待消费者连接成功
			_ = toxiProxy.Disable()
			time.Sleep(time.Second)

			// 恢复
			_ = toxiProxy.Enable()
			time.Sleep(time.Second * 3)

			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			wg.Wait()
		},
	},
	{
		Name: "long consumer tag",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			baseInfix := os.Args[0]
			for i := 0; i < 128; i++ {
				baseInfix += strconv.Itoa(i)
			}
			os.Args[0] = baseInfix

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
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
}

var testSessionNoAutoAckConsumeStream = []test.SyncUnitCase{
	{
		Name: "invalid mode",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			err = s.NoAutoAckConsumeStream(context.Background(), func(d amqp.Delivery) error {
				return nil
			}, ConsumeOption{
				AutoAck: true,
			})
			assert.Error(t, err)
		},
	},
	{
		Name: "ack",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			var beforeCount int
			err = s.ChannelDo(func(ch *amqp.Channel) error {
				bq, e := ch.QueueInspect(s.GetConfig().QueueName)
				if e != nil {
					return e
				}
				beforeCount = bq.Messages
				return nil
			})
			assert.NoError(t, err)

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			isConsume := make(chan bool)
			go func() {
				_ = s.NoAutoAckConsumeStream(context.Background(), func(d amqp.Delivery) error {
					b := string(d.Body)
					if b == data {
						e := d.Ack(false)
						assert.NoError(t, e)
						isConsume <- true
					}
					return nil
				}, ConsumeOption{})
			}()

			select {
			case res := <-isConsume:
				assert.Equal(t, true, res)
				err = s.Close()
				assert.NoError(t, err)

				after := inspectQueueByNewSession(t, q, "unit")
				assert.Equal(t, after.Messages, beforeCount)
			case tm := <-time.Tick(time.Second * 15):
				assert.Nil(t, tm)
			}
		},
	},
	{
		Name: "reject",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			var beforeCount int
			err = s.ChannelDo(func(ch *amqp.Channel) error {
				bq, e := ch.QueueInspect(s.GetConfig().QueueName)
				if e != nil {
					return e
				}
				beforeCount = bq.Messages
				return nil
			})
			assert.NoError(t, err)

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			isConsume := make(chan bool)
			go func() {
				_ = s.NoAutoAckConsumeStream(context.Background(), func(d amqp.Delivery) error {
					b := string(d.Body)
					if b == data {
						e := d.Reject(true)
						assert.NoError(t, e)
						isConsume <- true
					}
					return nil
				}, ConsumeOption{})
			}()

			select {
			case res := <-isConsume:
				assert.Equal(t, true, res)
				err = s.Close()
				assert.NoError(t, err)

				after := inspectQueueByNewSession(t, q, "unit")
				assert.Greater(t, after.Messages, beforeCount)
			case tm := <-time.Tick(time.Second * 15):
				assert.Nil(t, tm)
			}
		},
	},
	{
		Name: "un ack",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			var beforeCount int
			err = s.ChannelDo(func(ch *amqp.Channel) error {
				bq, e := ch.QueueInspect(s.GetConfig().QueueName)
				if e != nil {
					return e
				}
				beforeCount = bq.Messages
				return nil
			})
			assert.NoError(t, err)

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			isConsume := make(chan bool)
			go func() {
				_ = s.NoAutoAckConsumeStream(context.Background(), func(d amqp.Delivery) error {
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
				err = s.Close()
				assert.NoError(t, err)

				after := inspectQueueByNewSession(t, q, "unit")
				assert.Greater(t, after.Messages, beforeCount)
			case tm := <-time.Tick(time.Second * 15):
				assert.Nil(t, tm)
			}
		},
	},

	{
		Name: "consume error",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			err = s.NoAutoAckConsumeStream(context.Background(), func(d amqp.Delivery) error {
				err = s.Close()
				assert.NoError(t, err)

				err = d.Ack(false)
				return err
			}, ConsumeOption{})
			assert.Error(t, err)
		},
	},
}

var testSessionStream = []test.SyncUnitCase{
	{
		Name: "normal",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			var beforeCount int
			err = s.ChannelDo(func(ch *amqp.Channel) error {
				sessionConfig := s.GetConfig()
				bq, e := ch.QueueInspect(sessionConfig.QueueName)
				if e != nil {
					return e
				}
				beforeCount = bq.Messages

				e = ch.QueueBind(
					sessionConfig.QueueName,
					sessionConfig.RoutingKey,
					sessionConfig.ExchangeName,
					false,
					nil,
				)
				if e != nil {
					return e
				}

				e = ch.Qos(1, 0, false)
				if e != nil {
					return e
				}
				return nil
			})
			assert.NoError(t, err)

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			isConsume := make(chan bool)
			go func() {
				_ = s.Stream(context.Background(), func(d amqp.Delivery) error {
					b := string(d.Body)
					if b == data {
						isConsume <- true
					}
					return nil
				}, ConsumeOption{
					AutoAck: true,
				})
			}()

			select {
			case res := <-isConsume:
				assert.Equal(t, true, res)
				err = s.Close()
				assert.NoError(t, err)

				after := inspectQueueByNewSession(t, q, "unit")
				if beforeCount != 0 {
					assert.Less(t, after.Messages, beforeCount)
				}
			case tm := <-time.Tick(time.Second * 15):
				assert.Nil(t, tm)
			}
		},
	},
}
