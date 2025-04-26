package queue

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"context"
	"os"
	"strconv"
	"sync/atomic"
)

// 消费选项
type ConsumeOption struct {
	// 消费者名称
	ConsumerName string
	// 为true时
	// 将消息发送给消费者后，就会从内存中删除
	// 为false时
	// 若某个消费者没有回复，将会发送给其他消费者
	AutoAck bool
	// 为true时
	// 只能有一个消费者，其他消费者连接时会报错
	// 为false时
	// 无限制
	Exclusive bool
	// 为true时
	// 无需等待服务器确认，即可进行下个操作
	// 若无法消费，将会报错，并且关闭channel
	// 为false时
	// 等待服务器确认
	NoWait bool
	// 额外参数
	Args amqp.Table

	// 每次服务器在收到ack前，发送给消费者消息的最大数量
	PrefetchCount int
	// 每次服务器在收到ack前，刷新到网络的最大字节大小
	PrefetchSize int
	// 为true时
	// 则应用该连接上的所有channel
	// 为false时
	// 则只应用当前channel
	Global bool
}

// 消费者初始化函数
type ConsumerInitFunc func(ch *amqp.Channel, opt ConsumeOption) (<-chan amqp.Delivery, error)

// 消费函数
type ConsumeFunc func(d amqp.Delivery) error

// 消费者序列
var consumerSeq uint64

// 消费者tag最长限制
const consumerTagLengthMax = 0xFF

// 获取唯一消费者标示
func getUniqueConsumerTag() string {
	tagPrefix := "ctag-"
	tagInfix := os.Args[0]
	tagSuffix := "-" + strconv.FormatUint(atomic.AddUint64(&consumerSeq, 1), 10)

	if len(tagPrefix)+len(tagInfix)+len(tagSuffix) > consumerTagLengthMax {
		tagInfix = "streadway/amqp"
	}

	return tagPrefix + tagInfix + tagSuffix
}

// 消费流
func (session *Session) stream(ctx context.Context, initFunc ConsumerInitFunc, consumeFunc ConsumeFunc, opt ConsumeOption) error {
	if !session.isReady {
		return ErrNotConnected
	}
	if opt.ConsumerName == "" {
		opt.ConsumerName = getUniqueConsumerTag()
	}

	var (
		err               error
		curChannel        *amqp.Channel
		msgChannel        = make(<-chan amqp.Delivery)
		notifyStreamClose = session.notifyStreamClose()
		notifyInitChannel = session.notifyInitChannel
	)
	for {
		select {
		case <-notifyStreamClose:
			err = session.channelDo(curChannel, func(ch *amqp.Channel) error {
				return ch.Cancel(opt.ConsumerName, false)
			})
			if err != nil {
				return err
			}
			return nil
		// 监听channel初始化
		// 避免断连，导致旧channel不可用
		case ch := <-notifyInitChannel:
			if ch == nil {
				notifyInitChannel = session.notifyInitChannel
				continue
			}
			curChannel = ch
			session.log(context.Background(), &spanInfo{
				consumerName: opt.ConsumerName,
				extra:        "start init consumer",
			})
			err = session.channelDo(ch, func(ch *amqp.Channel) error {
				msgChannel, err = initFunc(ch, opt)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				session.log(ctx, &spanInfo{
					consumerName: opt.ConsumerName,
					extra:        "init consumer error",
					err:          err,
				})
				return err
			}
			session.log(context.Background(), &spanInfo{
				consumerName: opt.ConsumerName,
				extra:        "init consumer success",
			})
		case d, ok := <-msgChannel:
			if !ok {
				continue
			}
			err = session.ChannelDo(func(ch *amqp.Channel) error {
				return consumeFunc(d)
			})
			if err != nil {
				session.log(ctx, &spanInfo{
					consumerName: opt.ConsumerName,
					msg:          d.Body,
					contentType:  d.ContentType,
					err:          err,
					extra:        "consume error",
				})
			}
		}
	}
}

// 通道消息消费流
func (session *Session) Stream(ctx context.Context, consumeFunc ConsumeFunc, opt ConsumeOption) error {
	return session.stream(ctx,
		func(ch *amqp.Channel, opt ConsumeOption) (<-chan amqp.Delivery, error) {
			return ch.Consume(
				session.config.QueueName,
				opt.ConsumerName,
				opt.AutoAck,
				opt.Exclusive,
				false,
				opt.NoWait,
				opt.Args,
			)
		},
		consumeFunc, opt)
}

// 默认消费者初始化函数
var defaultConsumerInitFunc = func(session *Session, ch *amqp.Channel, opt ConsumeOption) (<-chan amqp.Delivery, error) {
	err := ch.QueueBind(
		session.config.QueueName,
		session.config.RoutingKey,
		session.config.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	// 设置通道消费速度
	err = ch.Qos(opt.PrefetchCount, opt.PrefetchSize, opt.Global)
	if err != nil {
		return nil, err
	}

	return ch.Consume(
		session.config.QueueName,
		opt.ConsumerName,
		opt.AutoAck,
		opt.Exclusive,
		false,
		opt.NoWait,
		opt.Args,
	)
}

// 获取非自动回复消费流
func (session *Session) NoAutoAckConsumeStream(ctx context.Context, consumeFunc ConsumeFunc, opt ConsumeOption) error {
	if opt.AutoAck {
		return errors.New("config `AutoAck` is enable, please use other method")
	}
	return session.stream(ctx,
		func(ch *amqp.Channel, opt ConsumeOption) (<-chan amqp.Delivery, error) {
			opt.AutoAck = false
			return defaultConsumerInitFunc(session, ch, opt)
		}, consumeFunc, opt)
}

// 简单获取消费流
func (session *Session) SingleConsumeStream(ctx context.Context, consumeFunc ConsumeFunc, opt ConsumeOption) error {
	return session.stream(ctx,
		func(ch *amqp.Channel, opt ConsumeOption) (<-chan amqp.Delivery, error) {
			if opt.PrefetchCount == 0 {
				opt.PrefetchCount = 1
			}
			opt.AutoAck = true
			return defaultConsumerInitFunc(session, ch, opt)
		}, consumeFunc, opt)
}
