package main

import (
	"rulai/config"
	"rulai/models/req"
	"rulai/service"

	"context"
	"fmt"
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
		CleanJenkinsCIJob(),
	)
}

func CleanJenkinsCIJob() framework.ServerInterface {
	svr := new(framework.JobServer)

	svr.SetJob("clean jenkins ci job", func(ctx context.Context) error {
		pageSize := 20
		page := 1
		for {
			ciJobs, err := service.SVC.GetProjectCIJobs(ctx, &req.GetProjectCIJobs{
				Limit: pageSize,
				Page:  page,
			})
			if err != nil {
				return err
			}

			for _, ciJob := range ciJobs {
				project, err := service.SVC.GetProjectDetail(ctx, ciJob.ProjectID)
				if err != nil {
					log.Errorc(ctx, "get project err: %v", err)
					continue
				}

				newHookURL := strings.Replace(ciJob.HookURL, "https://jenkins.qingtingfm.com",
					config.Conf.JenkinsCI.GoJenkins.BaseURL, 1)

				err = service.SVC.EditGitlabProjectHookByURL(ctx, project.ID,
					ciJob.HookURL, &req.GitlabProjectHookDetailReq{
						URL:   newHookURL,
						Token: config.Conf.JenkinsCI.GitlabSecretToken,
					})
				if err != nil {
					log.Errorc(ctx, "edit hook url err: %v", err)
					continue
				}

				err = service.SVC.UpdateProjectCIJob(ctx, ciJob.ProjectID, project.Name,
					ciJob.ID, &req.UpdateProjectCIJobReq{
						MessageNotification: ciJob.MessageNotification,
						PipelineStages:      ciJob.PipelineStages,
						ViewURL: fmt.Sprintf("%s/view/AMS-CI/job/ci-%s",
							config.Conf.JenkinsCI.GoJenkins.BaseURL, project.Name),
						HookURL: newHookURL,
					})
				if err != nil {
					log.Errorc(ctx, "update ci job err: %v", err)
				}
			}
			if len(ciJobs) < pageSize {
				break
			}
			page++
		}
		return nil
	})

	return svr
}
