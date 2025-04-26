package resp

import "rulai/models/entity"

type CIJobDetailResp struct {
	ID                  string                       `json:"id" deepcopy:"method:GenerateObjectIDString"`
	ProjectID           string                       `json:"project_id"`
	Name                string                       `json:"name"`
	ViewURL             string                       `json:"view_url"`
	HookURL             string                       `json:"hook_url"`
	MessageNotification []entity.NotificationType    `json:"message_notification"`
	AllowMergeSwitch    bool                         `json:"allow_merge_switch"`
	DeployBranchName    map[entity.AppEnvName]string `json:"deploy_branch_name"`
	PipelineStages      []entity.PipelineStage       `json:"pipeline_stages"`
}
