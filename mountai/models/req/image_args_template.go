package req

import (
	"rulai/models"
)

type UpdateImageArgsTemplateReq struct {
	Name       string `json:"name"`
	Content    string `json:"content"`
	OperatorID string
}

type CreateImageArgsTemplateReq struct {
	TeamID     string `json:"team_id" binding:"required"`
	Name       string `json:"name"  binding:"required,min=1"`
	Content    string `json:"content"  binding:"required,min=1"`
	OperatorID string
}

type GetImageArgsTemplateReq struct {
	models.BaseListRequest
	TeamID string `form:"team_id" binding:"required"`
	Name   string
}
