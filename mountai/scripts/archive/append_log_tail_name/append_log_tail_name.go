package main

import (
	"rulai/config"
	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"

	"context"
	"fmt"

	framework "gitlab.shanhai.int/sre/app-framework"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/log"
)

// GetAppendLogTailServer 获取追加logtail服务器
func GetAppendLogTailServer() framework.ServerInterface {
	svr := new(framework.JobServer)
	svr.SetJob("append logTail server", func(ctx context.Context) error {
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
				apps, e := service.SVC.GetApps(ctx, &req.GetAppsReq{
					ProjectID: project.ID,
				})
				if e != nil {
					log.Error("get apps failed! projectID:%v err:%v\n", project.ID, e)
					continue
				}
				for _, app := range apps {
					appDetail, errApp := service.SVC.GetAppDetail(ctx, app.ID)
					if errApp != nil {
						log.Error("get app detail failed app:%v, err:%v", appDetail.ID, errApp)
						continue
					}

					updateReq := make(map[entity.AppEnvName]req.UpdateAppEnvReq)
					for _, env := range entity.AppNormalEnvNames {
						if appDetail.Env[env].LogTailName == "" {
							updateReq[env] = req.UpdateAppEnvReq{LogTailName: fmt.Sprintf("%v-%v-%v", project.Name, appDetail.Name, env)}
						}
					}
					if len(updateReq) == 0 {
						continue
					}
					errApp = service.SVC.UpdateApp(ctx, appDetail.ID, &req.UpdateAppReq{
						Env: updateReq,
					})
					if errApp != nil {
						log.Error("update logtail failed! app: %v\n", appDetail.ID)
					}
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
		GetAppendLogTailServer(),
	)
}
