package entity

// Jenkins配置模版
type JenkinsConfigTemplate struct {
	// 项目的ssh地址
	ProjectSSHUrl string
	// 打包超时时间
	Timeout int
}

// Jenkins CI流程job配置模版
type JenkinsCIConfigTemplate struct {
	// 项目ID
	ProjectID string
	// Gitlab secret token
	GitlabSecretToken string
	// Pipeline branch
	PipelineBranch string
	// Script path
	ScriptPath string
	// Gitlab connection
	GitLabConnection string
	// Pipeline url
	PipelineURL string
	// Gitlab credential id for pipeline
	PipelineCredentialsID string
}
