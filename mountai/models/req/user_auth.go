package req

import (
	"rulai/models"
	"rulai/models/entity"
)

type UserAuthLoginReq struct {
	UserName string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type ValidateHasPermissionReq struct {
	OperateType       entity.OperateType `json:"operate_type"`
	CreateTaskEnvName entity.AppEnvName  `json:"create_task_env_name"`
	CreateTaskAction  entity.TaskAction  `json:"create_task_action"`
	ProjectID         string             `json:"project_id"`
	OperatorID        string             `json:"operator_id"`
}

type GetUsersReq struct {
	models.BaseListRequest
	UserIDs []string `form:"user_ids" json:"user_ids"`
	Keyword string   `form:"k" json:"k"`
}
