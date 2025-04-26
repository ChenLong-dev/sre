package req

import (
	"rulai/models"
	"rulai/models/entity"
)

type UpdateVariableReq struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	OperatorID string
}

type CreateVariableReq struct {
	Key        string              `json:"key" binding:"required,min=1"`
	Value      string              `json:"value" binding:"required,min=1"`
	Type       entity.VariableType `json:"type"`
	ProjectID  string              `json:"project_id"`
	OperatorID string
}

type GetVariablesReq struct {
	models.BaseListRequest
	Type      entity.VariableType `form:"type" binding:"required"`
	ProjectID string              `form:"project_id"`
	Key       string
}
