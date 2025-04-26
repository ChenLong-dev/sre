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
		config.Read(fmt.Sprintf("./config/config.%s.yaml", os.Getenv("env"))).Config,
		service.New(),
		worker.AppOpMsgConsumer(),
	)
}
