package req

import "rulai/models/entity"

// 配置中心格式化类型
type ConfigManagerFormatType string

const (
	ConfigManagerFormatTypeYaml ConfigManagerFormatType = "yaml"
	ConfigManagerFormatTypeJSON ConfigManagerFormatType = "json"
)

type GetConfigManagerFileReq struct {
	ProjectID   string                  `json:"project_id"`
	ProjectName string                  `json:"project_name"`
	EnvName     entity.AppEnvName       `json:"env_name"`
	CommitID    string                  `json:"commit_id"`
	FormatType  ConfigManagerFormatType `json:"format_type"`
	// 是否解密
	// 默认密码等私有配置加密
	IsDecrypt bool `json:"is_decrypt"`
	// 特殊配置重命名前缀和匹配模式
	ConfigRenamePrefix string                  `json:"config_rename_prefix"`
	ConfigRenameMode   entity.ConfigRenameMode `json:"config_rename_mode"`
}

type GetProjectResourceFromConfigReq struct {
	EnvName  entity.AppEnvName `json:"env_name"`
	CommitID string            `json:"commit_id"`
}
