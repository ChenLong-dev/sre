package framework

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	_errors "github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"
	ginUtil "gitlab.shanhai.int/sre/library/net/gin"
	"gitlab.shanhai.int/sre/library/net/metric"
	"gitlab.shanhai.int/sre/library/net/middleware"
	"gitlab.shanhai.int/sre/library/net/response"
	"gitlab.shanhai.int/sre/library/net/sentry"
	"gitlab.shanhai.int/sre/library/net/tracing"
	"gitlab.shanhai.int/sre/library/net/trafficshaping"
	"net/http"
	_ "net/http/pprof"
)

const (
	DefaultTrafficShapingQPS         = 1000
	DefaultTrafficShapingConcurrency = 1000
)

// Http服务器
type HttpServer struct {
	// gin服务引擎
	Engine *gin.Engine
	// gin简单分组
	SimpleRouterGroup *gin.RouterGroup
	// 添加外部路由方法
	Router func(*gin.Engine)
	// 添加外部中间件方法
	Middleware func(*gin.Engine)
	// 没有找到路由的处理方法
	NoRouteHandler func(*gin.Context)

	// http服务器
	server *http.Server
	// 服务器名
	name string
	// 配置文件
	config *Config
}

// 实现ServerInterface
func (svr *HttpServer) ShutDown(ctx context.Context) error {
	server := svr.server
	if server == nil {
		return _errors.New("http: no server")
	}

	return _errors.WithStack(server.Shutdown(ctx))
}

// 实现ServerInterface
func (svr *HttpServer) Start(c *Config, svc ServiceInterface) {
	if c.Gin == nil {
		panic("http config is nil")
	}

	svr.name = "HTTP"
	svr.config = c
	svr.setGinEngine()
	svr.setMiddleware()
	svr.setRouter()
	svr.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", c.Gin.Endpoint.Address, c.Gin.Endpoint.Port),
		Handler: svr.Engine,
	}

	svc.StartServer(svr.name, func(ctx context.Context) error {
		err := svr.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})
}

// 实现ServerInterface
func (svr *HttpServer) Name() string {
	return svr.name
}

// 设置中间件
func (svr *HttpServer) setMiddleware() {
	e := svr.Engine
	s := svr.SimpleRouterGroup

	// Gin日志打印
	e.Use(ginUtil.GetDefaultFormatter(svr.config.Gin))
	s.Use(ginUtil.GetDefaultFormatter(svr.config.Gin))

	// 链路跟踪
	if !svr.config.DisableTracing {
		e.Use(tracing.ExtractFromUpstream())
		e.Use(tracing.InjectToDownstream())

		s.Use(tracing.ExtractFromUpstream())
		s.Use(tracing.InjectToDownstream())
	}

	// 数据统计
	if !svr.config.DisableMetrics {
		e.Use(metric.PrometheusMiddleware())
		s.Use(metric.PrometheusMiddleware())
	}

	// 请求超时
	if svr.config.Gin.Timeout != 0 {
		e.Use(middleware.TimeoutMiddleware(svr.config.Gin.Timeout))
	}

	// 限流
	if !svr.config.DisableTrafficShaping {
		if svr.config.TrafficShapingQPS == 0 {
			svr.config.TrafficShapingQPS = DefaultTrafficShapingQPS
		}
		if svr.config.TrafficShapingConcurrency == 0 {
			svr.config.TrafficShapingConcurrency = DefaultTrafficShapingConcurrency
		}
		e.Use(middleware.TrafficShapingMiddleware([]*trafficshaping.Rule{
			{
				Type:            trafficshaping.QPS,
				ControlBehavior: trafficshaping.Reject,
				Limit:           svr.config.TrafficShapingQPS,
			},
			{
				Type:            trafficshaping.Concurrency,
				ControlBehavior: trafficshaping.Reject,
				Limit:           svr.config.TrafficShapingConcurrency,
			},
		}))
	}

	// 异常捕获
	if !svr.config.DisableCatchPanic {
		e.Use(middleware.CatchPanicMiddleware())
		s.Use(middleware.CatchPanicMiddleware())
	}

	// Sentry异常捕获
	if !svr.config.DisableSentry {
		e.Use(sentry.GinMiddleware(&sentry.GinOption{}))
		s.Use(sentry.GinMiddleware(&sentry.GinOption{}))
	}

	// 自定义中间件
	if svr.Middleware != nil {
		svr.Middleware(e)
	}

	// context变量
	e.Use(middleware.SetDefaultContextValueMiddleware())
	s.Use(middleware.SetDefaultContextValueMiddleware())

	// Sentry系统标签
	if !svr.config.DisableSentry {
		e.Use(sentry.GlobalTagsMiddleware(nil))
		s.Use(sentry.GlobalTagsMiddleware(nil))
	}
}

// 设置路由
func (svr *HttpServer) setRouter() {
	e := svr.Engine
	s := svr.SimpleRouterGroup

	// 没有找到路由
	if svr.NoRouteHandler == nil {
		svr.NoRouteHandler = func(c *gin.Context) {
			response.StandardJSON(c, nil, errcode.NotFound)
		}
	}
	e.NoRoute(svr.NoRouteHandler)

	// 健康检查
	if svr.config.HealthCheckRouter == "" {
		svr.config.HealthCheckRouter = "/health"
	}
	s.Any(svr.config.HealthCheckRouter, func(ctx *gin.Context) {
		response.StandardJSON(ctx, "heath check alive", nil)
	})

	// 数据统计
	if !svr.config.DisableMetrics {
		if svr.config.MetricsRouter == "" {
			svr.config.MetricsRouter = "/metrics"
		}
		s.GET(svr.config.MetricsRouter, metric.GinMetricsHandler)
	}

	// 自定义路由
	if svr.Router != nil {
		svr.Router(e)
	}
}

// 设置默认引擎
func (svr *HttpServer) setGinEngine() {
	binding.Validator = ginUtil.NewV10Validator()

	gin.DefaultWriter = ginUtil.GetInfoWriter(svr.config.Gin)
	gin.DefaultErrorWriter = ginUtil.GetErrorWriter(svr.config.Gin)
	gin.DebugPrintRouteFunc = ginUtil.GetDefaultRouterPrintFunc(svr.config.Gin)
	gin.SetMode(gin.ReleaseMode)

	svr.Engine = gin.New()
	svr.SimpleRouterGroup = svr.Engine.Group("")
}
