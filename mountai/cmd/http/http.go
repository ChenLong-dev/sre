package main

import (
	"fmt"
	"os"
	"rulai/config"
	"rulai/server/http"
	"rulai/service"

	framework "gitlab.shanhai.int/sre/app-framework"
)

func main() {
	// 启动服务
	framework.Run(
		// ========================
		// 读取并新建配置文件
		// ========================
		config.Read(fmt.Sprintf("./cm/config.%s.yaml", os.Getenv("env"))).Config,

		// ====================
		// >>>请勿删除<<<
		//
		// 新建服务
		// ====================
		service.New(),

		// 启动http服务器
		http.GetServer(),
	)
}
