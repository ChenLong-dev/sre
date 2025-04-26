package dao

import (
	"context"
	"sort"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	_redis "gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"rulai/models/entity"
)

// ConfigRenamePrefixesCacheKey 特殊配置重命名前缀缓存键
const ConfigRenamePrefixesCacheKey = "config_rename_prefixes"

// CreateConfigRenamePrefix 创建特殊配置重命名前缀
func (d *Dao) CreateConfigRenamePrefix(ctx context.Context, prefix *entity.ConfigRenamePrefix) error {
	res, err := d.Mongo.Collection(prefix.TableName()).InsertOne(ctx, prefix)
	if err != nil {
		return errors.Wrap(errcode.MongoError, err.Error())
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return errors.Wrap(errcode.InternalError, "inserted id is not object id")
	}

	// 永久缓存, mongo ID 是固定长度, 直接放在 name 之前即可, 读取时很容易分离
	return d.Redis.WrapDo(func(con *_redis.Conn) error {
		_, e := con.Do(ctx, "hset", ConfigRenamePrefixesCacheKey, prefix.Prefix, id.Hex()+prefix.Name)
		if e != nil {
			return errors.Wrap(errcode.RedisError, err.Error())
		}

		return nil
	})
}

// DeleteConfigRenamePrefix 删除特殊配置重命名前缀
func (d *Dao) DeleteConfigRenamePrefix(ctx context.Context, prefix string) error {
	filter := bson.M{"prefix": prefix}
	_, err := d.Mongo.Collection(new(entity.ConfigRenamePrefix).TableName()).
		DeleteOne(ctx, filter)
	if err != nil {
		return errors.Wrap(errcode.MongoError, err.Error())
	}

	return d.Redis.WrapDo(func(con *_redis.Conn) error {
		_, e := con.Do(ctx, "hdel", ConfigRenamePrefixesCacheKey, prefix)
		if e != nil {
			return errors.Wrap(errcode.RedisError, err.Error())
		}

		return nil
	})
}

// FindAllConfigRenamePrefixes 获取所有的特殊配置重命名前缀(数量很少, 无需分页机制)
func (d *Dao) FindAllConfigRenamePrefixes(ctx context.Context) ([]*entity.ConfigRenamePrefix, error) {
	// 由于是永久缓存, 直接从 redis 获取即可, 前缀理论上不多, 暂不考虑 hgetall 的阻塞问题
	var res []*entity.ConfigRenamePrefix
	err := d.Redis.WrapDo(func(con *_redis.Conn) error {
		values, e := redis.ByteSlices(con.Do(ctx, "hgetall", ConfigRenamePrefixesCacheKey))
		if e != nil {
			return errors.Wrap(errcode.RedisError, e.Error())
		}

		prefixesLength := len(values) / 2 // 每组数据两条 value
		res = make([]*entity.ConfigRenamePrefix, prefixesLength)
		for index := range values {
			if index&1 == 0 {
				res[index/2] = &entity.ConfigRenamePrefix{Prefix: string(values[index])}
			} else {
				// value 的前 24 个字符是 mongo ID, 后面是 name
				res[index/2].ID, e = primitive.ObjectIDFromHex(string(values[index][:24]))
				if e != nil {
					return errors.Wrapf(errcode.InternalError,
						"prefix(%s)'s ID(%s) is invalid mongo ID", res[index/2].Prefix, values[index][:24])
				}

				res[index/2].Name = string(values[index][24:])
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 保持按照 ID 正序输出
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].ID.Hex() < res[j].ID.Hex()
	})

	return res, nil
}
