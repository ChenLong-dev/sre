package subscribe

import (
	"context"
	"github.com/Shopify/sarama"
	framework "gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// ====================
// >>>请勿删除<<<
//
// 获取Kafka消息队列订阅服务器
//
// 该服务器展示了基本的消费组消费
// ====================
func GetKafkaServer() framework.ServerInterface {
	svr := new(framework.KafkaServer)

	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况，更改为配置文件中的消费组id
	// ====================
	// 设置消费组id
	svr.SetGroupID("app_framework")
	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况，更改为消费的主题
	// ====================
	// 设置消费主题
	svr.SetTopics([]string{"media_resource_staging"})
	// ====================
	// 设置消费失败函数
	//
	// 当开启消费失败返回时必填
	// ====================
	svr.ConsumerError = func(err error) {
		log.Errorv(context.Background(), errcode.GetErrorMessageMap(err))
	}
	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况，更改为配置文件中的队列名
	// ====================
	// 设置队列名
	svr.ConsumerConsume = func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
		for msg := range claim.Messages() {
			err := service.SVC.PrintKafkaMessageBody(context.Background(), msg)
			if err != nil {
				log.Errorv(context.Background(), errcode.GetErrorMessageMap(err))
			} else {
				session.MarkMessage(msg, "")
			}
		}
		return nil
	}

	return svr
}
