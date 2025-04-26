package dao

import (
	"rulai/config"

	"context"

	"github.com/Shopify/sarama"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/database/mongo"
	"gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/database/sql"
	"gitlab.shanhai.int/sre/library/kafka"
	"gitlab.shanhai.int/sre/library/net/redlock"
)

// ====================
// >>>请勿删除<<<
//
// 数据层
// ====================
type Dao struct {
	ApolloPrdMysql *sql.OrmDB
	ApolloStgMysql *sql.OrmDB
	Mongo          *mongo.DB
	Redis          *redis.Pool
	Redlock        *redlock.RedLock
	KafkaClient    *kafka.Client
	KafkaProducer  sarama.SyncProducer
}

// ====================
// >>>请勿删除<<<
//
// 新建数据层
// ====================
func New() (dao *Dao, err error) {
	r := redis.NewPool(config.Conf.Redis)

	kafkaClient := kafka.NewClient(config.Conf.KafkaProducer)
	kafkaProducer, err := kafkaClient.NewSyncProducer()
	if err != nil {
		return nil, err
	}

	dao = &Dao{
		// ApolloPrdMysql: sql.NewMySQL(config.Conf.ApolloPrdMysql),
		// ApolloStgMysql: sql.NewMySQL(config.Conf.ApolloStgMysql),
		Mongo:         mongo.NewMongo(config.Conf.Mongo),
		Redis:         r,
		Redlock:       redlock.New(config.Conf.Redlock, r),
		KafkaClient:   kafkaClient,
		KafkaProducer: kafkaProducer,
	}

	return dao, nil
}

// ====================
// >>>请勿删除<<<
//
// 拷贝数据层方法
// 常用于事务
// ====================
func (d *Dao) Clone() (*Dao, error) {
	cloneDao := new(Dao)

	err := deepcopy.Copy(d).To(cloneDao)
	if err != nil {
		return nil, err
	}

	return cloneDao, nil
}

// ====================
// >>>请勿删除<<<
//
// 实现数据层接口
// ====================
func (d *Dao) Close(c context.Context) {
	if d.ApolloPrdMysql != nil {
		d.ApolloPrdMysql.Close()
	}

	if d.ApolloStgMysql != nil {
		d.ApolloStgMysql.Close()
	}

	if d.Redis != nil {
		d.Redis.Close()
	}

	if d.Mongo != nil {
		d.Mongo.Close(c)
	}

	if d.KafkaClient != nil {
		if d.KafkaProducer != nil {
			d.KafkaProducer.Close()
		}
		d.KafkaClient.Close()
	}
}
