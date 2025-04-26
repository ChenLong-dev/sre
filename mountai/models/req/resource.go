package req

import (
	"rulai/models/entity"
)

type GetResourcesReq struct {
	Type     string `form:"type" json:"type" binding:"required,oneof=ecs cdn redis rds mongo hbase"`
	Provider string `form:"provider" json:"provider" binding:"omitempty,oneof=aliyun"`
}

type UpdateProjectResourcesReq struct {
	entity.ResourceList
	EnvName entity.AppEnvName `form:"env_name" json:"env_name" binding:"required"`
}
