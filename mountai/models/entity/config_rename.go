package entity

import "go.mongodb.org/mongo-driver/bson/primitive"

// 特殊配置重命名模式
type ConfigRenameMode int

const (
	ConfigRenameModeExact ConfigRenameMode = iota + 1 // 精确匹配模式
)

// 特殊配置重命名模式名称
const (
	ConfigRenameModeNameExact   = "精确匹配"
	ConfigRenameModeNameUnknown = "未知"
)

// SupportedConfigRenameModes 支持的特殊配置重命名模式列表
var SupportedConfigRenameModes = []ConfigRenameMode{
	ConfigRenameModeExact,
}

// ConfigRenamePrefix 特殊配置重命名前缀
type ConfigRenamePrefix struct {
	ID     primitive.ObjectID `bson:"_id" json:"_id"`
	Prefix string             `bson:"prefix" json:"prefix"`
	Name   string             `bson:"name" json:"name"`
}

func GetConfigRenameModeName(mode ConfigRenameMode) string {
	switch mode {
	case ConfigRenameModeExact:
		return ConfigRenameModeNameExact

	default:
	}

	return ConfigRenameModeNameUnknown
}

func (crp *ConfigRenamePrefix) TableName() string { return "config_rename_prefix" }
