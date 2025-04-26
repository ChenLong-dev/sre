package main

import (
	"rulai/config"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/service"

	"context"

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

	framework.Run(
		config.Conf.Config,
		service.New(),
		GetAppendCronJobAppEnvServer(),
	)
}

// GetAppendCronJobAppEnvServer 补全cronjob类型应用的环境参数
func GetAppendCronJobAppEnvServer() framework.ServerInterface {
	svr := new(framework.JobServer)

	svr.SetJob("AppendCronJobAppEnv", func(ctx context.Context) error {
		projects, err := service.SVC.GetProjects(ctx, &req.GetProjectsReq{})
		if err != nil {
			return err
		}

		for _, projectList := range projects {
			apps, err := service.SVC.GetApps(ctx, &req.GetAppsReq{
				Type:      entity.AppTypeCronJob,
				ProjectID: projectList.ID,
			})
			if err != nil {
				return err
			}
			if len(apps) == 0 {
				continue
			}

			project, err := service.SVC.GetProjectDetail(ctx, projectList.ID)
			if err != nil {
				log.Errorc(ctx, "get project detail error:%s", err)
				continue
			}

			for _, appList := range apps {
				app, err := service.SVC.GetAppDetail(ctx, appList.ID)
				if err != nil {
					log.Errorc(ctx, "get app detail error:%s", err)
					continue
				}

				createReq := &req.CreateAppReq{
					Name:              app.Name,
					Type:              app.Type,
					ServiceType:       app.ServiceType,
					ProjectID:         app.ProjectID,
					SentryProjectSlug: app.SentryProjectSlug,
					Description:       app.Description,
				}

				updateReq := &req.UpdateAppReq{
					Env: make(map[entity.AppEnvName]req.UpdateAppEnvReq),
				}
				cronjobEnvs := []entity.AppEnvName{entity.AppEnvFat, entity.AppEnvPre}

				for _, env := range cronjobEnvs {
					if _, ok := app.Env[env]; !ok {
						updateReq.Env[env] = req.UpdateAppEnvReq{
							AliAlarmName: createReq.GetAliAlarmName(project.Name, string(env)),
							LogStoreName: createReq.GetLogStoreName(project.Name, string(env)),
							// 默认七层负载均衡
							ServiceProtocol: entity.LoadBalancerProtocolHTTP,
						}
					}
				}

				if len(updateReq.Env) > 0 {
					err = service.SVC.UpdateApp(ctx, app.ID, updateReq)
					if err != nil {
						log.Errorc(ctx, "update app detail error:%s", err)
						continue
					}
					log.Infoc(ctx, "success update app detail, project:%s - %s app:%s",
						project.ID, project.Name, app.Name)
				}
			}
		}

		return nil
	})

	return svr
}
