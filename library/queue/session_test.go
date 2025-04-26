package queue

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	"gitlab.shanhai.int/sre/library/internal/test"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"context"
	"sync"
	"testing"
	"time"
)

var testSessionCloseStream = []test.SyncUnitCase{
	{
		Name: "normal",
		Func: func(t *testing.T) {
			globalConfig.Session["CloseStream"] = &SessionConfig{
				QueueName:    "unittest",
				ExchangeName: "boot",
				RoutingKey:   "",
				Durable:      true,
			}
			q := New(&globalConfig)
			s, err := q.NewSession("CloseStream")
			assert.NoError(t, err)
			defer s.Close()

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			isConsume := make(chan bool, 1)
			wg := new(sync.WaitGroup)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = s.SingleConsumeStream(context.Background(), func(d amqp.Delivery) error {
					b := string(d.Body)
					if b == data {
						isConsume <- true
					}
					return nil
				}, ConsumeOption{
					AutoAck: true,
				})
				assert.NoError(t, err)
			}()

			select {
			case res := <-isConsume:
				assert.Equal(t, true, res)

				err = s.ChannelDo(func(ch *amqp.Channel) error {
					bq, e := ch.QueueInspect(s.GetConfig().QueueName)
					if e != nil {
						return e
					}
					assert.Greater(t, bq.Consumers, 0)

					s.CloseStream()

					after := inspectQueueByNewSession(t, q, "CloseStream")
					assert.Less(t, after.Consumers, bq.Consumers)

					wg.Wait()
					return nil
				})
				assert.NoError(t, err)
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

			data := time.Now().Format(time.RFC3339Nano)
			err = s.Push(context.Background(), &amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(data),
			}, PushOption{})
			assert.NoError(t, err)

			isConsume := make(chan bool, 1)
			wg := new(sync.WaitGroup)
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = s.SingleConsumeStream(context.Background(), func(d amqp.Delivery) error {
					b := string(d.Body)
					if b == data {
						isConsume <- true
					}
					return nil
				}, ConsumeOption{
					AutoAck: true,
				})
				assert.Error(t, err)
			}()

			select {
			case res := <-isConsume:
				assert.Equal(t, true, res)

				// 断网
				_ = toxiProxy.Disable()
				defer func() {
					_ = toxiProxy.Enable()
				}()

				s.CloseStream()

				wg.Wait()
			case tm := <-time.Tick(time.Second * 15):
				assert.Nil(t, tm)
			}
		},
	},
}

var testSessionCloseChannel = []test.SyncUnitCase{
	{
		Name: "normal",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)

			err = s.CloseChannel()
			assert.NoError(t, err)
		},
	},
	{
		Name: "not ready",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			// 断网
			_ = toxiProxy.Disable()
			defer func() {
				_ = toxiProxy.Enable()
			}()

			e := s.CloseChannel()
			assert.Error(t, e, ErrCloseNotReady.Error())
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
			err = s.CloseChannel()
			assert.Error(t, err)
		},
	},
}

var testSessionConnectionDo = []test.SyncUnitCase{
	{
		Name: "normal",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			err = s.ConnectionDo(func(conn *amqp.Connection) error {
				res := conn.IsClosed()
				assert.Equal(t, false, res)
				return nil
			})
			assert.NoError(t, err)
		},
	},
	{
		Name: "not ready",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			// 断网
			_ = toxiProxy.Disable()
			defer func() {
				_ = toxiProxy.Enable()
			}()

			e := s.ConnectionDo(func(conn *amqp.Connection) error {
				return nil
			})
			assert.Error(t, e, ErrCloseNotReady.Error())
		},
	},
}

var testSessionChannelDo = []test.SyncUnitCase{
	{
		Name: "not ready",
		Func: func(t *testing.T) {
			q := New(&globalConfig)
			s, err := q.NewSession("unit")
			assert.NoError(t, err)
			defer s.Close()

			// 断网
			_ = toxiProxy.Disable()
			defer func() {
				_ = toxiProxy.Enable()
			}()

			e := s.ChannelDo(func(ch *amqp.Channel) error {
				return nil
			})
			assert.Error(t, e, ErrCloseNotReady.Error())
		},
	},
}

var testSessionClose = []test.SyncUnitCase{
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
			err = s.Close()
			assert.Error(t, err)
		},
	},
	{
		Name: "invalid declare",
		Func: func(t *testing.T) {
			session := Session{
				id: "invalid declare",
				config: &SessionConfig{
					QueueName:    "unittest",
					ExchangeName: "boot",
					RoutingKey:   "",
					Durable:      true,
					Args: amqp.Table{
						"some": uint(1),
					},
					ReInitDelay: ctime.Duration(time.Second * 15),
				},
				manager:            NewHookManager(globalConfig.Config),
				done:               make(chan bool),
				notifyStreamCloses: make([]chan bool, 0),
			}
			go session.handleReconnect(getConnectAddr(&globalConfig))
			time.Sleep(time.Second * 5)
			err := session.Close()
			assert.Error(t, err, ErrCloseNotReady)
		},
	},
}
