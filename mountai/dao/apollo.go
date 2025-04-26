package dao

import (
	"context"

	"rulai/models/entity"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

func (d *Dao) GetApolloNamespaceByAppID(ctx context.Context, appID string,
	env entity.AppEnvName) (namespaces []*entity.ApolloNamespace, err error) {
	db := d.ApolloStgMysql
	if env == entity.AppEnvPrd {
		db = d.ApolloPrdMysql
	}

	err = db.ReadOnlyTable(ctx, new(entity.ApolloNamespace).TableName()).
		Where("AppId = ?", appID).
		Where("IsDeleted = 0").
		Find(&namespaces).
		Error
	if err != nil {
		return nil, errors.Wrapf(errcode.MysqlError, "%s", err)
	}

	return namespaces, nil
}
