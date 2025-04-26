package req

import "rulai/models/entity"

type QueryPrometheusReq struct {
	EnvName entity.AppEnvName `form:"env_name"  json:"env_name" binding:"required"`
	SQL     string            `form:"sql"  json:"sql" binding:"required"`
}

type GetMaxTotalCPUReq struct {
	EnvName   entity.AppEnvName `form:"env_name"  json:"env_name" binding:"required"`
	CountTime string            `form:"count_time"  json:"count_time"`
}

type GetMinTotalCPUReq struct {
	EnvName   entity.AppEnvName `form:"env_name"  json:"env_name" binding:"required"`
	CountTime string            `form:"count_time"  json:"count_time"`
}

type GetMaxTotalMemReq struct {
	EnvName   entity.AppEnvName `form:"env_name"  json:"env_name" binding:"required"`
	CountTime string            `form:"count_time"  json:"count_time"`
}

type GetMinTotalMemReq struct {
	EnvName   entity.AppEnvName `form:"env_name"  json:"env_name" binding:"required"`
	CountTime string            `form:"count_time"  json:"count_time"`
}

type GetWastedMaxCPUUsageRateReq struct {
	EnvName   entity.AppEnvName `form:"env_name"  json:"env_name" binding:"required"`
	CountTime string            `form:"count_time"  json:"count_time"`
}

type GetWastedMaxMemUsageRateReq struct {
	EnvName   entity.AppEnvName `form:"env_name"  json:"env_name" binding:"required"`
	CountTime string            `form:"count_time"  json:"count_time"`
}
