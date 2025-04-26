package req

import (
	"rulai/models"

	"gitlab.shanhai.int/sre/library/base/null"
)

type GetImageJobsReq struct {
	models.BaseListRequest
	BranchName string      `form:"branch_name" json:"branch_name"`
	Status     null.String `form:"status" json:"status"`

	ProjectID   string `form:"-" json:"-"`
	ProjectName string `form:"-" json:"-"`
}

type GetImageTagsReq struct {
	models.BaseListRequest

	ProjectName string `form:"-" json:"-"`
}

type GetImageJobDetailReq struct {
	BuildID     string `form:"-" json:"-"`
	ProjectName string `form:"-" json:"-"`
}

type GetImageBuildLogReq struct {
	BuildID     string `form:"-" json:"-"`
	ProjectName string `form:"-" json:"-"`
}

type CreateImageJobReq struct {
	BuildArg            string `json:"build_arg"`
	BuildArgsTemplateID string `json:"build_args_template_id"`
	BuildArgWithMask    string
	BranchName          string `json:"branch_name" binding:"required"`
	CommitID            string `json:"commit_id" binding:"required"`
	UserID              string `json:"user_id"`
	SyncToken           string `json:"-"`
	Description         string `json:"description"`
	Timeout             int    `json:"timeout"`
}

type DeleteImageJobReq struct {
	BuildID     string `form:"-" json:"-"`
	ProjectName string `form:"-" json:"-"`
}

type GetImageBuildReq struct {
	BranchName  string `json:"branch_name"`
	CommitID    string `json:"commit_id"`
	ProjectName string `json:"project_name"`
	ImageTag    string `json:"image_tag"`
}
