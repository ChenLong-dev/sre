package req

import "rulai/models"

type CreateFavProjectReq struct {
	ProjectID string `json:"project_id" binding:"required"`
}

type GetFavProjectReq struct {
	models.BaseListRequest
	UserID    string `form:"user_id" json:"user_id"`
	ProjectID string `form:"project_id" json:"project_id"`
}
