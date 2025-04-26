package dao

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/entity"
)

// UpsertUnexpectedImageRecord 创建或更新非预期镜像记录
func (d *Dao) UpsertUnexpectedImageRecord(ctx context.Context, record *entity.UnexpectedImageRecord) error {
	_, err := d.Mongo.Collection(record.TableName()).
		UpdateOne(ctx, d.getUnexpectedImageRecordFilter(record), bson.M{"$set": record}, options.Update().SetUpsert(true))
	if err != nil {
		return errors.Wrap(errcode.MongoError, err.Error())
	}

	return nil
}

// DeleteUnexpectedImageRecord 删除非预期镜像记录
func (d *Dao) DeleteUnexpectedImageRecord(ctx context.Context, record *entity.UnexpectedImageRecord) error {
	_, err := d.Mongo.Collection(entity.EmptyUnexpectedImageRecord.TableName()).DeleteOne(ctx, d.getUnexpectedImageRecordFilter(record))
	if err != nil {
		return errors.Wrap(errcode.MongoError, err.Error())
	}

	return nil
}

// getUnexpectedImageRecordFilter 生成非预期镜像记录过滤条件
// 分三种情形(集群和命名空间是公共筛选条件)
//  1. AMS 项目生成的k8s资源按照项目名称和应用名称唯一确定
//  2. 非 AMS 项目，但由 Deployment/DaemonSet 等上级资源产生的 pod 则按照上级资源唯一确定
//  3. 其他 pod 按照 pod 名称唯一确定
func (d *Dao) getUnexpectedImageRecordFilter(record *entity.UnexpectedImageRecord) bson.M {
	filter := bson.M{"cluster": record.Cluster, "namespace": record.Namespace}
	if record.AMSProjectName != "" {
		filter["ams_project_name"] = record.AMSProjectName
		filter["ams_app_name"] = record.AMSAppName
	} else if record.OwnerReferenceName != "" {
		filter["owner_reference_kind"] = record.OwnerReferenceKind
		filter["owner_reference_name"] = record.OwnerReferenceName
	} else {
		filter["pod_name"] = record.PodName
	}

	return filter
}
