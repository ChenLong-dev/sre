package main

import (
	"rulai/config"

	"rulai/service"

	"context"

	framework "gitlab.shanhai.int/sre/app-framework"
	render "gitlab.shanhai.int/sre/library/base/logrender"
)

func main() {
	config.Read("./config/config.yaml")
	config.Conf.Log.Config = &render.Config{
		Stdout: true,
		OutDir: "./log",
	}
	config.Conf.Mongo.Config = &render.Config{
		Stdout: true,
		OutDir: "./log",
	}
	config.Conf.HTTPClient.Config = &render.Config{
		Stdout: true,
		OutDir: "./log",
	}

	framework.Run(
		config.Conf.Config,
		service.New(),
		MigrateJenkinsBuild(),
	)
}

func MigrateJenkinsBuild() framework.ServerInterface {
	svr := new(framework.JobServer)

	svr.SetJob("migrate jenkins build", func(ctx context.Context) error {
		return service.SVC.SyncOldJenkinsBuild(ctx)
	})

	return svr
}
