package req

import (
	"rulai/models/entity"
)

// CreatePProfReq 创建 pprof 监控信息请求参数
type CreatePProfReq struct {
	AppID       string             `form:"app_id" json:"app_id" binding:"required"`
	EnvName     entity.AppEnvName  `form:"env_name"  json:"env_name" binding:"required"`
	ClusterName entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
	Type        entity.PProfType   `form:"type" json:"type" binding:"required"`
	Action      entity.PProfAction `form:"action" json:"action" binding:"required"`
	Seconds     int                `form:"seconds" json:"seconds" binding:"lt=60"`
	PodPort     int                `form:"pod_port" json:"pod_port"`

	GenerateFilePath string `form:"-" json:"-"`
	SourceFilePath   string `form:"-" json:"-"`
	PodName          string `form:"-" json:"-"`
	Namespace        string `form:"namespace" json:"namespace"`
	Container        string `form:"version" json:"version"`
}
