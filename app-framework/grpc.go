package framework

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"gitlab.shanhai.int/sre/library/goroutine"
)

// GRPC服务器
type GRPCServer struct {
	// GRPC服务器
	server *grpc.Server
	// 服务名
	name string
	// 配置文件
	config *Config
	// 注册函数
	Register func(svr *grpc.Server)
}

// 实现ServerInterface
func (svr *GRPCServer) ShutDown(ctx context.Context) (err error) {
	// 尝试优雅关闭
	ch := make(chan struct{})
	eg := goroutine.New("RPC")
	eg.Go(ctx, "GracefulStop", func(ctx context.Context) error {
		svr.server.GracefulStop()
		close(ch)
		return nil
	})

	select {
	// 超时，强制关闭
	case <-ctx.Done():
		svr.server.Stop()
		err = ctx.Err()
	case <-ch:
	}

	return
}

// 实现ServerInterface
func (svr *GRPCServer) Start(c *Config, svc ServiceInterface) {
	svr.name = "GRPC"
	svr.config = c
	svr.server = grpc.NewServer()

	// 注册官方健康检查
	// 健康检查状态先置为未就绪状态，防止健康检查先于其他服务生效
	hsrv := health.NewServer()
	hsrv.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	healthpb.RegisterHealthServer(svr.server, hsrv)

	// 注册反射服务
	reflection.Register(svr.server)

	// 注册自定义服务
	svr.Register(svr.server)

	listener, err := net.Listen("tcp",
		fmt.Sprintf("%s:%d", c.RPC.Endpoint.Address, c.RPC.Endpoint.Port))
	if err != nil {
		panic(err)
	}
	svc.StartServer(svr.name, func(ctx context.Context) error {
		return svr.server.Serve(listener)
	})

	// gRPC 服务启动后更新健康检查状态
	hsrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
}

// 实现ServerInterface
func (svr *GRPCServer) Name() string {
	return svr.name
}
