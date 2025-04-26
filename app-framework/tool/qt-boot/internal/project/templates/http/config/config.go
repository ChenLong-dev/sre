package config

import (
	framework "gitlab.shanhai.int/sre/app-framework"
	"gitlab.shanhai.int/sre/library/net/httpclient"
)

var (
	// ====================
	// >>>请勿删除<<<
	//
	// 全局配置文件
	// ====================
	Conf *Config
)

// ====================
// >>>请勿删除<<<
//
// 配置文件
// ====================
type Config struct {
	// ====================
	// >>>请勿删除<<<
	//
	// 基础配置文件
	// ====================
	*framework.Config `yaml:",inline"`

	// HttpClient配置文件
	HttpClient *httpclient.Config `yaml:"httpClient"`
}

// ====================
// >>>请勿删除<<<
//
// 读取配置文件
//
//	configPath为配置文件的所属位置，例如 ./config/config.yaml
//
// ====================
func Read(configPath string) *Config {
	Conf = new(Config)

	// 解码yaml配置文件
	framework.DecodeConfig(configPath, Conf)

	return Conf
}
