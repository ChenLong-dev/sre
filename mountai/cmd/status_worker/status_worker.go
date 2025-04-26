package main

import (
	"fmt"
	"os"
	"rulai/config"
	"rulai/server/worker"
	"rulai/service"

	framework "gitlab.shanhai.int/sre/app-framework"
)

func main() {
	// 启动服务
	framework.Run(
		config.Read(fmt.Sprintf("./cm/config.%s.yaml", os.Getenv("env"))).Config,

		// ====================
		// >>>请勿删除<<<
		//
		// 新建服务
		// ====================
		service.New(),

		// 启动常驻任务
		worker.GetServer(),
	)
}
