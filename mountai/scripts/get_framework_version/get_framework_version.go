package main

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"

	"context"
	"strconv"
	"strings"

	framework "gitlab.shanhai.int/sre/app-framework"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/log"
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
		GetFrameworkVersionServer(),
	)
}

func GetFrameworkVersionServer() framework.ServerInterface {
	svr := new(framework.JobServer)

	svr.SetJob("get framework version", func(ctx context.Context) error {
		projects, err := service.SVC.GetProjects(ctx, &req.GetProjectsReq{
			Language: entity.LanguageGo,
		})
		if err != nil {
			panic(err)
		}

		for _, project := range projects {
			version, err := service.SVC.GetQTFrameworkVersion(ctx, project.ID, "master")
			if err != nil {
				log.Error("%s", err)
			}
			if version.FrameworkVersion != "" {
				lv := strings.Split(version.LibraryVersion, ".")
				c1, err := strconv.Atoi(lv[1])
				if err != nil {
					log.Error("%s", err)
				}
				c2, err := strconv.Atoi(lv[2])
				if err != nil {
					log.Error("%s", err)
				}
				if (c1 == 6 && c2 >= 6) || c1 == 7 {
					log.Info("F:%7s L:%7s P:%s %s %7s ", version.FrameworkVersion, version.LibraryVersion,
						project.ID, project.Name, version.LibraryVersion)
				}
			}
		}

		return nil
	})

	return svr
}
