package huawei

import (
	"context"
	"fmt"

	"gitlab.shanhai.int/sre/library/base/null"

	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/sentry"
	"go.mongodb.org/mongo-driver/bson/primitive"

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

	stream, err := c.DAO.FindSingleAppLTSStream(ctx, task.ClusterName, task.EnvName, app.ID)
	if err != nil {
		if !errcode.EqualError(errcode.NoRowsFoundError, err) {
			return err
		}

		log.Infoc(ctx, "stream of app(%s) not exists", app.ID)
	}

	// 已存在日志流记录时需要获取日志流实际的名字(有部分老日志流是多集群之前产生的, 命名规则不一样)
	// TODO: 老命名规则的日志流与业务协商迁移
	if stream == nil {
		// LTS 日志流记录不存在时时需要尝试创建日志流
		// 华为云日志采集器 AOM 不需要接入(配置在工作负载中通过 annotations 已经定义收集规则)
		// 但 AOM 日志需要与 LTS 日志服务进行关联(创建 LTS 日志流 -> 创建 AOM-LTS 接入规则)
		// 创建完日志流时必须先记录 logstream_id, 因为如果之后创建接入规则失败, 日志流已经存在会无法创建新日志流, 并且没有合适的 API 查询已创建日志流的 ID
		stream = &entity.AppLTSStream{
			ID:          primitive.NewObjectID(),
			AppID:       app.ID,
			ClusterName: task.ClusterName,
			EnvName:     task.EnvName,
			StreamName:  getLogStreamName(project.Name, app.Name, task.ClusterName, task.EnvName),
		}

		valid := true
		if app.LogTTLInDays == entity.LogTTLInDaysEmpty {
			valid = false
		}

		stream.StreamID, err = c.createHuaweiLTSStream(ctx, stream, null.NewInt(app.LogTTLInDays, valid))
		if err != nil {
			// 当前没有合适的查询已存在日志流的 API, 故所有错误均需要人工介入(API响应中有id人工添加记录)
			// TODO: 华为云支持通过名称查询日志流后忽略日志流已存在的错误, 调用 API 获取已存在的日志流 ID
			return err
		}

		err = c.DAO.CreateAppLTSStream(ctx, stream)
		if err != nil {
			return err
		}
	}

	// LTS 接入规则记录不存在时尝试创建接入规则
	ruleName := getRuleName(project.Name, app.Name, task.ClusterName, task.Namespace)
	containerName := utils.GetPodContainerName(project.Name, app.Name)
	// FIXME: 华为云 API 目前有 bug, 填写日志流名称时可能会报重名错误, 当前只能先填写 target_log_stream_id
	// ({"code":"LTS.0746","details":"AOM mapping rule log stream name already exsit in another log group"})
	stream.RuleID, err = c.createHuaweiAOMToLTSRule(ctx, task.EnvName,
		task.ClusterName, task.GetNamespace(app.EnableIstio), ruleName, containerName,
		stream.StreamID, stream.StreamName)
	if err != nil {
		return err
	}

	err = c.DAO.UpdateAppAOMToLTSRule(ctx, task.ClusterName, task.EnvName, app.ID, stream.RuleID)
	if err != nil {
		return err
	}

	return nil
}

// LogConfigExistanceCheck 检验日志采集器配置是否存在(华为创建配置成功后立即生效)
func (c *Controller) LogConfigExistanceCheck(_ context.Context, _ entity.ClusterName, _ entity.AppEnvName, _ *resp.AppDetailResp) error {
	return nil
}

