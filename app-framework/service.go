package framework

import (
	"context"
	"gitlab.shanhai.int/sre/library/goroutine"
)

// 服务层接口
type ServiceInterface interface {
	// 启动服务器
	// 该方法在服务层接口内的原因是便于启动多个服务器时的管理
	// 该方法实现应异步启动
	StartServer(serverName string, startFunc func(ctx context.Context) error)
	// 关闭服务
	Close(ctx context.Context)
	// 服务器运行时错误
	// 该方法实现应通过异步管道发送，以避免外部无法手动手动停止服务
	Error() <-chan error
}

// 默认服务结构体
type DefaultService struct {
	// 协程组
	eg *goroutine.ErrGroup
	// 数据层接口
	dao DaoInterface
	// 配置文件
	config *Config
}

// 实现ServiceInterface
func (s *DefaultService) StartServer(serverName string, startFunc func(ctx context.Context) error) {
	s.eg.Go(context.Background(), serverName, startFunc)
}

// 实现ServiceInterface
func (s *DefaultService) Close(ctx context.Context) {
	s.dao.Close(ctx)
}

// 实现ServiceInterface
func (s *DefaultService) Error() <-chan error {
	errChan := make(chan error, 1)
	go func() {
		err := s.eg.Wait()
		if err != nil {
			errChan <- err
		}
		// 全部结束，关闭管道
		close(errChan)
	}()
	return errChan
}

// 获取默认服务
func GetDefaultService(conf *Config, dao DaoInterface) *DefaultService {
	s := DefaultService{
		config: conf,
		dao:    dao,
		eg:     goroutine.New(conf.AppName),
	}

	return &s
}
