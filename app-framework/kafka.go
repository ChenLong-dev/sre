package framework

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/kafka"
	"gitlab.shanhai.int/sre/library/net/sentry"
)

// 消息队列监听服务
type KafkaServer struct {
	// 基础配置文件
	config *Config
	// 当前队列配置文件
	queueConfig *kafka.Config
	// 消费组id
	groupID string
	// 消费主题数组
	topics []string

	// Kafka客户端
	client *kafka.Client
	// 消费组
	group sarama.ConsumerGroup
	// 消费组内用于协程的WG
	groupWaitGroup *goroutine.ErrGroup

	// 消费前初始化函数
	ConsumerSetup func(session sarama.ConsumerGroupSession) error
	// 消费函数
	ConsumerConsume func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
	// 消费后清理函数
	ConsumerCleanup func(session sarama.ConsumerGroupSession) error
	// 消费错误函数
	ConsumerError func(err error)
}

// 处理前初始化
// 实现 `sarama.ConsumerGroupHandler` 接口
func (svr *KafkaServer) Setup(session sarama.ConsumerGroupSession) error {
	return svr.ConsumerSetup(session)
}

// 处理后清理
// 实现 `sarama.ConsumerGroupHandler` 接口
func (svr *KafkaServer) Cleanup(session sarama.ConsumerGroupSession) error {
	return svr.ConsumerCleanup(session)
}

// 消费消息
// 实现 `sarama.ConsumerGroupHandler` 接口
func (svr *KafkaServer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	return svr.ConsumerConsume(session, claim)
}

// 设置消费组id
func (svr *KafkaServer) SetGroupID(id string) {
	svr.groupID = id
}

// 设置消费主题
func (svr *KafkaServer) SetTopics(topics []string) {
	svr.topics = topics
}

// 获取消费组id
func (svr *KafkaServer) GetGroupID() (id string) {
	return svr.groupID
}

// 获取消费主题
func (svr *KafkaServer) GetTopics() (topics []string) {
	return svr.topics
}

// 实现ServerInterface
func (svr *KafkaServer) ShutDown(ctx context.Context) (err error) {
	svr.groupWaitGroup.CallCancel()

	groupError := svr.group.Close()
	if groupError != nil {
		err = groupError
	}

	clientError := svr.client.Close()
	if clientError != nil {
		err = clientError
	}

	return err
}

// 实现ServerInterface
func (svr *KafkaServer) Start(c *Config, svc ServiceInterface) {
	svr.config = c
	if svr.groupID == "" {
		panic("kafka server groupID haven't assign")
	}
	if len(svr.topics) == 0 {
		panic("kafka server topics is empty")
	}
	if svr.ConsumerConsume == nil {
		panic("kafka server consumer consume function is nil")
	}

	if svr.ConsumerSetup == nil {
		svr.ConsumerSetup = func(session sarama.ConsumerGroupSession) error {
			return nil
		}
	}
	if svr.ConsumerCleanup == nil {
		svr.ConsumerCleanup = func(session sarama.ConsumerGroupSession) error {
			return nil
		}
	}

	if queueConfig, ok := svr.config.Kafka[svr.groupID]; ok {
		svr.queueConfig = queueConfig
	} else {
		panic(fmt.Sprintf("kafka server %s haven't config", svr.groupID))
	}
	if svr.queueConfig.Consumer.ReturnError && svr.ConsumerError == nil {
		panic("kafka server consumer error function is nil")
	}

	svr.client = kafka.NewClient(svr.queueConfig)

	group, err := svr.client.NewConsumerGroup(svr.groupID)
	if err != nil {
		panic(fmt.Sprintf("kafka server %s create consumer group error: %s\n", svr.groupID, err))
	}
	svr.group = group

	goroutineCtx := context.Background()

	svr.groupWaitGroup = goroutine.New(svr.groupID, goroutine.SetCancelMode(goroutineCtx))

	svc.StartServer(svr.groupID, func(ctx context.Context) error {
		if svr.queueConfig.Consumer.ReturnError {
			svr.groupWaitGroup.Go(goroutineCtx, "error", func(ctx context.Context) error {
				for err := range svr.group.Errors() {
					svr.ConsumerError(err)
					sentry.CaptureWithBreadAndTags(ctx, err, &sentry.Breadcrumb{
						Category: "kafka",
						Data: map[string]interface{}{
							"groupID": svr.groupID,
						},
					})
				}
				return nil
			})
		}

		svr.groupWaitGroup.Go(goroutineCtx, "consume", func(ctx context.Context) error {
			for {
				if err := svr.group.Consume(context.Background(), svr.topics, svr); err != nil {
					svr.ConsumerError(err)
				}

				if ctx.Err() != nil {
					return nil
				}
			}
		})

		return svr.groupWaitGroup.Wait()
	})
}

// 实现ServerInterface
func (svr *KafkaServer) Name() string {
	return fmt.Sprintf("%s", svr.groupID)
}
