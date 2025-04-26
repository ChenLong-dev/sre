package kafka

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

const (
	// 生产者要求回复的类型
	// 不回复
	RequiredAckTypeNoResponse = "no"
	// 本地回复
	RequiredAckTypeWaitForLocal = "local"
	// 同步的节点全部回复
	RequiredAckTypeWaitForAll = "all"

	// 生产者投递分区策略
	// 手动投递
	ProducerPartitionStrategyManual = "manual"
	// 随机投递
	ProducerPartitionStrategyRandom = "random"
	// 轮询投递
	ProducerPartitionStrategyRoundRobin = "round_robin"
	// 哈希投递
	ProducerPartitionStrategyHash = "hash"

	// 消费者平衡策略
	// 区域平衡
	BalanceStrategyRange = "range"
	// 轮询平衡
	BalanceStrategyRoundRobin = "round_robin"
)

type EndpointConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

// 配置文件
type Config struct {
	// 集群地址
	Endpoints []*EndpointConfig `yaml:"endpoints"`
	// App ID
	AppID string `yaml:"appID"`
	// 缓冲区大小
	ChannelBufferSize int `yaml:"channelBufferSize"`
	// 版本号
	Version string `yaml:"version"`
	// 消费者配置
	Consumer *ConsumerConfig `yaml:"consumer"`
	// 生产者配置
	Producer *ProducerConfig `yaml:"producer"`

	// 日志配置
	*render.Config `yaml:",inline"`
}

// 消费者配置
type ConsumerConfig struct {
	// 平衡策略
	BalanceStrategy string `yaml:"balanceStrategy"`
	// 最大处理消息时间
	// 超时会抛出异常
	MaxProcessingTime ctime.Duration `yaml:"maxProcessingTime"`
	// 是否返回错误
	ReturnError bool `yaml:"returnError"`
	// 初始化偏移量
	// 只有在消费组模式下才有效
	// 只有非0值才有效
	InitialOffset int64 `yaml:"initialOffset"`
}

// 生产者配置
type ProducerConfig struct {
	// 要求回复的类型
	RequiredAckType string `yaml:"requiredAckType"`
	// 要求回复的超时时间
	// 只有在 `RequiredAckTypeWaitForAll` 类型下才有效
	RequiredAckTimeout ctime.Duration `yaml:"requiredAckTimeout"`
	// 分区策略
	PartitionStrategy string `yaml:"partitionStrategy"`
	// 是否返回成功
	ReturnSuccess bool `yaml:"returnSuccess"`
	// 是否返回失败
	ReturnError bool `yaml:"returnError"`
}

// 通过默认配置获取完整配置
func GetFullConfigByDefault(config *Config) (*sarama.Config, error) {
	kafkaConfig := sarama.NewConfig()

	if len(config.Endpoints) < 1 {
		return nil, errors.New("endpoints is empty")
	}

	if config.Version != "" {
		version, err := sarama.ParseKafkaVersion(config.Version)
		if err != nil {
			return nil, err
		}
		kafkaConfig.Version = version
	}

	if config.AppID != "" {
		kafkaConfig.ClientID = config.AppID
	}
	if config.ChannelBufferSize != 0 {
		kafkaConfig.ChannelBufferSize = config.ChannelBufferSize
	}

	if config.Producer != nil {
		switch config.Producer.RequiredAckType {
		case "":
		case RequiredAckTypeNoResponse:
			kafkaConfig.Producer.RequiredAcks = sarama.NoResponse
		case RequiredAckTypeWaitForLocal:
			kafkaConfig.Producer.RequiredAcks = sarama.WaitForLocal
		case RequiredAckTypeWaitForAll:
			kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
		default:
			return nil, errors.New(fmt.Sprintf("unknown producer required ack type:%s", config.Producer.RequiredAckType))
		}

		if config.Producer.RequiredAckTimeout != 0 {
			kafkaConfig.Producer.Timeout = time.Duration(config.Producer.RequiredAckTimeout)
		}

		switch config.Producer.PartitionStrategy {
		case "":
		case ProducerPartitionStrategyManual:
			kafkaConfig.Producer.Partitioner = sarama.NewManualPartitioner
		case ProducerPartitionStrategyRandom:
			kafkaConfig.Producer.Partitioner = sarama.NewRandomPartitioner
		case ProducerPartitionStrategyRoundRobin:
			kafkaConfig.Producer.Partitioner = sarama.NewRoundRobinPartitioner
		case ProducerPartitionStrategyHash:
			kafkaConfig.Producer.Partitioner = sarama.NewHashPartitioner
		default:
			return nil, errors.New(fmt.Sprintf("unknown producer parition:%s", config.Producer.PartitionStrategy))
		}

		kafkaConfig.Producer.Return.Errors = config.Producer.ReturnError
		kafkaConfig.Producer.Return.Successes = config.Producer.ReturnSuccess
	}

	if config.Consumer != nil {
		switch config.Consumer.BalanceStrategy {
		case "":
		case BalanceStrategyRange:
			kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
		case BalanceStrategyRoundRobin:
			kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
		default:
			return nil, errors.New(fmt.Sprintf("unknown consumer balance strategy:%s", config.Consumer.BalanceStrategy))
		}

		if config.Consumer.MaxProcessingTime != 0 {
			kafkaConfig.Consumer.MaxProcessingTime = time.Duration(config.Consumer.MaxProcessingTime)
		}
		if config.Consumer.InitialOffset != 0 {
			kafkaConfig.Consumer.Offsets.Initial = config.Consumer.InitialOffset
		}

		kafkaConfig.Consumer.Return.Errors = config.Consumer.ReturnError
	}

	return kafkaConfig, nil
}
