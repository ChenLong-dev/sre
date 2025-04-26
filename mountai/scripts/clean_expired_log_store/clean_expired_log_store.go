package main

import (
	"rulai/config"
	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/service"

	"context"

	framework "gitlab.shanhai.int/sre/app-framework"
	render "gitlab.shanhai.int/sre/library/base/logrender"
	"gitlab.shanhai.int/sre/library/log"
)

// 分两步清除：12.15删除近期日志为空logstore，删除由env环境自动生成的logstore；
// GetCleanExpiredLogStoreServer 获取清除过期logstore服务器
func GetCleanExpiredLogStoreServer() framework.ServerInterface {
	svr := new(framework.JobServer)
	svr.SetJob("clean expired logStore server", func(ctx context.Context) error {
		page := 1
		limit := 100
		projects, err := service.SVC.GetProjects(ctx, &req.GetProjectsReq{
			BaseListRequest: models.BaseListRequest{
				Page:  page,
				Limit: limit,
			},
		})
		if err != nil {
			log.Error("get projects failed! project page:%v limit:%v err:%v\n", page, limit, err)
			return err
		}
		for len(projects) > 0 {
			// 更新project
			for _, project := range projects {
				// 获取app
				apps, e := service.SVC.GetApps(ctx, &req.GetAppsReq{
					Page:      1,
					Limit:     100,
					ProjectID: project.ID,
				})
				if e != nil {
					log.Error("get apps failed! project:%v, err:%v\n", project.ID, e)
					continue
				}

				// 删除log sotre
				for _, app := range apps {
					errDel := deleteLogStore(ctx, app)
					if errDel != nil {
						log.Error("delete logstore failed! app:%v, err:%v\n", app.ID, errDel)
						continue
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
				continue
			}
		}
		return nil
	})
	return svr
}

// 3. 删除过期logstore
func deleteLogStore(ctx context.Context, app *resp.AppListResp) error {
	envs := entity.AppNormalEnvNames
	appResp, err := service.SVC.GetAppDetail(ctx, app.ID)
	if err != nil {
		log.Error("get app detail failed! app:%v err:%v\n", appResp.ID, err)
		return err
	}

	for _, env := range envs {
		// 判断集群名
		clusterName := config.Conf.Other.AliLogProjectStgName
		if env == entity.AppEnvPre || env == entity.AppEnvPrd {
			clusterName = config.Conf.Other.AliLogProjectPrdName
		}

		logStore, err := service.SVC.GetLogStoreDetail(ctx, &req.AliGetLogStoreDetailReq{
			ProjectName: clusterName,
			StoreName:   appResp.Env[env].LogStoreName,
		})
		if err == nil {
			// 永久有效日志仓库
			if logStore.TTL == resp.TTLPERMANNENT {
				log.Error("permanent:%v\n", logStore.LogStoreName)
				continue
			}
			// 保存时间超过15天
			if logStore.TTL > 15 {
				log.Error("greater than 15 days:%v\n", logStore.LogStoreName)
				continue
			}

			e := service.SVC.DeleteAliLogStore(ctx, &req.AliDeleteLogStoreReq{
				ProjectName: clusterName,
				StoreName:   appResp.Env[env].LogStoreName,
			})
			if e != nil {
				log.Error("delete log store failed! clusterName:%v, storeName:%v", clusterName, appResp.Env[env].LogStoreName)
				continue
			}
		}
	}
	return nil
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
		GetCleanExpiredLogStoreServer(),
	)
}
