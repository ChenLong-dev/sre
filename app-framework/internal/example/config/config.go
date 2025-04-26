package config

import (
	"gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/library/database/etcd"
	"gitlab.shanhai.int/sre/library/database/mongo"
	"gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/database/sql"
	"gitlab.shanhai.int/sre/library/kafka"
	"gitlab.shanhai.int/sre/library/net/httpclient"
	"gitlab.shanhai.int/sre/library/net/redlock"
	"gitlab.shanhai.int/sre/library/queue"
)

var (
	// ====================
	// >>>请勿删除<<<
	//
	// 全局配置文件
	// ====================
	Conf *Config
)

// ====================
// 自定义配置文件
//
// 根据实际情况，选择性添加
// ====================

// host配置文件
type HostsConfig struct {
	Github   string `yaml:"github"`
	UserAuth string `yaml:"userAuth"`
}

// apollo 配置
type ApolloConfig struct {
	AppID             string   `yaml:"appID"`
	Cluster           string   `yaml:"cluster"`
	ServerHost        string   `yaml:"serverHost"`
	PreloadNamespaces []string `yaml:"namespaceNames"`
}

// ====================
// >>>请勿删除<<<
//
// 配置文件
// ====================
type Config struct {
	// ====================
	// >>>请勿删除<<<
	//
	// 基础配置文件
	// ====================
	*framework.Config `yaml:",inline"`

	// ====================
	// 根据实际情况，选择性保留
	// ====================
	// 会员Mongo配置文件
	VipMongo *mongo.Config `yaml:"vipMongo"`
	// 支付Mongo配置文件
	PayMongo *mongo.Config `yaml:"payMongo"`
	// Http客户端配置文件
	HttpClient *httpclient.Config `yaml:"httpClient"`
	// Mysql配置文件
	Mysql *sql.Config `yaml:"mysql"`
	// Redis配置文件
	Redis *redis.Config `yaml:"redis"`
	// Redlock配置文件
	Redlock *redlock.Config `yaml:"redlock"`
	// Etcd配置文件
	Etcd *etcd.Config `yaml:"etcd"`
	// Kafka生产者配置文件
	KafkaProducer *kafka.Config `yaml:"kafkaProducer"`
	// AMQP生产者配置文件
	AMQPProducer *queue.Config `yaml:"amqpProducer"`
	// apollo配置
	Apollo *ApolloConfig `yaml:"apollo"`
	// host配置文件
	Host *HostsConfig `yaml:"host"`
}

// ====================
// >>>请勿删除<<<
//
// 读取配置文件
//
//	configPath为配置文件的所属位置，例如 ./config/config.yaml
//
// ====================
func Read(configPath string) *Config {
	Conf = new(Config)

	// 解码yaml配置文件
	framework.DecodeConfig(configPath, Conf)

	return Conf
}
