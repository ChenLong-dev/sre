package main

import (
	framework "gitlab.shanhai.int/sre/app-framework"

	"rulai/config"
	"rulai/server/worker"
	"rulai/service"
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
		worker.GetPodWatcher(),
	)
}
