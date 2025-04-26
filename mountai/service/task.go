package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"gitlab.shanhai.int/sre/library/base/deepcopy.v2"
	qtM "gitlab.shanhai.int/sre/library/database/mongo"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"rulai/dao"
	"rulai/models"
	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
	"rulai/utils"
	_errcode "rulai/utils/errcode"

	v1 "k8s.io/api/core/v1"
)

const (
	// 最新应用配置默认值
	latestConfigCommitID = "@latest"

	// branchChangeDeployMsg 切换分支部署时发送的消息格式
	branchChangeDeployMsg = `
### {{.Title}}
>- [项目名] {{.ProjectName}}
>- [应用名] {{.AppName}}
>- [环境名] {{.Env}}
>- [原分支] {{.OriginBranch}}
>- [当前分支] {{.CurrentBranch}} 
>- [操作人] {{.OperatorName}}
>- [安全关键字] {{.SecurityKeywords}}
#### [进入AMS, 查看详情]({{.DetailURL}})
`

	// 默认配置文件映射路径 k8s configmap
	defaultConfigMountPath = "/root/cm"

	// 默认会话保持过期时间 单位秒
	defaultSessionCookieMaxAge = 1800

	// 默认 prd 环境预执行命令
	defaultPrdPreStopCommand = "sleep 15"
	// 默认非 prd 环境预执行命令
	defaultNonPrdPreStopCommand = "sleep 5"
)

// GenerateTaskVersion : 生成任务版本
func (s *Service) GenerateTaskVersion(projectName, appName string, appType entity.AppType, taskAction entity.TaskAction, createTime time.Time) string {
	if appType == entity.AppTypeOneTimeJob {
		return fmt.Sprintf("%s-%s-%s", projectName, appName, createTime.Format("20060102150405"))
	}

	if taskAction == entity.TaskActionCanaryDeploy {
		return fmt.Sprintf("%s-%s%s", projectName, appName, entity.TaskCanaryVersionSuffix)
	}

	return fmt.Sprintf("%s-%s", projectName, appName)
}

