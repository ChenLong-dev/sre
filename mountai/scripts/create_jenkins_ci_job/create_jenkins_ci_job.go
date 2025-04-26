package main

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"
	_errcode "rulai/utils/errcode"

	"context"

	framework "gitlab.shanhai.int/sre/app-framework"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
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
		CreateJenkinsCIJob(),
	)
}

func CreateJenkinsCIJob() framework.ServerInterface {
	svr := new(framework.JobServer)

	svr.SetJob("create jenkins ci job", func(ctx context.Context) error {
		projects, err := service.SVC.GetProjects(ctx, &req.GetProjectsReq{})
		if err != nil {
			log.Error("get projects err:%v", err)
			return err
		}

		for _, project := range projects {
			err = service.SVC.CreateProjectCIJob(ctx, project.ID, project.Name, &req.CreateProjectCIJobReq{
				MessageNotification: []entity.NotificationType{entity.NotificationTypeDingDing},
				PipelineStages:      entity.DefaultPipelineStages,
			})
			if err != nil && !errcode.EqualError(_errcode.ProjectCIJobExistsError, err) {
				log.Error("create jenkins job failed: project: %s, err: %v", project.Name, err)
				continue
			}
		}

		return nil
	})

	return svr
}
