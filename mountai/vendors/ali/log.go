// TODO: 迁移阿里云日志相关逻辑至此
package ali

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/net/errcode"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"
)

// ApplyLogConfig 声明式创建/更新日志采集器配置
func (c *Controller) ApplyLogConfig(ctx context.Context,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, task *resp.TaskDetailResp, team *resp.TeamDetailResp) error {
	if c.DisableLogConfig {
		return _errcode.LogConfigDisabled
	}

	return errors.Wrap(errcode.InternalError, "not supported yet")
}

// LogConfigExistanceCheck 检验日志采集器配置是否存在(华为创建配置成功后立即生效)
func (c *Controller) LogConfigExistanceCheck(_ context.Context, _ entity.ClusterName, _ entity.AppEnvName, _ *resp.AppDetailResp) error {
	return errors.Wrap(errcode.InternalError, "not supported yet")
}

// DeleteLogConfig 删除日志采集器配置, 同时删除数据库中的资源记录
func (c *Controller) DeleteLogConfig(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error {
	return errors.Wrap(errcode.InternalError, "not supported yet")
}

// EnsureLogIndex 创建日志索引(华为云日志结构化目前不适用, 跳过该阶段)
func (c *Controller) EnsureLogIndex(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, app *resp.AppDetailResp) error {
	return errors.Wrap(errcode.InternalError, "not supported yet")
}

// ApplyLogDump 声明式创建/更新日志转储
func (c *Controller) ApplyLogDump(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error {
	return errors.Wrap(errcode.InternalError, "not supported yet")
}

// DeleteLogDump 删除日志转储
func (c *Controller) DeleteLogDump(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error {
	return errors.Wrap(errcode.InternalError, "not supported yet")
}

func (c *Controller) GetLogStoreURL(ctx context.Context, clusterName entity.ClusterName, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, envName entity.AppEnvName) (logStoreURL, logStoreURLBasedProject string, err error) {
	logStoreURL = c.getAppLogStoreURLWithFilter(envName, app.Env[envName].LogStoreName, nil)
	// TODO: 创建日志索引
	logStoreURLBasedProject = c.getAppLogStoreURLWithFilter(envName, project.LogStoreName, map[string]string{
		"_container_name_": utils.GetPodContainerName(project.Name, app.Name),
	})
	return logStoreURL, logStoreURLBasedProject, nil
}

func (c *Controller) getAppLogStoreURLWithFilter(envName entity.AppEnvName, logStoreName string, filter map[string]string) string {
	projectName := config.Conf.Other.AliLogProjectStgName
	if envName == entity.AppEnvPrd || envName == entity.AppEnvPre {
		projectName = config.Conf.Other.AliLogProjectPrdName
	}

	queryString := ""
	for k, v := range filter {
		if queryString == "" {
			queryString += k + ":" + v
			continue
		}
		queryString += " and " + k + ":" + v
	}

	if queryString == "" {
		return fmt.Sprintf(
			"%s/lognext/project/%s/logsearch/%s",
			config.Conf.Other.AliSLSConsoleURL, projectName, logStoreName,
		)
	}

	return fmt.Sprintf(
		"%s/lognext/project/%s/logsearch/%s?queryString=%s",
		config.Conf.Other.AliSLSConsoleURL, projectName, logStoreName, queryString,
	)
}
