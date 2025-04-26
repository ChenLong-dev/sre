package resp

import (
	"rulai/models/entity"
	"time"
)

type GitAuthTokenResp struct {
	Error     string `json:"error"`
	ErrorDesc string `json:"error_description"`

	AccessToken  string `json:"access_token"`
	CreatedAt    int    `json:"created_at"`
	RefreshToken string `json:"refresh_token"`
}

type GitUserProfileResp struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	UserName  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	State     string `json:"state"`
	Email     string `json:"email"`
}

type GitProjectResp struct {
	ID                               int    `json:"id"`
	Name                             string `json:"name"`
	Path                             string `json:"path"`
	Desc                             string `json:"description"`
	SSHURL                           string `json:"ssh_url_to_repo"`
	HTTPURL                          string `json:"http_url_to_repo"`
	WebURL                           string `json:"web_url"`
	OnlyAllowMergeIfPipelineSucceeds bool   `json:"only_allow_merge_if_pipeline_succeeds"`
}

type GitBranchResp struct {
	Name   string `json:"name"`
	Commit struct {
		ID             string `json:"id"`
		ShortID        string `json:"short_id"`
		Title          string `json:"title"`
		Message        string `json:"message"`
		AuthorEmail    string `json:"author_email"`
		AuthorName     string `json:"author_name"`
		AuthoredDate   string `json:"authored_date"`
		CommitterEmail string `json:"committer_email"`
		CommitterName  string `json:"committer_name"`
		CommittedDate  string `json:"committed_date"`
	} `json:"commit"`
}

type GetQTFrameworkVersionResp struct {
	FrameworkVersion string `json:"framework_version"`
	LibraryVersion   string `json:"library_version"`
}

// GitUserResp git用户信息(仅保留需要字段)
type GitUserResp struct {
	ID        int                   `json:"id"`
	Name      string                `json:"name"`
	Username  string                `json:"username"`
	State     entity.GitMemberState `json:"state"`
	AvatarURL string                `json:"avatar_url"`
	WebURL    string                `json:"web_url"`
	Email     string                `json:"email"`
}

// GitProjectMemberResp 获取git项目成员接口response
type GitProjectMemberResp struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	Username    string                 `json:"username"`
	State       entity.GitMemberState  `json:"state"`
	AvatarURL   string                 `json:"avatar_url"`
	WebURL      string                 `json:"web_url"`
	AccessLevel entity.GitMemberAccess `json:"access_level"`
}

type ListGitlabProjectHooksResp struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	ProjectID int    `json:"project_id"`
}

type GitCommitResp struct {
	ID        string          `json:"id"`
	Author    GitCommitAuthor `json:"author"`
	Title     string          `json:"title"`
	CreatedAt string          `json:"created_at"`
}

type GitCommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GitTagResp struct {
	Commit struct {
		Id             string    `json:"id"`
		ShortId        string    `json:"short_id"`
		Title          string    `json:"title"`
		CreatedAt      time.Time `json:"created_at"`
		ParentIds      []string  `json:"parent_ids"`
		Message        string    `json:"message"`
		AuthorName     string    `json:"author_name"`
		AuthorEmail    string    `json:"author_email"`
		AuthoredDate   time.Time `json:"authored_date"`
		CommitterName  string    `json:"committer_name"`
		CommitterEmail string    `json:"committer_email"`
		CommittedDate  time.Time `json:"committed_date"`
	} `json:"commit"`
	Release struct {
		TagName     string `json:"tag_name"`
		Description string `json:"description"`
	} `json:"release"`
	Name      string      `json:"name"`
	Target    string      `json:"target"`
	Message   interface{} `json:"message"`
	Protected bool        `json:"protected"`
}
