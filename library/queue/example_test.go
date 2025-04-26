package queue

import (
	"github.com/streadway/amqp"

	"context"
	"fmt"
	"strings"
	"time"
)

func ExampleSession_Push() {
	q := New(&globalConfig)
	s, err := q.NewSession("unit")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()
	msg := &amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(time.Now().Format(time.RFC3339Nano)),
	}
	err = s.Push(ctx, msg, PushOption{})
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ExampleSession_UnsafePush() {
	globalConfig.Session["UnsafePush"] = &SessionConfig{
		QueueName:          "unittest",
		ExchangeName:       "boot",
		RoutingKey:         "",
		Durable:            true,
		DisableConfirmMode: true,
	}
	q := New(&globalConfig)
	s, err := q.NewSession("UnsafePush")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()
	msg := &amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(time.Now().Format(time.RFC3339Nano)),
	}
	err = s.UnsafePush(ctx, msg, PushOption{})
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ExampleSession_SingleConsumeStream() {
	q := New(&globalConfig)
	s, err := q.NewSession("unit")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()
	err = s.SingleConsumeStream(ctx, func(d amqp.Delivery) error {
		b := string(d.Body)
		fmt.Println(b)
		return nil
	}, ConsumeOption{})
	if err != nil {
		fmt.Println(err)
		return
	}
}

func ExampleSession_NoAutoAckConsumeStream() {
	q := New(&globalConfig)
	s, err := q.NewSession("unit")
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()
	err = s.NoAutoAckConsumeStream(ctx, func(d amqp.Delivery) error {
		b := string(d.Body)
		fmt.Println(b)
		if strings.Contains(b, "2021") {
			return d.Ack(false)
		} else {
			return d.Reject(true)
		}
	}, ConsumeOption{
		ConsumerName:  "example",
		PrefetchCount: 20,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
}
