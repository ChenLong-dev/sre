package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"rulai/models/entity"
	"rulai/models/req"
)

// CreateConfigRenamePrefix 创建特殊配置重命名前缀
func (s *Service) CreateConfigRenamePrefix(ctx context.Context,
	createReq *req.CreateConfigRenamePrefixReq) error {
	prefix := &entity.ConfigRenamePrefix{
		ID:     primitive.NewObjectID(),
		Name:   createReq.Name,
		Prefix: createReq.Prefix,
	}

	return s.dao.CreateConfigRenamePrefix(ctx, prefix)
}

// DeleteConfigRenamePrefix 删除特殊配置重命名前缀
func (s *Service) DeleteConfigRenamePrefix(ctx context.Context, prefix string) error {
	return s.dao.DeleteConfigRenamePrefix(ctx, prefix)
}