// CreateTask : 创建任务
func (s *Service) CreateTask(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	createReq *req.CreateTaskReq, operatorID string) (*resp.TaskDetailResp, error) {
	// 生成task实体
	task, err := s.generateTaskEntity(ctx, project, app, createReq, operatorID)
	if err != nil {
		return nil, err
	}

	if task.Approval.Type == entity.DefaultTaskApprovalType {
		approvalReq, e := s.generateDingApprovalReq(ctx, task, project, app)
		if e != nil {
			return nil, e
		}

		_, err = s.dao.Mongo.Connection().Transaction(ctx, func(con *qtM.Connection, ctx mongo.SessionContext) (interface{}, error) {
			approval, e := s.CreateDingApprovalInstance(ctx, approvalReq)
			if e != nil {
				return nil, e
			}

			task.Approval.InstanceID = approval.Data.ProcessInstanceID

			_, e = s.dao.CreateSingleTask(ctx, task)
			return nil, e
		})
		if err != nil {
			return nil, errors.Wrapf(_errcode.MongoTransactionInternalError, "transaction err: %v", err)
		}
	} else {
		// 落库
		_, err = s.dao.CreateSingleTask(ctx, task)
		if err != nil {
			return nil, err
		}

		// createReq.Approval.Type != "", it's just to skip "batch create".
		if s.IsPrdP0LevelApp(createReq.EnvName, project) && createReq.Action == entity.TaskActionCanaryDeploy &&
			task.Approval.Type != "" {
			// Send urgent deployment ding message and create record.
			// TODO: 发送飞书通知
			// ret, e := s.SendUrgentDeployDingCropMessage(ctx, project, app, createReq)
			// if e != nil {
			// 	log.Errorc(ctx, "send urgent deployment ding crop message err: %v", e)
			// } else {
			// 	// Creates urgent deployment ding message record.
			// 	appID, e := primitive.ObjectIDFromHex(app.ID)
			// 	if e != nil {
			// 		return nil, errors.Wrapf(_errcode.InvalidHexStringError, "%s", e.Error())
			// 	}

			// 	now := time.Now()
			// 	e = s.CreateDingTalkUrgentDeployMsgRecord(ctx, &entity.UrgentDeploymentDingTalkMsgRecord{
			// 		ID:             primitive.NewObjectID(),
			// 		Env:            createReq.EnvName,
			// 		ProjectID:      project.ID,
			// 		IsP0Level:      s.IsP0LevelProject(project),
			// 		AppID:          appID,
			// 		CropMsgTaskID:  ret.TaskID,
			// 		TaskActionType: entity.GetTaskActionDisplay(createReq.Action),
			// 		Operator:       operatorID,
			// 		CreateTime:     &now,
			// 		UpdateTime:     &now,
			// 	})
			// 	if e != nil {
			// 		log.Errorc(ctx, "create ding talk urgent deployment message err: %v", e)
			// 	}
			// }
		}
	}

	// 创建 task 后, 如果是初始发布操作, 添加部署信息至 redis
	for _, act := range entity.TaskActionInitDeployList {
		if act != task.Action {
			continue
		}

		res := new(resp.TaskDetailResp)
		err = deepcopy.Copy(task).To(res)
		if err != nil {
			return nil, err
		}

		err = s.dao.SetAppRunningTasks(ctx, task.AppID, task.EnvName, []*resp.TaskDetailResp{res}, false)
		if err != nil {
			return nil, err
		}

		break
	}

	// 发送通知
	if err = s.BranchChangeSendNotice(ctx, createReq, project, app); err != nil {
		log.Errorc(ctx, "branch change send msg to ding talk err: %s", err.Error())
	}

	res := new(resp.TaskDetailResp)
	err = deepcopy.Copy(task).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// 生成task实体
func (s *Service) generateTaskEntity(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, createReq *req.CreateTaskReq, operatorID string) (*entity.Task, error) {
	now := time.Now()
	task := &entity.Task{
		ID:         primitive.NewObjectID(),
		OperatorID: operatorID,
		Status:     entity.TaskStatusInit,
		CreateTime: &now,
		UpdateTime: &now,
		Param: &entity.TaskParam{
			Vars:         make(map[string]string),
			ExposedPorts: make(map[string]int),
		},
		Namespace: s.GetNamespaceBase(s.GetApplicationIstioState(ctx, createReq.EnvName, createReq.ClusterName, app),
			createReq.EnvName),
	}
	err := deepcopy.Copy(createReq).
		SetConfig(&deepcopy.Config{
			NotZeroMode: true,
		}).
		To(task)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	if task.Version == "" {
		task.Version = s.GenerateTaskVersion(project.Name, app.Name, app.Type, createReq.Action, now)
	}

	if task.Action == entity.TaskActionFullCanaryDeploy {
		task.Action = entity.TaskActionFullDeploy
		task.Version = strings.TrimSuffix(task.Version, entity.TaskCanaryVersionSuffix)
	}

	// 校验 config 是否有问题
	if createReq.Param.ConfigCommitID != "" {
		getReq := &req.GetConfigManagerFileReq{
			ProjectID:   project.ID,
			ProjectName: project.Name,
			EnvName:     task.EnvName,
			CommitID:    "",
			IsDecrypt:   false,
			FormatType:  req.ConfigManagerFormatTypeJSON,
		}

		if task.Param.ConfigCommitID != latestConfigCommitID {
			getReq.CommitID = task.Param.ConfigCommitID
		}

		if createReq.Param.ConfigRenamePrefix != "" {
			getReq.ConfigRenamePrefix = createReq.Param.ConfigRenamePrefix
			getReq.ConfigRenameMode = createReq.Param.ConfigRenameMode
		}

		configData, e := s.GetAppConfig(ctx, getReq)
		if e != nil {
			return nil, e
		}

		if task.Param.ConfigCommitID == latestConfigCommitID {
			// 填充最新的配置id
			task.Param.ConfigCommitID = configData.CommitID
		}
	}

	// 重启任务时，最小实例数为上次部署实例数
	if task.Action == entity.TaskActionRestart {
		deploy, e := s.GetLatestDeploySuccessTaskFinalVersion(ctx, &req.GetLatestTaskReq{
			AppID:       app.ID,
			EnvName:     task.EnvName,
			ClusterName: task.ClusterName,
			Version:     task.Version,
		})
		if e != nil {
			return nil, e
		}

		task.Param.MinPodCount = deploy.Param.MinPodCount
	}

	// 创建手动启动Job名称
	if createReq.Action == entity.TaskActionManualLaunch {
		task.Param.ManualJobName = s.getManualJobVersion(createReq.Version)
	}

	// 修正默认terminationGracePeriodSeconds值
	if task.Param.TerminationGracePeriodSeconds == 0 {
		task.Param.TerminationGracePeriodSeconds = entity.TerminationGracePeriodSpanTiny
	}

	// 设置service类型应用默认preStop命令
	if app.Type == entity.AppTypeService && task.Param.PreStopCommand == "" {
		task.Param.PreStopCommand = defaultNonPrdPreStopCommand
		if task.EnvName == entity.AppEnvPrd {
			task.Param.PreStopCommand = defaultPrdPreStopCommand
		}
	}

	// 会话保持过期时间
	if createReq.Param.IsSupportStickySession && createReq.Param.SessionCookieMaxAge == 0 {
		task.Param.SessionCookieMaxAge = defaultSessionCookieMaxAge
	}

	// 补全配置文件挂载绝对路径
	if createReq.Param.ConfigMountPath == "" {
		task.Param.ConfigMountPath = defaultConfigMountPath
	} else {
		task.Param.ConfigMountPath = createReq.Param.ConfigMountPath
	}

	// 存活探针初始化延迟时长
	if createReq.Param.LivenessProbeInitialDelaySeconds == 0 {
		task.Param.LivenessProbeInitialDelaySeconds = entity.DefaultProbeDelaySeconds
	} else {
		task.Param.LivenessProbeInitialDelaySeconds = createReq.Param.LivenessProbeInitialDelaySeconds
	}

	// 可读探针初始化延迟时长
	if createReq.Param.ReadinessProbeInitialDelaySeconds == 0 {
		task.Param.ReadinessProbeInitialDelaySeconds = entity.DefaultProbeDelaySeconds
	} else {
		task.Param.ReadinessProbeInitialDelaySeconds = createReq.Param.ReadinessProbeInitialDelaySeconds
	}

	// Change task action to reload_config when it create a restart task and need config.
	// TODO: Remove me when frontend support this judgement
	if task.Param.ConfigCommitID != "" && task.Action == entity.TaskActionRestart {
		task.Action = entity.TaskActionReloadConfig
		log.Infoc(ctx, "transport task action from %s to %s", entity.TaskActionRestart, entity.TaskActionReloadConfig)
	}

	if len(createReq.Param.CronScaleJobGroups) != 0 {
		task.Param.CronScaleJobExcludeDates = createReq.Param.CronScaleJobExcludeDates
		task.Param.CronScaleJobGroups = createReq.Param.CronScaleJobGroups
	}

	if createReq.ScheduleTime != 0 {
		scheduleTime := time.Unix(createReq.ScheduleTime, 0)
		task.ScheduleTime = &scheduleTime
	}

	if task.Approval == nil {
		task.Approval = new(entity.Approval)
		task.Approval.Type = entity.SkipTaskApprovalType
	}
	if task.Approval.Type == entity.SkipTaskApprovalType {
		// Change approval status to approved when task don't need approval process.
		task.Approval.Status = entity.ApprovedTaskApprovalStatus
	}
	if task.Approval.Type == entity.DefaultTaskApprovalType {
		task.Approval.Status = entity.ApprovingTaskApprovalStatus
	}

	// 填充任务细节（务必在所有task.Param赋值后填充）
	taskJSON, err := json.MarshalIndent(task, "", "\t")
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	task.Detail = s.generateTaskDetailOneLine(fmt.Sprintf("create task, task entity:\n%s", taskJSON))

	return task, nil
}

// GetLatestDeploySuccessTaskBranch 获取最新成功部署的任务所属分支
func (s *Service) GetLatestDeploySuccessTaskBranch(ctx context.Context, createReq *req.CreateTaskReq) (string, error) {
	// Step: 获取上次该环境部署的分支
	lastTask, err := s.GetLatestSuccessTask(ctx, &req.GetLatestTaskReq{
		AppID:       createReq.AppID,
		EnvName:     createReq.EnvName,
		ClusterName: createReq.ClusterName,
		ActionList:  entity.TaskActionInitDeployList,
	})
	if errcode.EqualError(_errcode.NoRequiredTaskError, err) {
		return "", nil
	} else if err != nil {
		return "", err
	} else if lastTask.Param == nil {
		return "", _errcode.InvalidDeployTaskParamError
	} else if lastTask.Param.ImageVersion == "" {
		return "", nil
	}

	// Step: 获取分支
	_, _, lastTaskBranch, err := s.ExtraInfoFromImageVersion(lastTask.Param.ImageVersion)
	// 兼容老版本imageVersion无法通过校验，否则无法发布当前分支
	if errcode.EqualError(_errcode.InvalidImageVersionError, err) {
		log.Errorc(ctx, "the image version of latest success task : %s failed to parse err: %s",
			lastTask.Param.ImageVersion, err.Error())
		return "", nil
	} else if err != nil {
		return "", err
	}

	return lastTaskBranch, nil
}

// BranchChangeSendNotice 更换分支部署时发送通知到钉钉群
func (s *Service) BranchChangeSendNotice(ctx context.Context, createReq *req.CreateTaskReq,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp) error {
	env := createReq.EnvName

	if _, ok := app.Env[env]; !ok {
		return errcode.InvalidParams
	}
	// Step: 未开启通知，返回
	if !app.Env[env].EnableBranchChangeNotification {
		return nil
	}

	// Step: 获取分支, 比对分支
	lastTaskBranch, err := s.GetLatestDeploySuccessTaskBranch(ctx, createReq)
	if err != nil {
		return err
	} else if lastTaskBranch == "" {
		return nil
	}
	_, _, currTaskBranch, err := s.ExtraInfoFromImageVersion(createReq.Param.ImageVersion)
	if err != nil {
		return err
	}
	if lastTaskBranch == currTaskBranch {
		return nil
	}

	u, err := url.Parse(project.Team.DingHook)
	if err != nil {
		return err
	}

	queryParams, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return err
	}

	if len(queryParams["access_token"]) > 0 && queryParams["access_token"][0] != "" {
		// TODO: 改成飞书
		// Step: 组装消息内容
		// operator := ctx.Value(utils.ContextUserClaimKey).(entity.UserClaims)
		// msg := &req.BranchChangeRobotMessage{
		// 	Title:            fmt.Sprintf("%s 正在切换分支部署", project.Name),
		// 	ProjectName:      project.Name,
		// 	AppName:          app.Name,
		// 	Env:              string(env),
		// 	OriginBranch:     lastTaskBranch,
		// 	CurrentBranch:    currTaskBranch,
		// 	OperatorName:     operator.Name,
		// 	SecurityKeywords: "阿里云、告警、分支、切换、监控、故障",
		// 	DetailURL:        s.GetAmsFrontendProjectURL(project.ID, createReq.EnvName),
		// }

		// var buf = new(bytes.Buffer)
		// tmpl, err := template.New("robotMarkdownMessage").Parse(branchChangeDeployMsg)
		// if err != nil {
		// 	return errors.Wrapf(errcode.InternalError, "%s", err)
		// }
		// if err = tmpl.Execute(buf, msg); err != nil {
		// 	return errors.Wrapf(errcode.InternalError, "%s", err)
		// }
		// noticeMsg := req.RobotMarkdownMessageReq{
		// 	DingTalkMarkdownMessageReq: req.DingTalkMarkdownMessageReq{
		// 		Msgtype: req.DingTalkMsgTypeMarkdown,
		// 		Markdown: req.DingTalkMarkdownMessageBody{
		// 			Title: msg.Title,
		// 			Text:  buf.String(),
		// 		},
		// 	},
		// 	At: req.RobotMarkdownMessageAt{
		// 		IsAtAll: true,
		// 	},
		// }

		// Step: 发送消息
		// if _, err = s.SendRobotMsgToDingTalk(ctx, &req.SendRobotMessageReq{
		// 	Token:   queryParams["access_token"][0],
		// 	Message: noticeMsg,
		// }); err != nil {
		// 	return err
		// }
	}

	return nil
}

