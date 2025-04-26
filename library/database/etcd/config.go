package etcd

import (
	"gitlab.shanhai.int/sre/library/base/ctime"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

// 配置文件
type Config struct {
	// 服务器端点
	Endpoints []*EndpointConfig `yaml:"endpoints"`
	// 用户名
	UserName string `yaml:"userName"`
	// 密码
	Password string `yaml:"password"`
	// 连接超时时间
	DialTimeout ctime.Duration `yaml:"dialTimeout"`
	// 是否打印数据的具体值
	DataValueOut bool `yaml:"dataValueOut"`

	// 预加载配置
	Preload []*PreloadPrefixConfig `yaml:"preload"`

	// TLS加密配置
	Tls *TlsConfig `yaml:"tls"`
	// 日志配置
	*render.Config `yaml:",inline"`
}

// 端点配置
type EndpointConfig struct {
	// 地址
	Address string `yaml:"address"`
	// 端口
	Port int `yaml:"port"`
}

// 预加载配置
type PreloadPrefixConfig struct {
	// 是否开启实时监听
	EnableWatch bool `yaml:"enableWatch"`
	// 前缀
	Prefix string `yaml:"prefix"`
	// 值过滤数组
	ValueFilter []string `yaml:"valueFilter"`
}

// TLS配置文件
type TlsConfig struct {
	// 是否开启TLS
	Enable bool `yaml:"enable"`
	// Cert文件路径
	CertFilePath string `yaml:"certFilePath"`
	// Key文件路径
	KeyFilePath string `yaml:"keyFilePath"`
	// CA文件路径
	TrustedCAFilePath string `yaml:"trustedCAFilePath"`
}
