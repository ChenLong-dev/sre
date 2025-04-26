package resp

import (
	"rulai/models/entity"
)

// EmptyClusterDetailRespList 供复用的空集群详情列表
var EmptyClusterDetailRespList = make([]*ClusterDetailResp, 0)

// ClusterDetailResp 集群详情返回值
type ClusterDetailResp struct {
	Name          entity.ClusterName `json:"name"`
	IsDefault     bool               `json:"is_default"`
	ServerVersion string             `json:"server_version"`
	Env           entity.AppEnvName  `json:"env"`
}
