package req

import (
	"rulai/models/entity"

	"gitlab.shanhai.int/sre/library/base/null"
)

type CreateProjectCIJobReq struct {
	MessageNotification []entity.NotificationType    `json:"message_notification"`
	AllowMergeSwitch    null.Bool                    `json:"allow_merge_switch"`
	PipelineStages      []entity.PipelineStage       `json:"pipeline_stages"`
	DeployBranchName    map[entity.AppEnvName]string `json:"deploy_branch_name"`

	CIJobName string `json:"-"`
	HookURL   string `json:"-"`
	ViewURL   string `json:"-"`
}

type UpdateProjectCIJobReq struct {
	AllowMergeSwitch    null.Bool                    `json:"allow_merge_switch"`
	MessageNotification []entity.NotificationType    `json:"message_notification"`
	PipelineStages      []entity.PipelineStage       `json:"pipeline_stages"`
	DeployBranchName    map[entity.AppEnvName]string `json:"deploy_branch_name"`

	ViewURL string `json:"-"`
	HookURL string `json:"-"`
}

type GetProjectCIJobs struct {
	Limit int `json:"limit" binding:"max=50"`
	Page  int `json:"page" binding:"min=1"`
}
