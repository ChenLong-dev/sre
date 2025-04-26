package job

import (
	"rulai/service"

	framework "gitlab.shanhai.int/sre/app-framework"
)

func GetSyncApolloResourceServer() framework.ServerInterface {
	svr := new(framework.JobServer)

	// ====================
	// >>>请勿删除<<<
	//
	// 根据实际情况修改
	// ====================
	// 设置任务函数
	svr.SetJob("sync_apollo_resource", service.SVC.SyncResourcesFromApollo)

	return svr
}
