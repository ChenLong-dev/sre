package subscribe

import (
	"context"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
	framework "gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/app-framework/internal/example/service"
	_context "gitlab.shanhai.int/sre/library/base/context"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/queue"
)

// ====================
// >>>请勿删除<<<
//
// 获取AMQP消息队列订阅服务器
//
// 若消费服务无法处理消息或消费服务意外终止/报错时，应当在保证消费函数幂等性的前提下，
// 可以通过 手动回复(ack)的方式，避免消息消费失败或丢失
// ====================
func GetAMQPServer() framework.ServerInterface {
	svr := new(framework.AMQPServer)

	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况，更改为配置文件中的队列名
	// ====================
	// 设置队列名
	svr.SetQueueName("app_framework")

	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况，更改为配置文件中的会话名
	// ====================
	// 设置会话名
	svr.SetSessionName("refund")

	// ====================
	// >>>请勿删除<<<
	//
	// 设置订阅函数
	// ====================
	svr.SubscribeFunction = func(ctx context.Context, queueConfig *queue.Config, session *queue.Session) error {
		err := session.NoAutoAckConsumeStream(ctx, func(d amqp.Delivery) error {
			currentContext, cancel := context.WithCancel(
				context.WithValue(ctx, _context.ContextUUIDKey, uuid.NewV4().String()),
			)
			defer cancel()

			// ====================
			// 消费
			//
			// 根据实际情况，修改消费函数
			// ====================
			err := service.SVC.PrintAMQPMessageBody(currentContext, d.Body)
			if err != nil {
				log.Errorv(currentContext, errcode.GetErrorMessageMap(err))

				// ====================
				// 处理失败，手动拒绝
				// ====================
				rejectErr := d.Reject(true)
				if rejectErr != nil {
					log.Errorv(currentContext, errcode.GetErrorMessageMap(rejectErr))
					return rejectErr
				}

				return err
			}

			// ====================
			// 处理成功，手动回复
			// ====================
			err = d.Ack(false)
			if err != nil {
				log.Errorv(currentContext, errcode.GetErrorMessageMap(err))
				return err
			}

			return nil
		}, queue.ConsumeOption{
			ConsumerName:  "example",
			PrefetchCount: 20,
			Global:        true,
		})
		if err != nil {
			return err
		}

		return nil
	}

	return svr
}