// 生成手动启动Job名称
func (s *Service) getManualJobVersion(version string) (ret string) {
	if len(version) > 6 {
		ret = version[:len(version)-6]
	}
	suffix := fmt.Sprintf("%v", uuid.NewV4())
	ret = ret + "-" + suffix[:5]
	return ret
}

// 获取任务过滤器
func (s *Service) getTasksFilter(_ context.Context, getReq *req.GetTasksReq) bson.M {
	filter := bson.M{}

	if !getReq.Suspend.IsZero() {
		if getReq.Suspend.ValueOrZero() {
			filter["suspend"] = true
		} else {
			// 兼容老数据，需要判断空情况
			filter["suspend"] = bson.M{
				"$in": bson.A{
					primitive.Null{},
					false,
				},
			}
		}
	}

	if getReq.AppID != "" {
		filter["app_id"] = getReq.AppID
	}

	if getReq.Detail != "" {
		filter["detail"] = bson.M{
			"$regex": getReq.Detail,
		}
	}

	if getReq.Version != "" {
		filter["version"] = getReq.Version
	}

	if getReq.EnvName != "" {
		filter["env_name"] = getReq.EnvName
	}

	if getReq.ClusterName != "" {
		filter["cluster_name"] = getReq.ClusterName
	}

	if getReq.Action != "" {
		filter["action"] = getReq.Action
	}

	if getReq.OperatorID != "" {
		filter["operator_id"] = getReq.OperatorID
	}

	if len(getReq.AppIDList) != 0 {
		filter["app_id"] = bson.M{
			"$in": getReq.AppIDList,
		}
	}

	if getReq.Action != "" {
		filter["action"] = getReq.Action
	} else {
		actionFilter := bson.M{}
		if len(getReq.ActionList) != 0 {
			actionFilter["$in"] = getReq.ActionList
		}
		if len(getReq.ActionInverseList) != 0 {
			actionFilter["$nin"] = getReq.ActionInverseList
		}
		if len(actionFilter) > 0 {
			filter["action"] = actionFilter
		}
	}

	statusFilter := bson.M{}
	if len(getReq.StatusList) != 0 {
		statusFilter["$in"] = getReq.StatusList
	}
	if len(getReq.StatusInverseList) != 0 {
		statusFilter["$nin"] = getReq.StatusInverseList
	}
	if len(statusFilter) > 0 {
		filter["status"] = statusFilter
	}

	createTimeFilter := bson.M{}
	if getReq.MinTimestamp != 0 {
		createTimeFilter["$gte"] = time.Unix(int64(getReq.MinTimestamp), 0)
	}
	if getReq.MaxTimestamp != 0 {
		createTimeFilter["$lte"] = time.Unix(int64(getReq.MaxTimestamp), 0)
	}
	if len(createTimeFilter) > 0 {
		filter["create_time"] = createTimeFilter
	}

	approvalStatusFilter := bson.A{primitive.Null{}}
	for _, approvalStatus := range getReq.ApprovalStatusList {
		approvalStatusFilter = append(approvalStatusFilter, approvalStatus)
	}
	if len(getReq.ApprovalStatusList) > 0 {
		filter["approval.status"] = bson.M{
			"$in": approvalStatusFilter,
		}
	}

	// TODO: 可能会导致重复筛选，待前端上线后去掉空值并在状态机加上非部署任务的筛选
	deployTypeFilter := bson.A{primitive.Null{}, ""}
	for _, deployType := range getReq.DeployTypeList {
		deployTypeFilter = append(deployTypeFilter, deployType)
	}
	if len(getReq.DeployTypeList) > 0 {
		filter["deploy_type"] = bson.M{
			"$in": deployTypeFilter,
		}
	}

	if getReq.MinScheduleTime != 0 {
		filter["schedule_time"] = bson.M{
			"$gte": time.Unix(getReq.MinScheduleTime, 0),
		}
	}
	if getReq.MaxScheduleTime != 0 {
		filter["schedule_time"] = bson.M{
			"$lte": time.Unix(getReq.MaxScheduleTime, 0),
		}
	}

	if getReq.ApprovalInstanceID != "" {
		filter["approval.instance_id"] = getReq.ApprovalInstanceID
	}

	if len(getReq.ApprovalType) != 0 {
		filter["approval.type"] = bson.M{
			"$in": getReq.ApprovalType,
		}
	}

	return filter
}

