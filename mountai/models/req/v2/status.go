package v2

import (
	"rulai/models/entity"
)

// GetRunningStatusListReq 获取应用运行状态列表请求参数
type GetRunningStatusListReq struct {
	ProjectID   string             `form:"project_id" json:"project_id" binding:"required"`
	AppIDs      string             `form:"app_ids" json:"app_ids" binding:"required"`
	EnvName     entity.AppEnvName  `form:"env_name"  json:"env_name" binding:"required"`
	ClusterName entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
	Namespace   string             `form:"namespace" json:"namespace" binding:"omitempty"`
}
