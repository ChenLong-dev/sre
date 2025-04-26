package resp

import "time"

// jenkins任务的结果
type JenkinsJobResult string

const (
	// 成功
	JenkinsJobResultSuccess JenkinsJobResult = "SUCCESS"
	// 失败
	JenkinsJobResultFailure JenkinsJobResult = "FAILURE"
	// 终止
	JenkinsJobResultAborted JenkinsJobResult = "ABORTED"
	// 运行中
	JenkinsJobResultRunning JenkinsJobResult = ""
)

type ImageListResp struct {
	BuildID             string           `json:"build_id"`
	JobURL              string           `json:"job_url"`
	Status              JenkinsJobResult `json:"status"`
	BranchName          string           `json:"branch_name"`
	BuildArgsTemplateID string           `json:"build_args_template_id"`
	BuildArgWithMask    string           `json:"build_args_with_mask"`
	ImageTag            string           `json:"image_tag"`
	CommitID            string           `json:"commit_id"`
	LastComment         string           `json:"last_comment"`
	CreateTime          string           `json:"create_time"`
	ImageRepoURL        string           `json:"image_repo_url"`
	Description         string           `json:"description"`
	UserProfile         *UserProfileResp `json:"user_profile"`
	Duration            string           `json:"duration"`
}

type ImageDetailResp struct {
	BuildID             string           `json:"build_id"`
	JobName             string           `json:"job_name"`
	JobURL              string           `json:"job_url"`
	Status              JenkinsJobResult `json:"status"`
	ImageRepoURL        string           `json:"image_repo_url"`
	BuildArgsTemplateID string           `json:"build_args_template_id"`
	BuildArgWithMask    string           `json:"build_args_with_mask"`
	BranchName          string           `json:"branch_name"`
	ImageTag            string           `json:"image_tag"`
	CommitID            string           `json:"commit_id"`
	ConsoleOutput       string           `json:"console_output"`
	Description         string           `json:"description"`
	UserProfile         *UserProfileResp `json:"user_profile"`
	Duration            string           `json:"duration"`
	// CreateTime is primary for frontend display
	CreateTime string `json:"create_time"`

	// Timestamp is for storing in db
	Timestamp time.Time `json:"-"`
	BuildArg  string    `json:"-"`
}

type ImageTagResp struct {
	BranchName          string             `json:"branch_name"`
	CommitID            string             `json:"commit_id"`
	BuildArgsTemplateID string             `json:"build_args_template_id"`
	BuildArgWithMask    string             `json:"build_args_with_mask"`
	ImageTag            string             `json:"image_tag"`
	CreateTime          string             `json:"create_time"`
	UpdateTime          string             `json:"update_time"`
	Version             string             `json:"version"`
	Description         string             `json:"description"`
	Template            *ImageArgsTemplate `json:"template"`
}

type ImageLastArgsResp struct {
	BranchName          string `json:"branch_name"`
	CommitID            string `json:"commit_id"`
	BuildArgsTemplateID string `json:"build_args_template_id"`
	BuildArgWithMask    string `json:"build_args_with_mask"`
	ImageTag            string `json:"image_tag"`
	CreateTime          string `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime          string `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	Description         string `json:"description"`
}
