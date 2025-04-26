package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 客户端
type Client struct {
	// 基础配置文件
	config *Config
	// 完整配置文件
	nativeConfig *sarama.Config
	// 客户端
	sarama.Client
}

func getAddrs(endpoints []*EndpointConfig) []string {
	addrs := make([]string, 0)
	for _, endpoint := range endpoints {
		addrs = append(addrs, fmt.Sprintf("%s:%d", endpoint.Address, endpoint.Port))
	}
	return addrs
}

// 新建原始客户端
func NewNativeClient(endpoints []*EndpointConfig, config *sarama.Config) (client *Client) {
	client = new(Client)
	client.nativeConfig = config

	kafkaClient, err := sarama.NewClient(getAddrs(endpoints), config)
	if err != nil {
		panic(err)
	}
	client.Client = kafkaClient

	return client
}

// 新建客户端
func NewClient(config *Config) (client *Client) {
	if config.Config == nil {
		config.Config = &render.Config{}
	}

	sarama.Logger = getDefaultWriter(config)

	nativeConfig, err := GetFullConfigByDefault(config)
	if err != nil {
		panic(err)
	}

	client = NewNativeClient(config.Endpoints, nativeConfig)
	client.config = config

	return client
}

// 新建异步生产者
func (c *Client) NewAsyncProducer() (sarama.AsyncProducer, error) {
	return sarama.NewAsyncProducerFromClient(c)
}

// 新建同步生产者
func (c *Client) NewSyncProducer() (sarama.SyncProducer, error) {
	return sarama.NewSyncProducerFromClient(c)
}

// 新建单个消费者
func (c *Client) NewConsumer() (sarama.Consumer, error) {
	return sarama.NewConsumerFromClient(c)
}

// 新建消费组
func (c *Client) NewConsumerGroup(groupID string) (sarama.ConsumerGroup, error) {
	return sarama.NewConsumerGroupFromClient(groupID, c)
}

// 获取完整配置文件
func (c *Client) GetFullConfig() *sarama.Config {
	return c.nativeConfig
}

// 重设完整配置文件
func (c *Client) ResetFullConfig(config *sarama.Config) error {
	kafkaClient, err := sarama.NewClient(getAddrs(c.config.Endpoints), config)
	if err != nil {
		return err
	}

	c.Client = kafkaClient
	return nil
}

// 关闭客户端
func (c *Client) Close() error {
	return c.Client.Close()
}
