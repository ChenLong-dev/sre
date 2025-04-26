package queue

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"

	"github.com/streadway/amqp"

	"time"
)

const (
	// DefaultConnectTimeout 默认连接超时时间
	DefaultConnectTimeout = time.Second * 5
	// DefaultReconnectDelay 默认连接失败时重连的延迟
	DefaultReconnectDelay = time.Second * 3
	// DefaultReInitDelay 默认channel异常时重新初始化的延迟
	DefaultReInitDelay = time.Second * 3

	// connectCheckInterval 连接检查时间间隔
	connectCheckInterval = time.Millisecond * 500
)

type EndpointConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type Config struct {
	// 服务器地址
	Endpoint *EndpointConfig `yaml:"endpoint"`
	// 用户名
	UserName string `yaml:"userName"`
	// 密码
	Password string `yaml:"password"`
	// 虚拟主机
	VHost string `yaml:"vHost"`
	// 会话配置
	Session map[string]*SessionConfig `yaml:"session"`

	// 连接超时时间
	ConnectTimeout ctime.Duration `yaml:"connectTimeout"`
	// 当连接失败时重连的延迟
	ReconnectDelay ctime.Duration `yaml:"reconnectDelay"`
	// 当channel异常时重新初始化的延迟
	ReInitDelay ctime.Duration `yaml:"reInitDelay"`

	// 日志配置
	*render.Config `yaml:",inline"`
}

type SessionConfig struct {
	// 是否需要关闭消息发送确认机制
	// 若开启该模式，重发消息可能会导致死锁
	DisableConfirmMode bool `yaml:"disableConfirmMode"`

	// 队列名
	// Deprecated: 请使用 QueueName
	Name string `yaml:"name"`
	// 队列名
	QueueName string `yaml:"queueName"`
	// 交换器名
	ExchangeName string `yaml:"exchangeName"`
	// 路由key
	RoutingKey string `yaml:"routingKey"`
	// 是否持久化
	Durable bool `yaml:"durable"`
	// 是否自动删除
	AutoDelete bool `yaml:"autoDelete"`
	// 是否设置排他
	Exclusive bool `yaml:"exclusive"`
	// 是否非阻塞
	NoWait bool `yaml:"noWait"`
	// 额外参数
	Args amqp.Table `yaml:"args"`

	// 连接超时时间
	ConnectTimeout ctime.Duration `yaml:"connectTimeout"`
	// 当连接失败时重连的延迟
	ReconnectDelay ctime.Duration `yaml:"reconnectDelay"`
	// 当channel异常时重新初始化的延迟
	ReInitDelay ctime.Duration `yaml:"reInitDelay"`
}
