package main

import (
	"rulai/config"
	"rulai/models"
	"rulai/models/req"
	"rulai/service"

	"context"

	framework "gitlab.shanhai.int/sre/app-framework"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/log"
)

// GetAppendLogStoreServer 获取追加logstore服务器
func GetAppendLogStoreServer() framework.ServerInterface {
	svr := new(framework.JobServer)
	svr.SetJob("append logStore server", func(ctx context.Context) error {
		page := 1
		limit := 100
		projects, err := service.SVC.GetProjects(ctx, &req.GetProjectsReq{
			BaseListRequest: models.BaseListRequest{
				Page:  page,
				Limit: limit,
			},
		})
		if err != nil {
			log.Error("get projects failed! project page:%v limit:%v\n err:%v", page, limit, err)
			return err
		}
		for len(projects) > 0 {
			// 更新project
			for _, project := range projects {
				pdetail, e := service.SVC.GetProjectDetail(ctx, project.ID)
				if e != nil {
					log.Error("get project detail failed! project:%v err:%v\n", project.ID, e)
					continue
				}
				if pdetail.LogStoreName != "" {
					continue
				}
				e = service.SVC.UpdateProject(ctx, project.ID, &req.UpdateProjectReq{
					LogStoreName: project.Name,
				})
				if e != nil {
					log.Error("update project detail failed! project:%v err:%v\n", project.ID, e)
					continue
				}
			}

			// 循环获取project
			page++
			projects, err = service.SVC.GetProjects(ctx, &req.GetProjectsReq{
				BaseListRequest: models.BaseListRequest{
					Page:  page,
					Limit: limit,
				},
			})
			if err != nil {
				log.Error("get projects failed! project page:%v limit:%v err:%v\n", page, limit, err)
				break
			}
		}
		return nil
	})
	return svr
}

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
		GetAppendLogStoreServer(),
	)
}
