package resp

import "time"

// 响应模型
type GithubOwnerResp struct {
	ID              int    `json:"id"`
	Login           string `json:"login"`
	Name            string `json:"name"`
	Location        string `json:"location,omitempty"`
	CreateAt        string `json:"created_at,omitempty"`
	UpdateAt        string `json:"updated_at,omitempty"`
	FollowersSize   int    `json:"followers,omitempty"`
	FollowingSize   int    `json:"following,omitempty"`
	PublicReposSize int    `json:"public_repos,omitempty"`
}

// 响应模型
type GithubRepositoryResp struct {
	ID        int                        `json:"id"`
	Name      string                     `json:"name"`
	Private   bool                       `json:"private"`
	CreatedAt *time.Time                 `json:"created_at"`
	Owner     *GithubRepositoryOwnerResp `json:"owner"`
}

// 响应模型
type GithubRepositoryOwnerResp struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}

// 响应模型
type GithubRepositoryAggregationResp struct {
	Owner *GithubOwnerResp        `json:"owner"`
	Repos []*GithubRepositoryResp `json:"repos"`
}
