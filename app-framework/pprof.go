package framework

import (
	"context"
	"errors"
	"fmt"
	_errors "github.com/pkg/errors"
	"net/http"
	_ "net/http/pprof"
)

const (
	// 规定pprof端口，以便采集
	DefaultPProfPort = 8089
)

// 任务服务器
type PProfServer struct {
	// 配置文件
	config *Config

	// http服务器
	server *http.Server
}

// 实现ServerInterface
func (svr *PProfServer) ShutDown(ctx context.Context) (err error) {
	server := svr.server
	if server == nil {
		return _errors.New("pprof: no server")
	}

	return _errors.WithStack(server.Shutdown(ctx))
}

// 实现ServerInterface
func (svr *PProfServer) Start(c *Config, svc ServiceInterface) {
	svr.config = c

	port := svr.config.PProfPort
	if port == 0 {
		port = DefaultPProfPort
	}
	svr.server = &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%d", port),
	}

	// 单独开协程，避免阻塞其他server
	go func() {
		err := svr.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(_errors.Errorf("pprof server error:%s", err))
		}
	}()
}

// 实现ServerInterface
func (svr *PProfServer) Name() string {
	return "pprof"
}
