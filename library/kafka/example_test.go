package kafka

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/Shopify/sarama"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/goroutine"
)

// 异步生产者示例
func ExampleAsyncProducer() {
	// 初始化goroutine
	goroutine.Init(nil)

	// 新建客户端
	client := NewClient(&Config{
		Endpoints: []*EndpointConfig{
			{
				Address: "localhost",
				Port:    9092,
			},
		},
		AppID:   "example",
		Version: "2.2.0",
		Producer: &ProducerConfig{
			ReturnSuccess: true,
			ReturnError:   true,
		},
		Config: &render.Config{
			Stdout: true,
		},
	})
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("client close error: %s\n", err)
		}
	}()

	// 新建异步生产者
	producer, err := client.NewAsyncProducer()
	if err != nil {
		fmt.Printf("new async producer error: %s\n", err)
		return
	}

	wg := goroutine.New("async")
	// 发送成功监听
	wg.Go(context.Background(), "success", func(ctx context.Context) error {
		for msg := range producer.Successes() {
			fmt.Printf("push message success: topic=%s partition=%d offset=%d\n",
				msg.Topic, msg.Partition, msg.Offset)
		}
		return nil
	})
	// 发送失败监听
	wg.Go(context.Background(), "error", func(ctx context.Context) error {
		for err := range producer.Errors() {
			fmt.Printf("push message error: topic=%s partition=%d offset=%d error=%s\n",
				err.Msg.Topic, err.Msg.Partition, err.Msg.Offset, err.Err)
		}
		return nil
	})
	// 发送消息
	wg.Go(context.Background(), "push", func(ctx context.Context) error {
		for i := 0; i < 10; i++ {
			message := &sarama.ProducerMessage{
				Topic: "example",
				Key:   sarama.StringEncoder(strconv.Itoa(i)),
				Value: sarama.StringEncoder(time.Now().String()),
			}
			producer.Input() <- message
		}
		producer.AsyncClose()

		return nil
	})

	wg.Wait()
}

// 同步生产者示例
func ExampleSyncProducer() {
	// 新建客户端
	client := NewClient(&Config{
		Endpoints: []*EndpointConfig{
			{
				Address: "localhost",
				Port:    9092,
			},
		},
		AppID:   "example",
		Version: "2.2.0",
		Producer: &ProducerConfig{
			ReturnSuccess: true,
			ReturnError:   true,
		},
		Config: &render.Config{
			Stdout: true,
		},
	})
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("client close error: %s\n", err)
		}
	}()

	// 新建同步生产者
	producer, err := client.NewSyncProducer()
	if err != nil {
		fmt.Printf("new sync producer error: %s\n", err)
		return
	}
	defer func() {
		if err := producer.Close(); err != nil {
			fmt.Printf("producer close error: %s\n", err)
		}
	}()

	// 发送消息
	for i := 0; i < 10; i++ {
		message := &sarama.ProducerMessage{
			Topic: "example",
			Value: sarama.StringEncoder(fmt.Sprintf("message index is %d", i)),
		}

		partition, offset, err := producer.SendMessage(message)
		if err != nil {
			fmt.Printf("push message error: topic=%s partition=%d offset=%d error=%s\n",
				message.Topic, partition, offset, err)
			continue
		}

		fmt.Printf("push message success: topic=%s partition=%d offset=%d\n",
			message.Topic, message.Partition, message.Offset)
	}
}

// 单个消费者示例
func ExampleConsumer() {
	// 初始化goroutine
	goroutine.Init(nil)

	// 新建客户端
	client := NewClient(&Config{
		Endpoints: []*EndpointConfig{
			{
				Address: "localhost",
				Port:    9092,
			},
		},
		AppID:   "example",
		Version: "2.2.0",
		Consumer: &ConsumerConfig{
			MaxProcessingTime: ctime.Duration(time.Millisecond * 500),
			ReturnError:       true,
		},
		Config: &render.Config{
			Stdout: true,
		},
	})
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("client close error: %s\n", err)
		}
	}()

	// 新建单个消费者
	consumer, err := client.NewConsumer()
	if err != nil {
		fmt.Printf("new consumer error: %s\n", err)
		return
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			fmt.Printf("consumer close error: %s\n", err)
		}
	}()

	// 新建分区消费者
	partitionConsumer, err := consumer.ConsumePartition("example", 0, sarama.OffsetNewest)
	if err != nil {
		fmt.Printf("partition consumer error: %s\n", err)
		return
	}

	wg := goroutine.New("consumer")
	// 消费失败
	wg.Go(context.Background(), "error", func(ctx context.Context) error {
		for err := range partitionConsumer.Errors() {
			fmt.Printf("consume error: %s\n", err)
		}
		return nil
	})
	// 消费消息
	wg.Go(context.Background(), "message", func(ctx context.Context) error {
		for msg := range partitionConsumer.Messages() {
			fmt.Printf("consumer success: key=%s value=%s\n", string(msg.Key), string(msg.Value))
		}
		return nil
	})

	// 监听关闭信号量
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for {
		select {
		case <-signals:
			if err := partitionConsumer.Close(); err != nil {
				fmt.Printf("partition consumer close error: %s\n", err)
			}
			wg.Wait()
			return
		}
	}
}

// 异步消费组处理器
type ExampleAsyncConsumerGroupHandler struct{}

// 处理前初始化
func (ExampleAsyncConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// 处理后清除
func (ExampleAsyncConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// 消费消息
func (h ExampleAsyncConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Message topic:%q partition:%d offset:%d message=%s\n",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
		// 消费成功
		session.MarkMessage(msg, "")
	}
	return nil
}

// 消费组示例
func ExampleConsumerGroup() {
	// 初始化goroutine
	goroutine.Init(nil)

	// 新建客户端
	client := NewClient(&Config{
		Endpoints: []*EndpointConfig{
			{
				Address: "localhost",
				Port:    9092,
			},
		},
		AppID:   "example",
		Version: "2.2.0",
		Consumer: &ConsumerConfig{
			MaxProcessingTime: ctime.Duration(time.Millisecond * 500),
			ReturnError:       true,
			InitialOffset:     sarama.OffsetOldest,
		},
		Config: &render.Config{
			Stdout: true,
		},
	})
	defer func() {
		if err := client.Close(); err != nil {
			fmt.Printf("client close error: %s\n", err)
		}
	}()

	// 新建消费组
	group, err := client.NewConsumerGroup("consumer-group")
	if err != nil {
		fmt.Printf("new consumer group error: %s\n", err)
		return
	}

	wg := goroutine.New("consumer-group")
	// 消费错误
	wg.Go(context.Background(), "error", func(ctx context.Context) error {
		for err := range group.Errors() {
			fmt.Printf("receive message error: %s\n", err)
		}
		return nil
	})
	// 消费消息
	wg.Go(context.Background(), "consume", func(ctx context.Context) error {
		topics := []string{"example"}
		handler := ExampleAsyncConsumerGroupHandler{}

		err := group.Consume(context.Background(), topics, handler)
		if err != nil {
			fmt.Printf("group consumer error: %s\n", err)
		}

		return nil
	})

	// 监听关闭信号量
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	for {
		select {
		case <-signals:
			if err := group.Close(); err != nil {
				fmt.Printf("consumer group close error: %s\n", err)
			}
			wg.Wait()
			return
		}
	}
}
