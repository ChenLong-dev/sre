package dao

import (
	"context"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"rulai/models/entity"
)

// CreateAppLTSStream 创建AMS应用对应的华为云 LTS 日志流记录, 不需要关注返回的主键
// TODO: 华为云支持通过名称查询日志流后可以去除
func (d *Dao) CreateAppLTSStream(ctx context.Context, stream *entity.AppLTSStream) error {
	_, err := d.Mongo.Collection(entity.EmptyAppLTSStream.TableName()).InsertOne(ctx, stream)
	if err != nil {
		return errors.Wrap(errcode.MongoError, err.Error())
	}
	return nil
}

// DeleteAppLTSStream 删除AMS应用对应的华为云 LTS 日志流记录
// TODO: 华为云支持通过名称查询日志流后可以去除
func (d *Dao) DeleteAppLTSStream(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error {
	_, err := d.Mongo.Collection(entity.EmptyAppLTSStream.TableName()).DeleteOne(ctx, bson.M{
		"cluster_name": clusterName,
		"env_name":     envName,
		"app_id":       appID,
	})
	if err != nil {
		return errors.Wrap(errcode.MongoError, err.Error())
	}
	return nil
}

// UpdateAppAOMToLTSRule 更新华为云 AOM 到 LTS 接入规则记录
// TODO: 华为云支持通过名称查询日志流以及通过名称查询接入规则后可以去除
func (d *Dao) UpdateAppAOMToLTSRule(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, appID, ruleID string) error {
	filter := bson.M{
		"cluster_name": clusterName,
		"env_name":     envName,
		"app_id":       appID,
	}
	change := bson.M{"$set": bson.M{"rule_id": ruleID}}
	_, err := d.Mongo.Collection(entity.EmptyAppLTSStream.TableName()).UpdateOne(ctx, filter, change)
	if err != nil {
		return errors.Wrap(errcode.MongoError, err.Error())
	}
	return nil
}

// FindSingleAppLTSStream 获取AMS应用对应的华为云 LTS 日志流记录
func (d *Dao) FindSingleAppLTSStream(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, appID string) (*entity.AppLTSStream, error) {
	filter := bson.M{
		"cluster_name": clusterName,
		"env_name":     envName,
		"app_id":       appID,
	}
	stream := new(entity.AppLTSStream)

	err := d.Mongo.ReadOnlyCollection(entity.EmptyAppLTSStream.TableName()).
		FindOne(ctx, filter).
		Decode(stream)
	if err == mongo.ErrNoDocuments {
		return nil, errors.Wrap(errcode.NoRowsFoundError, err.Error())
	}
	if err != nil {
		return nil, errors.Wrap(errcode.MongoError, err.Error())
	}
	return stream, nil
}
