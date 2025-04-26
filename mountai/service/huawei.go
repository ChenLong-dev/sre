package service

import (
	"context"
	"fmt"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
)

func (s *Service) HuaWeiGetRepoTags(ctx context.Context, getReq *req.GetRepoTagsReq) ([]*resp.GetDockerTagsResp, int, error) {
	vendor, err := s.getVendorController(entity.VendorHuawei)
	if err != nil {
		return nil, 0, err
	}

	return vendor.GetRepoTags(ctx, getReq)
}

// 获取镜像仓库地址
func (s *Service) GetHuaWeiImageRepoURL(version string) string {
	return fmt.Sprintf("crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/shanhaii/%s", version)
}
