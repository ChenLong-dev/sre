package queue

import (
	"context"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// 投递选项
type PushOption struct {
	// 为true时
	// 如果exchange无法找到合适的queue，会将消息返回给生产者
	// 需使用channel的`NotifyReturn`方法监听返回
	// 为false时
	// 直接丢弃消息
	Mandatory bool
	// 为true时
	// 如果exchange绑定的queue上没有消费者，则不会投递
	// 若所有queue上均没有消费者，会将消息返回给生产者
	// 需使用channel的`NotifyReturn`方法监听返回
	// 为false时
	// 直接投递消息
	Immediate bool
}

// 发送消息
func (session *Session) push(_ context.Context, msg *amqp.Publishing, opt PushOption) error {
	if !session.isReady {
		return ErrNotConnected
	}

	return session.ChannelDo(func(ch *amqp.Channel) error {
		return ch.Publish(
			session.config.ExchangeName,
			session.config.RoutingKey,
			opt.Mandatory,
			opt.Immediate,
			*msg,
		)
	})
}

// 发送消息，不会确认
func (session *Session) UnsafePush(ctx context.Context, msg *amqp.Publishing, opt PushOption) error {
	if !session.config.DisableConfirmMode {
		return errors.New("confirm mode is enable, please use `Push` method")
	}
	return session.push(ctx, msg, opt)
}

// 发送消息，直到收到确认消息
func (session *Session) Push(ctx context.Context, msg *amqp.Publishing, opt PushOption) error {
	if session.config.DisableConfirmMode {
		return errors.New("confirm mode is disable, please use `UnsafePush` method")
	}

	for {
		err := session.push(ctx, msg, opt)
		// 发送错误
		if err != nil {
			session.log(ctx, &spanInfo{
				msg:         msg.Body,
				contentType: msg.ContentType,
				extra:       "push failed",
				err:         err,
			})
			return err
		}

		confirm, ok := <-session.notifyConfirm
		if !ok {
			return errors.New("current confirm channel is closed")
		}
		// 确认收到
		if confirm.Ack {
			session.log(ctx, &spanInfo{
				msg:         msg.Body,
				contentType: msg.ContentType,
				extra:       "push success",
			})
			return nil
		}
	}
}
