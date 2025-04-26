package test

import (
	"gopkg.in/yaml.v3"

	"io/ioutil"
	"os"
)

// 单测配置文件目录
const UnitConfigPath = "./internal/test/unit.yaml"

// 单测代理配置
type UnitProxyConfig struct {
	// 监听地址
	Listen string `yaml:"listen"`
	// 上游地址
	Upstream string `yaml:"upstream"`
	// 是否开启
	Enabled bool `yaml:"enabled"`
}

// 单测配置
type UnitConfig struct {
	// 代理配置
	Proxy map[string]UnitProxyConfig `yaml:"proxy"`
	// queue包配置
	Queue struct {
		// 地址
		Address string `yaml:"address"`
		// 端口
		Port int `yaml:"port"`
		// 用户名
		UserName string `yaml:"userName"`
		// 密码
		Password string `yaml:"password"`
		// vhost
		VHost string `yaml:"vHost"`
	} `yaml:"queue"`
}

// 从本地解析单测配置文件
func DecodeUnitConfigFromLocal(configPath string) *UnitConfig {
	configFile, err := os.Open(configPath)
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	configData, err := ioutil.ReadAll(configFile)
	if err != nil {
		panic(err)
	}

	cfg := new(UnitConfig)
	err = yaml.Unmarshal(configData, cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
