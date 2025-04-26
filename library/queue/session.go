package queue

import (
	"gitlab.shanhai.int/sre/library/base/hook"
	render "gitlab.shanhai.int/sre/library/base/logrender"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"context"
	"time"
)

// 会话
type Session struct {
	// ID
	id string
	// 会话配置文件
	config *SessionConfig

	// 钩子管理器
	manager *hook.Manager
	// 连接
	curConnection *amqp.Connection
	// channel通道
	curChannel *amqp.Channel

	// 会话是否关闭
	done chan bool
	// 通知消费流关闭
	notifyStreamCloses []chan bool
	// 通知通道初始化
	notifyInitChannel chan *amqp.Channel
	// 用于在重连逻辑中，通知连接关闭
	notifyConnClose chan *amqp.Error
	// 用于在重连逻辑中，通知通道关闭
	notifyChanClose chan *amqp.Error
	// 通知确认收到消息
	notifyConfirm chan amqp.Confirmation

	// 通道是否可用
	isReady bool
}

var (
	// ErrNotConnected 连接错误
	ErrNotConnected = errors.New("not connected to a server")
	// ErrCloseNotReady 关闭未就绪的连接错误
	ErrCloseNotReady = errors.New("can't close not ready channel or connection")
	// ErrCurrentChannelClosed channel已关闭错误
	ErrCurrentChannelClosed = errors.New("current channel closed")
	// ErrCurrentConnectionClosed 连接已关闭错误
	ErrCurrentConnectionClosed = errors.New("current connection closed")
)

// 打印日志
func (session *Session) log(ctx context.Context, spanInfo *spanInfo) {
	hk := session.manager.CreateHook(ctx).
		AddArg("session_id", session.id).
		AddArg("queue_name", session.config.QueueName).
		AddArg("extra", spanInfo.extra).
		AddArg("exchange_name", session.config.ExchangeName).
		AddArg("routing_key", session.config.RoutingKey).
		AddArg("consumer_name", spanInfo.consumerName)
	if spanInfo.err != nil {
		hk = hk.AddArg(render.ErrorArgKey, spanInfo.err)
	}
	if spanInfo.msg != nil {
		hk = hk.AddArg("body", string(spanInfo.msg))
	}
	if spanInfo.contentType != "" {
		hk = hk.AddArg("content_type", spanInfo.contentType)
	}

	hk.ProcessAfterHook()
}

func (session *Session) connectionDo(conn *amqp.Connection, f func(conn *amqp.Connection) error) error {
	notifyClose := make(chan *amqp.Error, 1)
	conn.NotifyClose(notifyClose)

	finish := make(chan error)
	go func() {
		finish <- f(conn)
	}()

	select {
	case err := <-finish:
		return err
	case err := <-notifyClose:
		if err != nil {
			return errors.Wrapf(ErrCurrentConnectionClosed, "%s", err)
		}
		return ErrCurrentConnectionClosed
	}
}

func (session *Session) ConnectionDo(f func(conn *amqp.Connection) error) error {
	if !session.isReady {
		return ErrNotConnected
	}
	return session.connectionDo(session.curConnection, f)
}

func (session *Session) channelDo(ch *amqp.Channel, f func(ch *amqp.Channel) error) error {
	notifyConnClose := make(chan *amqp.Error, 1)
	session.curConnection.NotifyClose(notifyConnClose)
	notifyChanClose := make(chan *amqp.Error, 1)
	ch.NotifyClose(notifyChanClose)

	finish := make(chan error)
	go func() {
		finish <- f(ch)
	}()

	select {
	case err := <-finish:
		return err
	case err := <-notifyConnClose:
		if err != nil {
			return errors.Wrapf(ErrCurrentConnectionClosed, "%s", err)
		}
		return ErrCurrentConnectionClosed
	case err := <-notifyChanClose:
		if err != nil {
			return errors.Wrapf(ErrCurrentChannelClosed, "%s", err)
		}
		return ErrCurrentChannelClosed
	}
}

func (session *Session) ChannelDo(f func(ch *amqp.Channel) error) error {
	if !session.isReady {
		return ErrNotConnected
	}
	return session.channelDo(session.curChannel, f)
}

// 连接
func (session *Session) connect(addr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	session.curConnection = conn
	session.notifyConnClose = make(chan *amqp.Error, 1)
	session.curConnection.NotifyClose(session.notifyConnClose)

	return conn, nil
}