// DeleteLogConfig 删除日志采集器配置, 同时删除数据库中的资源记录
func (c *Controller) DeleteLogConfig(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error {
	stream, err := c.DAO.FindSingleAppLTSStream(ctx, clusterName, envName, appID)
	if err != nil {
		if errcode.EqualError(errcode.NoRowsFoundError, err) {
			return nil
		}
		return err
	}

	err = c.deleteHuaweiAOMToLTSRule(ctx, stream.RuleID)
	if err != nil && errcode.EqualError(_errcode.LTSResourceNotFound, err) {
		// 忽略AOM-日志流介入规则不存在的错误
		return err
	}

	err = c.deleteHuaweiLTSStream(ctx, envName, clusterName, stream.StreamID)
	if err != nil && !errcode.EqualError(_errcode.LTSResourceNotFound, err) {
		// 忽略日志流不存在的错误
		// FIXME: 临时忽略日志流存在关联的错误(避免阻塞status_worker的流程), 后续需要接入华为云日志转储的API
		if errcode.EqualError(_errcode.LTSResourceAssociated, err) {
			log.Errorc(ctx, "deleteHuaweiLTSStream failed with known error: %s", err)
			sentry.CaptureWithBreadAndTags(ctx, err, &sentry.Breadcrumb{
				Category: "deleteHuaweiLTSStream",
				Data:     map[string]interface{}{"LTSStreamID": stream.StreamID},
			})
		} else {
			return err
		}
	}

	return c.DAO.DeleteAppLTSStream(ctx, clusterName, envName, appID)
}

// EnsureLogIndex 创建日志索引(华为云日志结构化目前不适用, 跳过该阶段)
func (c *Controller) EnsureLogIndex(ctx context.Context,
	clusterName entity.ClusterName, envName entity.AppEnvName, app *resp.AppDetailResp) error {
	log.Infoc(ctx, "log index create phase skipped for app(%s) in cluster(%s) and env(%s)", app.ID, clusterName, envName)
	return nil
}

// ApplyLogDump 声明式创建/更新日志转储
// TODO: 华为云提供删除日志转储 API 后进行对接, 当前对接会使得删除流程有问题
func (c *Controller) ApplyLogDump(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error {
	log.Infoc(ctx, "log dump create phase skipped for app(%s) in cluster(%s) and env(%s)", appID, clusterName, envName)
	return nil
}

// DeleteLogDump 删除日志转储
// TODO: 华为云提供删除日志转储 API 后进行对接, 当前对接会使得删除流程有问题
func (c *Controller) DeleteLogDump(ctx context.Context, clusterName entity.ClusterName, envName entity.AppEnvName, appID string) error {
	log.Infoc(ctx, "log dump delete phase skipped for app(%s) in cluster(%s) and env(%s)", appID, clusterName, envName)
	return nil
}

func (c *Controller) GetLogStoreURL(ctx context.Context, clusterName entity.ClusterName, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, envName entity.AppEnvName) (logStoreURL, logStoreURLBasedProject string, err error) {
	clusterCfg, err := c.getClusterConfig(envName, clusterName)
	if err != nil {
		return "", "", err
	}

	appLST, err := c.DAO.FindSingleAppLTSStream(ctx, clusterName, envName, app.ID)
	if err != nil {
		if errcode.EqualError(errcode.NoRowsFoundError, err) {
			return "", "", nil
		}
		return "", "", err
	}

	// 实际查询使用的是 topicId(日志流ID), 界面显示使用的是 topicName(并不会从日志流获取实际名称)
	logStoreURL = fmt.Sprintf(
		"%s/lts/?region=%s#/cts/logEventsLeftMenu/events?groupId=%s&groupName=%s&topicId=%s&topicName=%s&epsId=0",
		config.Conf.Other.HWConsoleURL, clusterCfg.Region, clusterCfg.LogGroupID, clusterCfg.LogGroupName,
		appLST.StreamID, getLogStreamName(project.Name, app.Name, clusterName, envName))
	logStoreURLBasedProject = logStoreURL

	return logStoreURL, logStoreURLBasedProject, nil
}

func getRuleName(projectName, appName string, clusterName entity.ClusterName, envName string) string {
	// 华为云日志接入规则名称当前是全局唯一, 必须严格区分集群和环境防止重名
	return fmt.Sprintf("AMS-log-%s-%s-%s-%s", projectName, appName, envName, clusterName)
}

func getLogStreamName(projectName, appName string, clusterName entity.ClusterName, envName entity.AppEnvName) string {
	// 华为云日志流名称当前是全局唯一, 必须严格区分集群和环境防止重名
	return fmt.Sprintf("%s%s-%s-%s-%s", config.Conf.Ali.PrivateZonePrefix, projectName, appName, envName, clusterName)
}
