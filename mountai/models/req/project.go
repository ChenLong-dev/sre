package req

import (
	"gitlab.shanhai.int/sre/library/base/null"

	"rulai/models"
	"rulai/models/entity"
)

type CreateProjectReq struct {
	Name               string                       `json:"name" binding:"required"`
	Language           string                       `json:"language" binding:"required,oneof=Go JavaScript Python PHP Lua Others"`
	Desc               string                       `json:"desc"`
	GitID              string                       `json:"git_id" binding:"required"`
	APIDocURL          string                       `json:"api_doc_url"`
	DevDocURL          string                       `json:"dev_doc_url"`
	Labels             []string                     `json:"labels"`
	TeamID             string                       `json:"team_id" binding:"required"`
	ApolloAppID        string                       `json:"apollo_appid"`
	DisableCI          bool                         `json:"disable_ci"`
	EnableIstio        null.Bool                    `json:"enable_istio"`
	OwnerIDs           []string                     `json:"owner_ids"`
	QAEngineers        []*entity.DingDingUserDetail `json:"qa_engineers"`
	OperationEngineers []*entity.DingDingUserDetail `json:"operation_engineers"`
	ProductManagers    []*entity.DingDingUserDetail `json:"product_managers"`
}

type UpdateProjectReq struct {
	Language           string                       `json:"language" binding:"omitempty,oneof=Go JavaScript Python PHP Lua Others"`
	Desc               string                       `json:"desc"`
	APIDocURL          string                       `json:"api_doc_url"`
	DevDocURL          string                       `json:"dev_doc_url"`
	Labels             []string                     `json:"labels"`
	ImageArgs          map[string]string            `json:"image_args"`
	TeamID             string                       `json:"team_id"`
	ApolloAppID        string                       `json:"apollo_appid"`
	LogStoreName       string                       `json:"log_store_name"`
	OwnerIDs           []string                     `json:"owner_ids"`
	QAEngineers        []*entity.DingDingUserDetail `json:"qa_engineers"`
	OperationEngineers []*entity.DingDingUserDetail `json:"operation_engineers"`
	ProductManagers    []*entity.DingDingUserDetail `json:"product_managers"`
	EnableIstio        null.Bool                    `json:"enable_istio"`
}

type GetProjectsReq struct {
	models.BaseListRequest
	Keyword      string   `form:"keyword" json:"keyword"`
	Name         string   `form:"name" json:"name"`
	Language     string   `form:"language" json:"language"`
	TeamID       string   `form:"team_id" json:"team_id"`
	OwnerID      string   `form:"owner_id" json:"owner_id"`
	Labels       string   `form:"labels" json:"labels"`
	ProjectIDs   string   `form:"project_ids" json:"project_ids"`
	IDs          []string `form:"ids" json:"ids"`
	KeywordField string   `form:"keyword_field" json:"keyword_field"`
}

type GetProjectConfigReq struct {
	EnvName    entity.AppEnvName       `form:"env_name" json:"env_name" binding:"required"`
	FormatType ConfigManagerFormatType `form:"format_type" json:"format_type"`
}

type GetProjectResourceReq struct {
	EnvName entity.AppEnvName `form:"env_name" json:"env_name" binding:"required"`
}

// GetProjectAppsClustersWithWorkloadReq 批量获取项目下多个应用各自在指定环境下有工作负载的集群列表请求参数
type GetProjectAppsClustersWithWorkloadReq struct {
	EnvName entity.AppEnvName `form:"env_name" json:"env_name" binding:"required"`
	AppIDs  string            `form:"app_ids" json:"app_ids" binding:"required"`
}
