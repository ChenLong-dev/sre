package framework

import (
	"context"
	"fmt"
	"gitlab.shanhai.int/sre/library/queue"
)

// AMQP消息队列监听服务
type AMQPServer struct {
	// 消息队列客户端
	client *queue.Queue
	// 消息队列会话
	session *queue.Session
	// 队列名
	queueName string
	// 会话名
	sessionName string
	// 基础配置文件
	config *Config
	// 当前队列配置文件
	queueConfig *queue.Config
	// 订阅函数
	SubscribeFunction func(ctx context.Context, queueConfig *queue.Config, session *queue.Session) error
}

// 设置会话名
func (svr *AMQPServer) SetSessionName(name string) {
	svr.sessionName = name
}

// 设置队列名
func (svr *AMQPServer) SetQueueName(name string) {
	svr.queueName = name
}

// 获取会话名
func (svr *AMQPServer) GetSessionName() (name string) {
	return svr.sessionName
}

// 获取队列名
func (svr *AMQPServer) GetQueueName() (name string) {
	return svr.queueName
}

// 实现ServerInterface
func (svr *AMQPServer) ShutDown(ctx context.Context) (err error) {
	return svr.session.Close()
}

// 实现ServerInterface
func (svr *AMQPServer) Start(c *Config, svc ServiceInterface) {
	if svr.queueName == "" {
		panic("subscribe server queueName haven't assign")
	}

	if svr.sessionName == "" {
		panic("subscribe server sessionName haven't assign")
	}

	svr.config = c
	if queueConfig, ok := svr.config.AMQP[svr.queueName]; ok {
		svr.queueConfig = queueConfig
	} else {
		panic(fmt.Sprintf("subscribe server %s haven't config", svr.queueName))
	}

	svr.client = queue.New(svr.queueConfig)
	if session, err := svr.client.NewSession(svr.sessionName); err != nil {
		panic(err)
	} else {
		svr.session = session
	}

	svc.StartServer(svr.queueName, func(ctx context.Context) error {
		return svr.SubscribeFunction(ctx, svr.queueConfig, svr.session)
	})
}

// 实现ServerInterface
func (svr *AMQPServer) Name() string {
	return fmt.Sprintf("%s-%s", svr.queueName, svr.sessionName)
}
