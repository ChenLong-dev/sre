package req

import (
	"rulai/models"
	"rulai/models/entity"
)

// GetClustersReq 获取集群列表请求参数
type GetClustersReq struct {
	models.BaseListRequest
	Namespace    string               `form:"namespace" json:"namespace"`
	Version      string               `form:"version" json:"version"`
	Names        string               `form:"names" json:"names"`
	ClusterNames []entity.ClusterName `form:"-" json:"-"`
}
