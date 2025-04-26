package framework

import "context"

// 服务器接口
type ServerInterface interface {
	// 启动服务器
	Start(*Config, ServiceInterface)
	// 服务器名称
	Name() string
	// 关闭服务器
	ShutDown(context.Context) (err error)
}
