package framework

import (
	"context"
	"github.com/pkg/errors"
)

// 任务服务器
type JobServer struct {
	// 任务名
	name string
	// 配置文件
	config *Config
	// 任务函数
	job func(ctx context.Context) error
}

// 实现ServerInterface
func (svr *JobServer) ShutDown(ctx context.Context) (err error) {
	return nil
}

// 实现ServerInterface
func (svr *JobServer) Start(c *Config, svc ServiceInterface) {
	svr.config = c
	if svr.job == nil {
		panic(errors.New("cron job is nil"))
	}

	svc.StartServer(svr.name, func(ctx context.Context) error {
		return svr.job(ctx)
	})
}

// 实现ServerInterface
func (svr *JobServer) Name() string {
	return svr.name
}

// 设置任务函数
func (svr *JobServer) SetJob(name string, job func(ctx context.Context) error) {
	svr.name = name
	svr.job = job
}
