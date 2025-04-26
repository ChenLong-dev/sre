package entity

import (
	"strings"
	"time"
)

const (
	LanguageGo         = "Go"
	LanguageJavaScript = "JavaScript"
	LanguagePython     = "Python"
	LanguagePHP        = "PHP"
	LanguageOthers     = "Others"
)

// 项目
type Project struct {
	// git的项目id
	ID string `bson:"_id" json:"_id"`
	// 项目名
	Name string `bson:"name" json:"name"`
	// 使用的语言
	Language string `bson:"language" json:"language"`
	// 项目描述
	Desc string `bson:"desc" json:"desc"`
	// api文档地址
	APIDocURL string `bson:"api_doc_url" json:"api_doc_url"`
	// 开发文档地址
	DevDocURL string `bson:"dev_doc_url" json:"dev_doc_url"`
	// 标签
	Labels []string `bson:"labels" json:"labels"`
	// 镜像构建参数
	ImageArgs map[string]string `bson:"image_args" json:"image_args"`
	// 资源可用规格
	ResourceSpec map[AppEnvName]ProjectResourceSpec `bson:"resource_spec" json:"resource_spec"`
	// 团队id
	TeamID string `bson:"team_id" json:"team_id"`
	// apollo appid
	ApolloAppID string `bson:"apollo_appid" json:"apollo_appid"`
	// 项目负责人
	OwnerIDs []string `bson:"owner_ids" json:"owner_ids"`
	// 测试工程师
	QAEngineers []*DingDingUserDetail `bson:"qa_engineers" json:"qa_engineers"`
	// 运维工程师
	OperationEngineers []*DingDingUserDetail `bson:"operation_engineers" json:"operation_engineers"`
	// 产品经理
	ProductManagers []*DingDingUserDetail `bson:"product_managers" json:"product_managers"`
	// 日志仓库
	LogStoreName string `bson:"log_store_name" json:"log_store_name"`
	// 特殊配置重命名前缀
	ConfigRenamePrefixes []string `bson:"config_rename_prefixes" json:"config_rename_prefixes"`
	// 特殊配置重命名模式
	ConfigRenameModes []ConfigRenameMode `bson:"config_rename_modes" json:"config_rename_modes"`
	// 是否启用 Istio
	EnableIstio bool       `bson:"enable_istio" json:"enable_istio"`
	CreateTime  *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime  *time.Time `bson:"update_time" json:"update_time"`
}

func (*Project) TableName() string {
	return "project"
}

func (p *Project) DecodeImageArgs(args map[string]interface{}) map[string]string {
	res := make(map[string]string)
	// mongo不支持 `.` 以及 `$` 等符号，需要特殊处理
	for branch, arg := range p.ImageArgs {
		branch = strings.ReplaceAll(branch, "\\u002e", ".")
		branch = strings.ReplaceAll(branch, "\\u0024", "$")
		branch = strings.ReplaceAll(branch, "\\\\", "\\")
		res[branch] = arg
	}
	return res
}

// 项目资源可用规格
type ProjectResourceSpec struct {
	CPURequestList []CPUResourceType `bson:"cpu_request_list" json:"cpu_request_list"`
	CPULimitList   []CPUResourceType `bson:"cpu_limit_list" json:"cpu_limit_list"`
	MemRequestList []MemResourceType `bson:"mem_request_list" json:"mem_request_list"`
	MemLimitList   []MemResourceType `bson:"mem_limit_list" json:"mem_limit_list"`
}
