package agollo

import (
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

type Config struct {
	// Apollo的appID
	AppID string `yaml:"appID"`
	// Apollo的cluster
	Cluster string `yaml:"cluster"`
	// Apollo的namespace
	PreloadNamespaces []string `yaml:"namespaceNames"`
	// Apollo的监听Ip
	ServerHost string `yaml:"serverHost"`
	// 设置false实时获取最新配置，设置true需手动watch
	NotDaemon bool `yaml:"notDaemon"`
	// 日志配置
	*render.Config `yaml:",inline"`
}

// 打印日志信息
type logInfo struct {
	AppID         string
	Cluster       string
	NamespaceName string
	ServerHost    string
	extra         string
	source        string
	changes       string
	watcherType   WatcherType
}

// 校验参数
func configIdentify(app *Config) {
	if app == nil {
		panic("Apollo config is nil")
	}
	if app.ServerHost == "" {
		panic("Apollo serverHost must be set")
	}
	if app.Cluster == "" {
		panic("Apollo cluster must be set")
	}
}
