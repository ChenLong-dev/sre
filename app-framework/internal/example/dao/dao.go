package dao

import (
	"context"
	"github.com/Shopify/sarama"
	"gitlab.shanhai.int/sre/app-framework/internal/example/config"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/database/etcd"
	"gitlab.shanhai.int/sre/library/database/mongo"
	"gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/database/sql"
	"gitlab.shanhai.int/sre/library/kafka"
	"gitlab.shanhai.int/sre/library/net/agollo"
	"gitlab.shanhai.int/sre/library/net/redlock"
	"gitlab.shanhai.int/sre/library/queue"
)

// ====================
// >>>请勿删除<<<
//
// 数据层
// ====================
type Dao struct {
	// ====================
	// 根据实际情况，选择性保留
	// ====================
	// 会员Mongo客户端
	VipMongo *mongo.DB
	// 支付Mongo客户端
	PayMongo *mongo.DB
	// Redis客户端
	Redis *redis.Pool
	// Mysql客户端
	MySQL *sql.OrmDB
	// Redlock客户端
	RedLock *redlock.RedLock
	// Etcd客户端
	Etcd *etcd.DB
	// Etcd中payment相关数据
	EtcdPaymentData *etcd.StoreData
	// Kafka客户端
	KafkaClient *kafka.Client
	// Kafka同步生产者
	KafkaProducer sarama.SyncProducer
	// AMQP客户端
	AMQPClient *queue.Queue
	// AMQP会话
	AMQPSession *queue.Session
	// Apollo客户端
	ApolloClient *agollo.Client
}

// ====================
// >>>请勿删除<<<
//
// 新建数据层
// ====================
func New() (dao *Dao) {
	// ====================
	// 根据实际情况，选择性保留
	// ====================
	redisClient := redis.NewPool(config.Conf.Redis)

	etcdClient := etcd.NewClient(config.Conf.Etcd)
	paymentData, err := etcdClient.GetPrefix("root/payment/staging/")
	if err != nil {
		panic(err)
	}

	kafkaClient := kafka.NewClient(config.Conf.KafkaProducer)
	kafkaProducer, err := kafkaClient.NewSyncProducer()
	if err != nil {
		panic(err)
	}

	amqpClient := queue.New(config.Conf.AMQPProducer)
	amqpSession, err := amqpClient.NewSession("refund")
	if err != nil {
		panic(err)
	}

	dao = &Dao{
		AMQPClient:      amqpClient,
		AMQPSession:     amqpSession,
		KafkaClient:     kafkaClient,
		KafkaProducer:   kafkaProducer,
		MySQL:           sql.NewMySQL(config.Conf.Mysql),
		Redis:           redisClient,
		Etcd:            etcdClient,
		EtcdPaymentData: paymentData,
		VipMongo:        mongo.NewMongo(config.Conf.VipMongo),
		PayMongo:        mongo.NewMongo(config.Conf.PayMongo),
		RedLock:         redlock.New(config.Conf.Redlock, redisClient),
		ApolloClient: agollo.NewClient(&agollo.Config{
			AppID:             config.Conf.Apollo.AppID,
			Cluster:           config.Conf.Apollo.Cluster,
			ServerHost:        config.Conf.Apollo.ServerHost,
			PreloadNamespaces: config.Conf.Apollo.PreloadNamespaces,
		}),
	}
	return
}

// ====================
// >>>请勿删除<<<
//
// 实现数据层接口
// ====================
func (d *Dao) Close(c context.Context) {
	// ====================
	// 根据实际情况，选择性保留
	// ====================
	if d.MySQL != nil {
		d.MySQL.Close()
	}
	if d.Redis != nil {
		d.Redis.Close()
	}
	if d.VipMongo != nil {
		d.VipMongo.Close(c)
	}
	if d.PayMongo != nil {
		d.PayMongo.Close(c)
	}
	if d.Etcd != nil {
		d.Etcd.Close()
	}
	if d.KafkaClient != nil {
		if d.KafkaProducer != nil {
			d.KafkaProducer.Close()
		}
		d.KafkaClient.Close()
	}
	if d.AMQPSession != nil {
		d.AMQPSession.Close()
	}
	if d.ApolloClient != nil {
		d.ApolloClient.Close()
	}
}

// ====================
// >>>请勿删除<<<
//
// 用于拷贝数据层的键结构体
// ====================
type CloneKey string

// ====================
// >>>请勿删除<<<
//
// 用于拷贝数据层的键
// ====================
const (
	CKMySQL CloneKey = "MySQL"
)

// ====================
// >>>请勿删除<<<
//
// 用于拷贝数据层的选项
// ====================
type CloneOption struct {
	Key   CloneKey
	Value interface{}
}

// ====================
// >>>请勿删除<<<
//
// 拷贝数据层方法
// 常用于事务
// ====================
func (d *Dao) Clone(options ...CloneOption) (*Dao, error) {
	cloneDao := new(Dao)

	err := deepcopy.Copy(d).To(cloneDao)
	if err != nil {
		return nil, err
	}

	// 手动拷贝指定类型
	for _, option := range options {
		switch option.Key {
		case CKMySQL:
			cloneDao.MySQL = option.Value.(*sql.OrmDB)
		}
	}

	return cloneDao, nil
}
