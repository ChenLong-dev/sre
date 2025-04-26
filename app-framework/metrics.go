package framework

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	_errors "github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/metric"
	"net/http"
	_ "net/http/pprof"
)

const (
	// 规定metric端口，以便采集
	DefaultMetricsPort = 8088
)

// 任务服务器
type MetricsServer struct {
	// 配置文件
	config *Config

	// gin服务引擎
	Engine *gin.Engine

	// http服务器
	server *http.Server
}

// 实现ServerInterface
func (svr *MetricsServer) ShutDown(ctx context.Context) (err error) {
	server := svr.server
	if server == nil {
		return _errors.New("metrics: no server")
	}

	return _errors.WithStack(server.Shutdown(ctx))
}

// 实现ServerInterface
func (svr *MetricsServer) Start(c *Config, svc ServiceInterface) {
	svr.config = c

	gin.SetMode(gin.ReleaseMode)
	svr.Engine = gin.New()
	if svr.config.MetricsRouter == "" {
		svr.config.MetricsRouter = "/metrics"
	}
	svr.Engine.GET(svr.config.MetricsRouter, metric.GinMetricsHandler)

	port := svr.config.MetricsPort
	if port == 0 {
		port = DefaultMetricsPort
	}
	svr.server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: svr.Engine,
	}

	// 单独开协程，避免阻塞其他server
	go func() {
		err := svr.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(_errors.Errorf("metrics server error:%s", err))
		}
	}()
}

// 实现ServerInterface
func (svr *MetricsServer) Name() string {
	return "metrics"
}
