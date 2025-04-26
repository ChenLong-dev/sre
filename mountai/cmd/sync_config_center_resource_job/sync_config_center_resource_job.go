package main

import (
	"rulai/config"
	"rulai/server/job"
	"rulai/service"

	framework "gitlab.shanhai.int/sre/app-framework"
)

func main() {
	// 启动服务
	framework.Run(
		config.Read("./config/config.yaml").Config,

		// ====================
		// >>>请勿删除<<<
		//
		// 新建服务
		// ====================
		service.New(),

		// 启动定时任务
		job.GetSyncConfigCenterResourceServer(),
	)
}
