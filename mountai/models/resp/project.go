package resp

import "rulai/models/entity"

type ProjectDetailResp struct {
	ID                   string                                        `json:"id"`
	Name                 string                                        `json:"name"`
	Language             string                                        `json:"language"`
	Desc                 string                                        `json:"desc"`
	APIDocURL            string                                        `json:"api_doc_url"`
	DevDocURL            string                                        `json:"dev_doc_url"`
	Labels               []string                                      `json:"labels"`
	ImageArgs            map[string]string                             `json:"image_args" deepcopy:"method:DecodeImageArgs"`
	ResourceSpec         map[entity.AppEnvName]ProjectResourceSpecResp `json:"resource_spec"`
	Team                 *TeamDetailResp                               `json:"team"`
	LogStoreName         string                                        `json:"log_store_name"`
	IsFav                bool                                          `json:"is_fav"`
	Owners               []*UserProfileResp                            `json:"owners"`
	QAEngineers          []*entity.DingDingUserDetail                  `json:"qa_engineers"`
	OperationEngineers   []*entity.DingDingUserDetail                  `json:"operation_engineers"`
	ProductManagers      []*entity.DingDingUserDetail                  `json:"product_managers"`
	ConfigRenamePrefixes []*ConfigRenamePrefixDetail                   `json:"config_rename_prefixes"`
	ConfigRenameModes    []*ConfigRenameModeDetail                     `json:"config_rename_modes"`
	EnableIstio          bool                                          `json:"enable_istio"`
	CreateTime           string                                        `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime           string                                        `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
}

// ProjectResourceSpecResp 项目资源可用规格
type ProjectResourceSpecResp struct {
	CPURequestList []entity.CPUResourceType `json:"cpu_request_list"`
	CPULimitList   []entity.CPUResourceType `json:"cpu_limit_list"`
	MemRequestList []entity.MemResourceType `json:"mem_request_list"`
	MemLimitList   []entity.MemResourceType `json:"mem_limit_list"`
}

// ProjectListResp 项目列表精简反馈
type ProjectListResp struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Language   string             `json:"language"`
	Desc       string             `json:"desc"`
	Labels     []string           `json:"labels"`
	Team       *TeamListResp      `json:"team"`
	Owners     []*UserProfileResp `json:"owners"`
	IsFav      bool               `json:"is_fav"`
	CreateTime string             `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime string             `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
}

type ActiveProjectResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`

	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`

	TaskCreateTime string `json:"task_create_time"`
}

type ProjectResourceResp struct {
	Library *GetQTFrameworkVersionResp `json:"library"`
	*GetProjectResourceFromConfigResp
}

type ProjectMemberRoleResp struct {
	AccessLevel entity.GitMemberAccess `json:"access_level"`
	Role        string                 `json:"role"`
}

// ProjectAppsClustersWithWorkloadResp 批量获取项目下多个应用各自在指定环境下有工作负载的集群列表响应参数
type ProjectAppsClustersWithWorkloadResp struct {
	AppID    string               `json:"app_id"`
	Clusters []*ClusterDetailResp `json:"clusters"`
}
