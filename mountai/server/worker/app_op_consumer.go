package worker

import (
	"rulai/config"
	"rulai/service"

	"context"
	"strings"

	"github.com/Shopify/sarama"
	uuid "github.com/satori/go.uuid"
	framework "gitlab.shanhai.int/sre/app-framework"
	_context "gitlab.shanhai.int/sre/library/base/context"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// AppOpMsgConsumer 应用操作消息订阅
func AppOpMsgConsumer() framework.ServerInterface {
	svr := new(framework.KafkaServer)

	svr.SetGroupID(config.Conf.AppOpConsumer.GroupID)
	svr.SetTopics([]string{config.Conf.AppOpConsumer.Topic})

	svr.ConsumerError = func(err error) {
		log.Errorv(context.Background(), errcode.GetErrorMessageMap(err))
	}

	svr.ConsumerConsume = func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
		for msg := range claim.Messages() {
			uid := strings.ReplaceAll(uuid.NewV4().String(), "-", "")
			opCtx := context.WithValue(
				context.Background(),
				_context.ContextUUIDKey,
				uid,
			)

			log.Infoc(opCtx, "Message topic:%q partition:%d offset:%d message=%s uuid:%s",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Value), uid)

			err := service.SVC.HandleAppOpMsgEvent(opCtx, msg)
			if err != nil {
				log.Errorv(opCtx, errcode.GetErrorMessageMap(err))
			}
			session.MarkMessage(msg, "")
		}

		return nil
	}

	return svr
}
