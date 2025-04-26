package rpc

import (
	"gitlab.shanhai.int/sre/app-framework"
	commonv1 "gitlab.shanhai.int/sre/app-framework/internal/example/api/common/v1"
	"gitlab.shanhai.int/sre/app-framework/internal/example/server/rpc/handler"
	"google.golang.org/grpc"
)

// ====================
// >>>请勿删除<<<
//
// 获取rpc服务器
// ====================
func GetServer() framework.ServerInterface {
	svr := new(framework.GRPCServer)

	svr.Register = func(s *grpc.Server) {
		// 注册入口处理器
		commonv1.RegisterEntryServer(s, new(handler.EntryHandler))
	}

	return svr
}
