package main

import (
	"rulai/config"
	"rulai/server/worker"
	"rulai/service"

	framework "gitlab.shanhai.int/sre/app-framework"
)

func main() {
	framework.Run(
		config.Read("./config/config.yaml").Config,
		service.New(),
		worker.DingTalkCallbackConsumer(),
	)
}
