package resp

import "rulai/models/entity"

// ConfigRenamePrefixDetail 允许使用的特殊配置重命名前缀详情
type ConfigRenamePrefixDetail struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

// ConfigRenameModeDetail 允许使用的特殊配置重命名模式详情
type ConfigRenameModeDetail struct {
	Enum entity.ConfigRenameMode `json:"enum"`
	Name string                  `json:"name"`
}