// GetTaskDetail : 获取任务详情
func (s *Service) GetTaskDetail(ctx context.Context, id string) (*resp.TaskDetailResp, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.Wrap(_errcode.InvalidHexStringError, err.Error())
	}

	task, err := s.GetTaskByObjectID(ctx, objectID)
	if err != nil {
		return nil, err
	}

	res := &resp.TaskDetailResp{}
	err = deepcopy.Copy(task).To(res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

func (s *Service) GetTaskByObjectID(ctx context.Context, objectID primitive.ObjectID) (*entity.Task, error) {
	return s.dao.FindSingleTask(ctx, bson.M{"_id": objectID})
}

// GetSingleTask : 获取单个任务
func (s *Service) GetSingleTask(ctx context.Context, getReq *req.GetTasksReq) (*resp.TaskDetailResp, error) {
	getReq.Limit = 1
	getReq.Page = 1

	list, err := s.GetTasks(ctx, getReq)
	if err != nil {
		return nil, err
	}
	if len(list) < 1 {
		return nil, _errcode.NoRequiredTaskError
	}

	return list[0], nil
}

// 获取上次部署成功的任务最后的运行参数(加上hpa)
func (s *Service) GetLatestDeploySuccessTaskFinalVersion(ctx context.Context, getReq *req.GetLatestTaskReq) (*resp.TaskDetailResp, error) {
	// 获取最后deploy成功的任务
	getReq.ActionList = entity.TaskActionInitDeployList
	task, err := s.GetLatestSuccessTask(ctx, getReq)
	if err != nil {
		return nil, err
	}

	getReq.Version = task.Version

	wg := goroutine.New("get-task")
	// Update Pod count when user update HPA
	wg.Go(ctx, fmt.Sprintf("get-hpa-task-%d", time.Now().UnixNano()), func(ctx context.Context) error {
		hpaTask, e := s.GetLatestSuccessTask(ctx, &req.GetLatestTaskReq{
			AppID:       getReq.AppID,
			EnvName:     getReq.EnvName,
			ClusterName: getReq.ClusterName,
			Version:     getReq.Version,
			ActionList:  entity.TaskActionUpdateHPADeployList,
		})
		if e != nil && !errcode.EqualError(_errcode.NoRequiredTaskError, e) {
			return e
		}

		if e == nil && task.CreateTime < hpaTask.CreateTime {
			task.Param.IsAutoScale = true
			task.Param.MinPodCount = hpaTask.Param.MinPodCount
			task.Param.MaxPodCount = hpaTask.Param.MaxPodCount
		}
		return nil
	})
	// Update configCommitID and configURL when user reload config
	wg.Go(ctx, fmt.Sprintf("get-reload-task-%d", time.Now().UnixNano()), func(ctx context.Context) error {
		reloadTask, e := s.GetLatestSuccessTask(ctx, &req.GetLatestTaskReq{
			AppID:       getReq.AppID,
			EnvName:     getReq.EnvName,
			ClusterName: getReq.ClusterName,
			Version:     getReq.Version,
			ActionList: []entity.TaskAction{
				entity.TaskActionReloadConfig,
			},
		})
		if e != nil && !errcode.EqualError(_errcode.NoRequiredTaskError, e) {
			return e
		}
		if e == nil && task.CreateTime < reloadTask.CreateTime {
			task.Param.ConfigCommitID = reloadTask.Param.ConfigCommitID
			task.Param.ConfigURL = reloadTask.Param.ConfigURL
		}
		return nil
	})
	err = wg.Wait()
	if err != nil {
		return nil, err
	}

	return task, nil
}

// GetLatestSuccessTask : 获取最后成功的任务
func (s *Service) GetLatestSuccessTask(ctx context.Context, getReq *req.GetLatestTaskReq) (*resp.TaskDetailResp, error) {
	taskReq := &req.GetTasksReq{
		BaseListRequest: models.BaseListRequest{
			Page:  1,
			Limit: 1,
		},
		AppID:       getReq.AppID,
		EnvName:     getReq.EnvName,
		ClusterName: getReq.ClusterName,
		Version:     getReq.Version,
	}

	if len(getReq.ActionList) > 0 {
		taskReq.ActionList = getReq.ActionList
	}

	if !getReq.IgnoreStatus {
		taskReq.StatusList = entity.TaskStatusSuccessStateList
	}

	res, err := s.GetSingleTask(ctx, taskReq)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetTasks : 获取任务列表
func (s *Service) GetTasks(ctx context.Context, getReq *req.GetTasksReq) ([]*resp.TaskDetailResp, error) {
	filter := s.getTasksFilter(ctx, getReq)

	findOptions := &options.FindOptions{
		Sort: dao.MongoSortByCreateTimeDesc,
	}

	if getReq.Limit != 0 && getReq.Page != 0 {
		limit := int64(getReq.Limit)
		skip := int64(getReq.Page-1) * limit
		findOptions.Limit = &limit
		findOptions.Skip = &skip
	}

	tasks, err := s.dao.FindTasks(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	res := make([]*resp.TaskDetailResp, 0)
	err = deepcopy.Copy(&tasks).To(&res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	for _, detail := range res {
		// 兼容没有设置节点亲和性的历史数据
		// 设置 importance 为默认值 medium
		if detail.Param.NodeAffinityLabelConfig.Importance == "" {
			detail.Param.NodeAffinityLabelConfig.Importance = entity.ApplicationImportanceTypeMedium
		}
	}

	return res, nil
}

// GetTasksCount : 获取任务数量
func (s *Service) GetTasksCount(ctx context.Context, getReq *req.GetTasksReq) (int, error) {
	filter := s.getTasksFilter(ctx, getReq)

	res, err := s.dao.CountTasks(ctx, filter)
	if err != nil {
		return 0, err
	}

	return res, nil
}

// IncreaseTaskRetryCount : 增加任务重试次数
func (s *Service) IncreaseTaskRetryCount(ctx context.Context, id, text string) error {
	err := s.dao.UpdateSingleTask(ctx, id, bson.A{
		bson.M{
			"$set": bson.M{
				"detail": bson.M{
					"$concat": bson.A{
						"$detail", fmt.Sprintf("\n%s", text),
					},
				},
				"update_time": time.Now(),
			},
		},
	})
	if err != nil {
		return err
	}
	err = s.dao.UpdateSingleTask(ctx, id,
		bson.M{
			"$inc": bson.M{
				"retry_count": 1,
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// ConvertToActiveTasksReq : 转换最近活跃任务请求
func (s *Service) ConvertToActiveTasksReq(ctx context.Context, getReq *req.GetActivitiesReq,
	currentUserID string) (*req.GetTasksReq, error) {
	// 获取最近活跃的任务
	// 当前获取最近活跃任务的场景均不应当区分集群
	getTaskReq := &req.GetTasksReq{
		BaseListRequest:   getReq.BaseListRequest,
		OperatorID:        getReq.OperatorID,
		EnvName:           getReq.EnvName,
		Action:            getReq.Action,
		ActionInverseList: entity.TaskActionSystemList,
		MinTimestamp:      getReq.MinTimestamp,
		MaxTimestamp:      getReq.MaxTimestamp,
	}

	appIDs := make([]string, 0)
	projectIDs := make([]string, 0)
	if getReq.IsFav {
		favProjects, err := s.dao.GetFavProjectList(ctx, bson.M{"user_id": currentUserID}, dao.MongoFindOptionWithSortByIDAsc)
		if err != nil {
			return nil, err
		}

		if len(favProjects) == 0 {
			getTaskReq.NoNeedQuery = true
			return getTaskReq, nil
		}

		for _, favProject := range favProjects {
			projectIDs = append(projectIDs, favProject.ProjectID)
		}
	}

	if getReq.ProjectName != "" || getReq.TeamID != "" {
		filter := s.getProjectsFilter(ctx, &req.GetProjectsReq{
			Keyword: getReq.ProjectName,
			TeamID:  getReq.TeamID,
			IDs:     projectIDs,
		})
		projects, err := s.dao.FindProjects(ctx, filter, dao.MongoFindOptionWithSortByIDAsc)
		if err != nil {
			return nil, err
		}

		if len(projects) == 0 {
			getTaskReq.NoNeedQuery = true
			return getTaskReq, nil
		}

		projectIDs = make([]string, 0)

		for _, project := range projects {
			projectIDs = append(projectIDs, project.ID)
		}
	}

	if getReq.ProjectID != "" || getReq.AppName != "" || getReq.AppType != "" || len(projectIDs) > 0 {
		apps, err := s.GetApps(ctx, &req.GetAppsReq{
			ProjectID:  getReq.ProjectID,
			Name:       getReq.AppName,
			Type:       getReq.AppType,
			ProjectIDs: projectIDs,
		})
		if err != nil {
			return nil, err
		}

		if len(apps) == 0 {
			getTaskReq.NoNeedQuery = true
			return getTaskReq, nil
		}

		for _, app := range apps {
			appIDs = append(appIDs, app.ID)
		}
	}

	getTaskReq.AppIDList = appIDs

	return getTaskReq, nil
}

// GetRecentlyActiveTasksCount : 获取最近活跃的任务
func (s *Service) GetRecentlyActiveTasksCount(ctx context.Context, getReq *req.GetTasksReq) (int, error) {
	count, err := s.GetTasksCount(ctx, getReq)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetRecentlyActiveTasks : 获取最近活跃的任务列表
func (s *Service) GetRecentlyActiveTasks(ctx context.Context, getReq *req.GetTasksReq) ([]*resp.ActiveTaskResp, error) {
	tasks, err := s.GetTasks(ctx, getReq)
	if err != nil {
		return nil, err
	}

	// 增加缓存，优化速度
	userCache := make(map[string]*entity.UserAuth)
	appCache := make(map[string]*resp.AppDetailResp)
	projectCache := make(map[string]*resp.ProjectDetailResp)
	res := make([]*resp.ActiveTaskResp, len(tasks))

	for i, task := range tasks {
		user, ok := userCache[task.OperatorID]
		if !ok {
			latest, err := s.dao.FindSingleUserAuth(ctx, bson.M{
				"_id": task.OperatorID,
			})
			if err != nil {
				return nil, err
			}
			userCache[task.OperatorID] = latest
			user = latest
		}

		app, ok := appCache[task.AppID]
		if !ok {
			latest, err := s.GetAppDetail(ctx, task.AppID)
			if err != nil {
				return nil, err
			}
			appCache[task.AppID] = latest
			app = latest
		}

		project, ok := projectCache[app.ProjectID]
		if !ok {
			latest, err := s.GetProjectDetail(ctx, app.ProjectID)
			if err != nil {
				return nil, err
			}
			projectCache[app.ProjectID] = latest
			project = latest
		}

		res[i] = &resp.ActiveTaskResp{
			ID:                task.ID,
			Action:            task.Action,
			Status:            task.Status,
			ActionDisplay:     task.ActionDisplay,
			StatusDisplay:     task.StatusDisplay,
			EnvName:           task.EnvName,
			ClusterName:       task.ClusterName,
			CreateTime:        task.CreateTime,
			Version:           task.Version,
			RetryCount:        task.RetryCount,
			DisplayIcon:       task.DisplayIcon,
			Detail:            task.Detail,
			ProjectID:         app.ProjectID,
			ProjectName:       project.Name,
			ProjectDesc:       project.Desc,
			TeamID:            project.Team.ID,
			TeamName:          project.Team.Name,
			AppID:             task.AppID,
			AppName:           app.Name,
			AppType:           string(app.Type),
			AppServiceType:    string(app.ServiceType),
			OperatorID:        task.OperatorID,
			OperatorName:      user.Name,
			OperatorAvatarURL: user.AvatarURL,
			ImageVersion:      task.Param.ImageVersion,
		}
	}

	return res, nil
}

// GetTasksGroupByAppIDAndEnvName : 通过应用id及环境名分组获取任务
func (s *Service) GetTasksGroupByAppIDAndEnvName(ctx context.Context, getReq *req.GetTasksReq) ([]*resp.TaskDetailResp, error) {
	filter := s.getTasksFilter(ctx, getReq)

	tasks, err := s.dao.FindTasksGroupByAppIDAndEnvName(ctx, filter)
	if err != nil {
		return nil, err
	}

	res := make([]*resp.TaskDetailResp, 0)
	err = deepcopy.Copy(&tasks).To(&res)
	if err != nil {
		return nil, errors.Wrap(errcode.InternalError, err.Error())
	}

	return res, nil
}

// CreateLatestDeployTask : 创建一个新的上一次部署成功的任务
func (s *Service) CreateLatestDeployTask(ctx context.Context, clusterName entity.ClusterName,
	project *resp.ProjectDetailResp, app *resp.AppDetailResp, envName entity.AppEnvName, operatorID string) error {
	task, err := s.GetLatestDeploySuccessTaskFinalVersion(ctx, &req.GetLatestTaskReq{
		AppID:       app.ID,
		EnvName:     envName,
		ClusterName: clusterName,
	})
	if err != nil {
		return err
	}

	if task.ClusterName == "" {
		return errors.Wrapf(errcode.InternalError, "no cluster name in latest task(%s)", task.ID)
	}

	createReq := new(req.CreateTaskReq)
	err = deepcopy.Copy(task).To(createReq)
	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	} else if createReq.Param.ImageVersion == "" {
		return nil
	}

	createReq.Version = ""
	createReq.Action = entity.TaskActionFullDeploy
	createReq.Approval, createReq.DeployType, createReq.ScheduleTime = new(req.ApprovalReq), "", 0

	_, err = s.CreateTask(ctx, project, app, createReq, operatorID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteAppTasks : 删除应用所有的任务
func (s *Service) DeleteAppTasks(ctx context.Context, appID string) error {
	err := s.dao.DeleteTasks(ctx, bson.M{
		"app_id": appID,
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteSingleTask deletes single task
// 原则上只有在等待审批的 task 才可以调 DeleteSingleTask 方法（task.Approval != nil）
func (s *Service) DeleteSingleTask(ctx context.Context, task *resp.TaskDetailResp, operatorID string) error {
	objectID, err := primitive.ObjectIDFromHex(task.ID)
	if err != nil {
		return errors.Wrapf(_errcode.InvalidHexStringError, "%s", err)
	}

	err = s.dao.DeleteSingleTask(ctx, bson.M{
		"_id": objectID,
	})
	if err != nil {
		return err
	}

	if operatorID == entity.K8sSystemUserID || task.Approval == nil {
		return nil
	}

	operator := ctx.Value(utils.ContextInternalUserKey).(entity.InternalUser)

	if task.Approval.Status == entity.ApprovingTaskApprovalStatus {
		if task.Approval.InstanceID == "" {
			return errors.Wrap(_errcode.NoApprovalProcessError, "approval instance id is empty")
		}
		// Terminates ding approval process instance.
		err = s.TerminateDingApprovalInstance(ctx, task.Approval.InstanceID, &req.TerminateDingApprovalInstanceReq{
			OperatingUserID: operator.DingTalkUserID,
		})
		return err
	}

	// send task approval deleted message.
	err = s.sendTaskApprovalDeletedMsg(ctx, task, &operator)
	if err != nil {
		log.Errorc(ctx, "delete task approval err: %v", err)
	}

	return nil
}

// 仅用于参数校验
func (s *Service) ValidateResourceRequirements(requirements *v1.ResourceRequirements) error {
	for resourceName, requestQuantity := range requirements.Requests {
		limitQuantity, exists := requirements.Limits[resourceName]
		if exists && requestQuantity.Cmp(limitQuantity) > 0 {
			return errors.Wrapf(errcode.InvalidParams,
				"%s request must be less than or equal to limit", resourceName)
		}
	}

	return nil
}

// UpdateTask updates task.
func (s *Service) UpdateTask(ctx context.Context, project *resp.ProjectDetailResp, app *resp.AppDetailResp,
	task *resp.TaskDetailResp, updateReq *req.UpdateTaskReq) error {
	err := s.dao.UpdateSingleTask(ctx, task.ID, bson.A{
		bson.M{
			"$set": s.generateUpdateTaskMap(ctx, task, updateReq),
		},
	})
	if err != nil {
		return err
	}

	// User changes manual deployment to immediate, this situation does not need to send message.
	if updateReq.DeployType == "" || task.EnvName != entity.AppEnvPrd || updateReq.OperatorID == entity.K8sSystemUserID ||
		(task.DeployType == entity.ManualTaskDeployType && updateReq.DeployType == entity.ImmediateTaskDeployType) {
		return nil
	}

	if !s.IsPrdP0LevelApp(task.EnvName, project) {
		return nil
	}

	// TODO: 替换成飞书
	// Send message to project owners.
	// err = s.sendTaskDeployTypeAlterationMsg(ctx, project, app, task, updateReq)
	// if err != nil {
	// 	log.Errorc(ctx, "send ding crop message err: %v", err)
	// }

	return nil
}

func (s *Service) generateUpdateTaskMap(_ context.Context, task *resp.TaskDetailResp, updateReq *req.UpdateTaskReq) map[string]interface{} {
	change := make(map[string]interface{})
	change["update_time"] = time.Now()

	if !updateReq.Suspend.IsZero() {
		change["suspend"] = updateReq.Suspend.ValueOrZero()
	}

	if updateReq.DeployType != "" {
		change["deploy_type"] = updateReq.DeployType
		// Set schedule_time to null.
		if updateReq.DeployType != entity.ScheduledTaskDeployType {
			change["schedule_time"] = primitive.Null{}
		}
	}

	if !updateReq.ScheduleTime.IsZero() {
		change["schedule_time"] = time.Unix(updateReq.ScheduleTime.ValueOrZero(), 0)
	}

	if updateReq.ApprovalInstanceID != "" {
		change["approval.instance_id"] = updateReq.ApprovalInstanceID
	}

	if updateReq.ApprovalStatus != "" {
		change["approval.status"] = updateReq.ApprovalStatus
	}

	if updateReq.Status != "" {
		change["status"] = updateReq.Status
	}

	change["detail"] = s.generateUpdateTaskDetail(task, updateReq)

	return change
}

func (s *Service) generateTaskDetailOneLine(desc string) string {
	return fmt.Sprintf("\n[%23s] %s",
		time.Now().Format("2006-01-02 15:04:05.999"), desc)
}

func (s *Service) generateUpdateTaskDetail(task *resp.TaskDetailResp, updateReq *req.UpdateTaskReq) bson.M {
	detail := strings.Builder{}

	if !updateReq.Suspend.IsZero() {
		detail.WriteString(s.generateTaskDetailOneLine(fmt.Sprintf("user:%s updates suspend:%v",
			updateReq.OperatorID, updateReq.Suspend.ValueOrZero())))
	}

	if updateReq.Status != "" {
		detail.WriteString(s.generateTaskDetailOneLine(fmt.Sprintf("enter the status:%s", updateReq.Status)))
	}

	if updateReq.DeployType != "" {
		detail.WriteString(s.generateTaskDetailOneLine(fmt.Sprintf("user:%s updates deploy_type from %s to %s",
			updateReq.OperatorID, task.DeployType, updateReq.DeployType)))
	}

	if updateReq.DeployType == entity.ScheduledTaskDeployType && !updateReq.ScheduleTime.IsZero() {
		detail.WriteString(s.generateTaskDetailOneLine(fmt.Sprintf("user:%s updates schedule_time: %v", updateReq.OperatorID,
			time.Unix(updateReq.ScheduleTime.ValueOrZero(), 0).Format(utils.DefaultTimeFormatLayout))))
	}

	if updateReq.ApprovalStatus != "" {
		detail.WriteString(s.generateTaskDetailOneLine(fmt.Sprintf("ding callback updates approval_status from %s to %s",
			task.Approval.Status, updateReq.ApprovalStatus)))
	}

	return bson.M{
		"$concat": bson.A{
			"$detail", detail.String(),
		},
	}
}

func (s *Service) sendTaskDeployTypeAlterationMsg(ctx context.Context, project *resp.ProjectDetailResp,
	app *resp.AppDetailResp, task *resp.TaskDetailResp, updateReq *req.UpdateTaskReq) error {
	operator, err := s.GetUserInfo(ctx, updateReq.OperatorID)
	if err != nil {
		return err
	}

	emails := make([]string, 0)
	for _, owner := range project.Owners {
		emails = append(emails, owner.Email)
	}

	// Sends ding crop message to owners that the task deployment type is being changed.
	curDeployTime := ""
	if updateReq.DeployType == entity.ScheduledTaskDeployType && !updateReq.ScheduleTime.IsZero() {
		curDeployTime = time.Unix(updateReq.ScheduleTime.ValueOrZero(), 0).Format(utils.DefaultTimeFormatLayout)
	}
	_, err = s.SendTaskDeployAlterationDingCropMessage(ctx, &req.TaskDeployAlterationMessage{
		Title: fmt.Sprintf("%s 的部署方式被 %s 修改为 %s", project.Name,
			operator.Name, entity.GetTaskDeployTypeDisplay(updateReq.DeployType)),
		ProjectName:   project.Name,
		AppName:       app.Name,
		UserName:      operator.Name,
		Env:           string(task.EnvName),
		Version:       task.Version,
		OldDeploy:     entity.GetTaskDeployTypeDisplay(task.DeployType),
		OldDeployTime: task.ScheduleTime,
		CurDeploy:     entity.GetTaskDeployTypeDisplay(updateReq.DeployType),
		CurDeployTime: curDeployTime,
		DetailURL:     s.GetAmsFrontendProjectURL(project.ID, task.EnvName),
	}, emails, taskDeployAlterationTmpName, taskDeployAlterationMsg)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) sendTaskApprovalDeletedMsg(ctx context.Context, task *resp.TaskDetailResp, operator *entity.InternalUser) error {
	app, err := s.GetAppDetail(ctx, task.AppID)
	if err != nil {
		return err
	}
	project, err := s.GetProjectDetail(ctx, app.ProjectID)
	if err != nil {
		return err
	}

	if !s.IsPrdP0LevelApp(task.EnvName, project) {
		return nil
	}

	emails := make([]string, len(project.Owners))
	for _, owner := range project.Owners {
		emails = append(emails, owner.Email)
	}

	// Sends urgent deployment ding crop message to project owners.
	_, err = s.SendDingCropMessage(ctx, &req.AppOpMessage{
		Title:       fmt.Sprintf("%s 被 %s 取消了部署", project.Name, operator.Nickname),
		ProjectName: project.Name,
		AppName:     app.Name,
		Env:         string(task.EnvName),
		Action:      "取消部署",
		OpTime:      time.Now().Format(utils.DefaultTimeFormatLayout),
		UserName:    operator.Nickname,
		DetailURL:   s.GetAmsFrontendProjectURL(project.ID, task.EnvName),
	}, emails, entity.UrgentDeployTmpName, entity.UrgentDeployMsg)
	return err
}

func (s *Service) IsAppExistsARecord(ctx context.Context, app *resp.AppDetailResp, envName entity.AppEnvName) (bool, error) {
	records, err := s.getAliPrivateZoneRecord(ctx, &req.GetQDNSRecordsReq{
		DomainRecordName: s.getAliPrivateZoneK8sDomainName(app.ServiceName, envName),
		PrivateZone:      AliK8sPrivateZone,
		DomainType:       entity.ARecord,
		PageNumber:       1,
		PageSize:         req.GetQDNSRecordsPageSizeLimit,
	})

	if err != nil {
		return false, err
	}

	return len(records) > 0, nil
}

func (s *Service) CheckMultiClusterCreateTask(ctx context.Context, app *resp.AppDetailResp, createReq *req.CreateTaskReq) error {
	multiClusterSupported, err := s.CheckMultiClusterSupport(ctx, createReq.EnvName, app.ProjectID)
	if err != nil {
		return err
	}

	if !multiClusterSupported || app.ServiceType == entity.AppServiceTypeGRPC {
		return nil
	}

	domainName := s.getAliPrivateZoneK8sDomainName(app.ServiceName, createReq.EnvName)

	records, err := s.getAliPrivateZoneRecord(ctx, &req.GetQDNSRecordsReq{
		DomainRecordName: domainName,
		PrivateZone:      AliK8sPrivateZone,
		PageNumber:       1,
		PageSize:         req.GetQDNSRecordsPageSizeLimit,
	})
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	if len(records) != 1 {
		return errors.Wrap(errcode.InvalidParams, "task should have only one private zone record")
	}

	if records[0].Type != entity.CNAMERecord {
		return errors.Wrap(errcode.InvalidParams, "task private zone type should be cname")
	}

	return nil
}
