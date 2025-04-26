package cm

import (
	"fmt"
	"os"
)

const (
	// host
	Host = "config-manager.qingting-hz.com"

	// 默认获取文件api
	DefaultFileApi = "/v1/api/file"
)

var (
	EnvName = []string{"ENV", "env"}
)

// 配置
type Config struct {
	// host
	Host string `yaml:"host"`
	// 文件api
	FileAPI string `yaml:"fileApi"`
}

// 默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:    fmt.Sprintf("http://%s", Host),
		FileAPI: DefaultFileApi,
	}
}

// 获取环境变量
func getEnvWithDefault(defaultValue string) string {
	for _, name := range EnvName {
		value := os.Getenv(name)
		if value != "" {
			return value
		}
	}

	return defaultValue
}
