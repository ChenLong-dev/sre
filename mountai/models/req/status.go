package req

import (
	"rulai/models/entity"
)

// GetRunningStatusListReq 获取应用运行状态列表请求参数
type GetRunningStatusListReq struct {
	AppID       string             `form:"app_id" json:"app_id" binding:"required"`
	EnvName     entity.AppEnvName  `form:"env_name"  json:"env_name" binding:"required"`
	ClusterName entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
	Namespace   string             `form:"namespace" json:"namespace"`
}

// GetRunningStatusDetailReq 获取应用运行状态详情请求参数
type GetRunningStatusDetailReq struct {
	AppID       string             `form:"app_id" json:"app_id" binding:"required"`
	EnvName     entity.AppEnvName  `form:"env_name"  json:"env_name" binding:"required"`
	ClusterName entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
	Version     string             `form:"-" json:"-"`
	Namespace   string             `form:"namespace" json:"namespace"`
}

// GetRunningPodLogsReq 获取正在运行中的 pod 日志请求参数
type GetRunningPodLogsReq struct {
	EnvName       entity.AppEnvName  `form:"env_name"  json:"env_name" binding:"required"`
	ClusterName   entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
	Namespace     string             `form:"namespace" json:"namespace"`
	ContainerName string             `form:"container_name" json:"container_name" binding:"required"`
}

// GetRunningStatusDescriptionReq 获取应用运行状态信息请求参数
type GetRunningStatusDescriptionReq struct {
	AppID       string             `form:"app_id" json:"app_id" binding:"required"`
	EnvName     entity.AppEnvName  `form:"env_name"  json:"env_name" binding:"required"`
	ClusterName entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
	Namespace   string             `form:"namespace" json:"namespace"`
}
