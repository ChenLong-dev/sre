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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetCreateLogStoreServer 获取创建logstore服务器
func GetCreateLogStoreServer() framework.ServerInterface {
	svr := new(framework.JobServer)
	svr.SetJob("create logStore server", func(ctx context.Context) error {
		page := 1
		limit := 100
		projects, err := service.SVC.GetProjects(ctx, &req.GetProjectsReq{
			BaseListRequest: models.BaseListRequest{
				Page:  page,
				Limit: limit,
			},
		})
		if err != nil {
			log.Error("get projects failed: project page:%v limit:%v err:%v\n", page, limit, err)
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
					log.Error("get apps failed: project:%v err:%v", project.ID, e)
					continue
				}

				// 应用app阿里云配置
				for _, app := range apps {
					errLog := applyLogStoreAndLogTail(ctx, project, app)
					if errLog != nil {
						log.Error("apply logstore and logtail failed: project:%v app:%v err:%v", project.ID, app.ID, errLog)
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
				log.Error("get projects failed project page:%v limit:%v\n", page, limit)
				break
			}
		}
		return nil
	})
	return svr
}

// 2. 为应用生成logstore和logtail
func applyLogStoreAndLogTail(ctx context.Context, project *resp.ProjectListResp, app *resp.AppListResp) error {
	projectResp, err := service.SVC.GetProjectDetail(ctx, project.ID)
	if err != nil {
		return err
	}

	appResp, err := service.SVC.GetAppDetail(ctx, app.ID)
	if err != nil {
		return err
	}

	clusters, err := service.SVC.GetClusters(ctx, new(req.GetClustersReq))
	if err != nil {
		return err
	}

	team := projectResp.Team

	envs := entity.AppNormalEnvNames
	for _, env := range envs {
		for _, cluster := range clusters {
			aliconfig, err := service.SVC.GetAliLogConfigDetail(ctx, cluster.Name,
				&req.GetAliLogConfigDetailReq{
					Namespace: string(env),
					Name:      appResp.AliLogConfigName,
				})
			if err != nil {
				log.Error("get aliconfig failed env:%v, projetc:%v, app:%v\n, err:%v", env, projectResp.ID, appResp.ID, err)
				continue
			}

			logStore, ok, err := unstructured.NestedString(aliconfig.UnstructuredContent(), "spec", "logstore")

			if err != nil {
				log.Error("get aliconfig logstore failed env:%v, projetc:%v, app:%v\n, err:%v", env, projectResp.ID, appResp.ID, err)
				continue
			}
			if ok && logStore == projectResp.LogStoreName {
				continue
			}

			taskResp := &resp.TaskDetailResp{EnvName: env}

			// 渲染模版
			data, err := service.SVC.RenderAliLogConfigTemplate(ctx, projectResp, appResp, taskResp, team)
			if err != nil {
				log.Error("render alilogconfig failed env:%v, projetc:%v, app:%v\n, err:%v", env, projectResp.ID, appResp.ID, err)
				continue
			}

			// 生成AliLogConfig
			_, err = service.SVC.ApplyAliLogConfig(ctx, cluster.Name, string(env), data)
			if err != nil {
				log.Error("apply alilogconfig failed env:%v, projetc:%v, app:%v\n, err:%v", env, projectResp.ID, appResp.ID, err)
				continue
			}

			// create index for log store
			err = service.SVC.CreateAliLogIndexFromLogConfig(ctx, cluster.Name,
				&req.AliCreateStoreLogIndexReq{
					Namespace:        string(env),
					AliLogConfigName: appResp.AliLogConfigName,
					EnvName:          string(env),
				})
			if err != nil {
				log.Error("create index failed! %v", projectResp.LogStoreName)
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
		GetCreateLogStoreServer(),
	)
}
