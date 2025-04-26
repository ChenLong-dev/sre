package framework

import (
	"context"
	"flag"
	"gitlab.shanhai.int/sre/library/base/ctime"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/metric"
	"gitlab.shanhai.int/sre/library/net/sentry"
	"gitlab.shanhai.int/sre/library/net/tracing"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 启动服务
//
//	conf 配置文件
//	svc	服务接口
//	svrSlice 需要启动的服务器接口切片
func Run(conf *Config, svc ServiceInterface, svrSlice ...ServerInterface) {
	flag.Parse()

	// 初始化日志
	log.Init(conf.Log)
	defer log.Close()
	log.Infoc(context.Background(), "Env:%s AppName:%s ProjectName:%s ProjectID:%s  start",
		conf.Env,
		conf.AppName,
		conf.ProjectName,
		conf.ProjectID,
	)

	// 初始化sentry
	if !conf.DisableSentry && conf.Sentry.DSN != "" {
		sentry.Init(conf.Sentry)
	}

	// 初始化链路跟踪
	if !conf.DisableTracing {
		tracing.New(conf.Trace)
		defer tracing.Close()
	}

	// 初始化数据统计
	if !conf.DisableMetrics {
		metric.Init()
		svrSlice = append(svrSlice, new(MetricsServer))
	}

	// 初始化goroutine
	goroutine.Init(conf.Goroutine)

	// 开启pprof
	if !conf.DisablePProf {
		svrSlice = append(svrSlice, new(PProfServer))
	}
	// 启动服务器
	for _, svr := range svrSlice {
		svr.Start(conf, svc)
	}

	// 获取服务运行时错误管道
	errChan := svc.Error()

	// 监听系统信号量
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		// 系统退出
		case s := <-osSignal:
			log.Infoc(context.Background(), "Env:%s AppName:%s ProjectName:%s ProjectID:%s  get a signal %s",
				conf.Env,
				conf.AppName,
				conf.ProjectName,
				conf.ProjectID,
				s.String(),
			)

			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				generateShutdown(conf, svc, svrSlice...)
				return
			case syscall.SIGHUP:
			default:
				return
			}
		// 运行异常/正常结束
		case e := <-errChan:
			if e != nil {
				log.Infoc(context.Background(), "Env:%s AppName:%s ProjectName:%s ProjectID:%s run server error %s",
					conf.Env,
					conf.AppName,
					conf.ProjectName,
					conf.ProjectID,
					e,
				)
				sentry.CaptureWithBreadAndTags(context.Background(), e, &sentry.Breadcrumb{
					Category: "runServer",
					Data: map[string]interface{}{
						"env":   conf.AppConfig.Env,
						"appID": conf.AppConfig.AppID,
					},
				})
			} else {
				log.Infoc(context.Background(), "Env:%s AppName:%s ProjectName: %s ProjectID: %s server finish",
					conf.Env,
					conf.AppName,
					conf.ProjectName,
					conf.ProjectID,
				)
			}
			generateShutdown(conf, svc, svrSlice...)
			return
		}
	}
}

// 优雅关闭
func generateShutdown(conf *Config, svc ServiceInterface, svrSlice ...ServerInterface) {
	if conf.ShunDownTimeout == 0 {
		conf.ShunDownTimeout = ctime.Duration(time.Second * 30)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conf.ShunDownTimeout))

	// 关闭服务器
	for _, svr := range svrSlice {
		if err := svr.ShutDown(ctx); err != nil {
			log.Errorc(context.Background(), "Env:%s AppName:%s ProjectName: %s ProjectID: %s Server:%s  Shutdown error(%v)",
				conf.Env,
				conf.AppName,
				conf.ProjectName,
				conf.ProjectID,
				svr.Name(),
				err,
			)
			sentry.CaptureWithBreadAndTags(context.Background(), err, &sentry.Breadcrumb{
				Category: "runServer",
				Data: map[string]interface{}{
					"srvName": svr.Name(),
					"env":     conf.AppConfig.Env,
					"appID":   conf.AppConfig.AppID,
				},
			})
		}
	}
	// 关闭服务
	svc.Close(ctx)

	log.Infoc(context.Background(), "Env:%s AppName:%s ProjectName: %s ProjectID: %s  exit",
		conf.Env,
		conf.AppName,
		conf.ProjectName,
		conf.ProjectID,
	)
	cancel()
	time.Sleep(time.Second)
}