// 处理重连
func (session *Session) handleReconnect(addr string) {
	for {
		session.isReady = false
		session.log(context.Background(), &spanInfo{
			extra: "start connect",
		})

		conn, err := session.connect(addr)
		// 连接错误
		if err != nil {
			session.log(context.Background(), &spanInfo{
				extra: "failed to connect, retrying",
				err:   err,
			})

			select {
			case <-session.done:
				return
			case <-time.After(time.Duration(session.config.ReconnectDelay)):
			}

			continue
		}
		session.log(context.Background(), &spanInfo{
			extra: "connect success",
		})

		// 初始化
		if done := session.handleReInit(conn); done {
			break
		}
	}
}

// 初始化队列
func (session *Session) init(c *amqp.Connection) error {
	err := session.connectionDo(c, func(conn *amqp.Connection) error {
		ch, err := conn.Channel()
		if err != nil {
			return err
		}
		session.curChannel = ch
		return nil
	})
	if err != nil {
		return err
	}

	err = session.channelDo(session.curChannel, func(ch *amqp.Channel) error {
		err = ch.Confirm(false)
		if err != nil {
			return err
		}

		_, err = ch.QueueDeclare(
			session.config.QueueName,
			session.config.Durable,
			session.config.AutoDelete,
			session.config.Exclusive,
			session.config.NoWait,
			session.config.Args,
		)
		if err != nil {
			return err
		}

		session.notifyChanClose = make(chan *amqp.Error, 1)
		ch.NotifyClose(session.notifyChanClose)
		if !session.config.DisableConfirmMode {
			session.notifyConfirm = make(chan amqp.Confirmation, 1)
			ch.NotifyPublish(session.notifyConfirm)
		}

		if session.notifyInitChannel != nil {
			close(session.notifyInitChannel)
		}
		session.notifyInitChannel = make(chan *amqp.Channel, 1)
		session.notifyInitChannel <- ch

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// 处理重新初始化
func (session *Session) handleReInit(conn *amqp.Connection) bool {
	for {
		session.isReady = false
		session.log(context.Background(), &spanInfo{
			extra: "start init channel",
		})

		err := session.init(conn)
		// 初始化错误
		if err != nil {
			session.log(context.Background(), &spanInfo{
				extra: "failed to init channel, retrying",
				err:   err,
			})
			if errors.Is(err, ErrCurrentConnectionClosed) {
				return false
			}

			select {
			case <-session.done:
				return true
			case <-time.After(time.Duration(session.config.ReInitDelay)):
			}

			continue
		}
		session.isReady = true
		session.log(context.Background(), &spanInfo{
			extra: "init channel success",
		})

		select {
		case <-session.done:
			return true
		case e := <-session.notifyConnClose:
			info := &spanInfo{
				extra: "connection closed, need reconnect",
			}
			if e != nil {
				info.err = e
			}
			session.log(context.Background(), info)
			return false
		case e := <-session.notifyChanClose:
			info := &spanInfo{
				extra: "channel closed, rerun init",
			}
			if e != nil {
				info.err = e
			}
			session.log(context.Background(), info)
		}
	}
}

// 是否就绪
func (session *Session) IsReady() bool {
	return session.isReady
}

// 获取配置文件
func (session *Session) GetConfig() *SessionConfig {
	return session.config
}

// 通知流关闭
func (session *Session) notifyStreamClose() chan bool {
	c := make(chan bool)
	session.notifyStreamCloses = append(session.notifyStreamCloses, c)
	return c
}

// 关闭所有消费流
func (session *Session) CloseStream() {
	for _, c := range session.notifyStreamCloses {
		close(c)
	}
	session.notifyStreamCloses = nil
}

// 关闭当前channel
func (session *Session) CloseChannel() error {
	defer func() {
		session.isReady = false
	}()

	if !session.isReady {
		return ErrCloseNotReady
	}

	err := session.curChannel.Close()
	if err != nil {
		return err
	}

	return nil
}

// 关闭会话
func (session *Session) Close() error {
	defer func() {
		// 通知关闭
		close(session.done)
		session.done = nil

		session.isReady = false
		session.manager.Close()
	}()

	if !session.isReady {
		return ErrCloseNotReady
	}

	session.CloseStream()
	err := session.curConnection.Close()
	if err != nil {
		return err
	}

	return nil
}
