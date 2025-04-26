package service

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// 打印AMQP消息体
func (s *Service) PrintAMQPMessageBody(ctx context.Context, msg []byte) (err error) {
	// 防止panic中断整个程序
	defer func() {
		if e := recover(); e != nil {
			err = errors.Wrapf(errcode.InternalError, "%s", e)
		}
	}()

	log.Infoc(ctx, "Message queue:app_framework session:order message=%s", string(msg))
	return nil
}

// 打印Kafka消息体
func (s *Service) PrintKafkaMessageBody(ctx context.Context, msg *sarama.ConsumerMessage) (err error) {
	// 防止panic中断整个程序
	defer func() {
		if e := recover(); e != nil {
			err = errors.Wrapf(errcode.InternalError, "%s", e)
		}
	}()

	log.Infoc(ctx, "Message topic:%q partition:%d offset:%d message=%s",
		msg.Topic, msg.Partition, msg.Offset, string(msg.Value))
	return nil
}
