package service

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"

	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/models/req"
	"rulai/models/resp"
)

// GetNodeLabelLists : 获取支持的节点标签列表(用于ams前端选择框显示)
// 不指定类型时返回全部标签列表
func (s *Service) GetNodeLabelLists(ctx context.Context, getReq *req.GetNodeLabelListReq) ([]*resp.NodeLabelListResp, error) {
	// 标签值有限，无需分页
	filter := bson.M{}
	if getReq.Type != "" {
		filter["type"] = getReq.Type
	}

	labels, err := s.dao.FindNodeLabelLists(ctx, filter)
	if err != nil {
		return nil, err
	}

	res := make([]*resp.NodeLabelListResp, 0)
	err = deepcopy.Copy(&labels).To(&res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}
	return res, nil
}
