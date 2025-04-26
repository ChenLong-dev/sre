package req

type GetRepoBranchesReq struct {
	Page     int    `form:"page" binding:"gte=0"`
	PageSize int    `form:"limit" binding:"gte=0,lte=100"`
	Keyword  string `form:"keyword"`
}

type GetProjectMembersReq struct {
	Page     int `form:"page" binding:"gte=0"`
	PageSize int `form:"limit" binding:"gte=0,lte=100"`
}

// 添加gitlab钩子参数
type GitlabProjectHookDetailReq struct {
	URL                   string `json:"url"`
	PushEvents            bool   `json:"push_events"`
	MergeRequestsEvents   bool   `json:"merge_requests_events"`
	EnableSSLVerification bool   `json:"enable_ssl_verification"`
	// Gitlab hook secret token for webHook trigger jenkins pipeline
	Token string `json:"token"`
}

type AddGitlabProjectHookReq struct {
	ProjectID  string                      `json:"project_id"`
	HookDetail *GitlabProjectHookDetailReq `json:"hook_detail"`
}

type ListGitlabProjectHooksReq struct {
	Page     int `form:"page" binding:"gte=0"`
	PageSize int `form:"limit" binding:"gte=0,lte=100"`
}

type EditGitlabProjectReq struct {
	OnlyAllowMergeIfPipelineSucceeds bool `json:"only_allow_merge_if_pipeline_succeeds"`
}
