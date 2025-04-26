package service

import (
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	"rulai/config"
	"rulai/models/entity"
	"rulai/models/resp"
	"rulai/utils"
)

var (
	// ====================
	// >>>请勿删除<<<
	//
	// 用于测试环境的服务变量
	// ====================
	s *Service

	testProject = &resp.ProjectDetailResp{
		ID:           "1449",
		Name:         "ams-app-framework",
		Language:     "Go",
		Desc:         "ams示例",
		APIDocURL:    "",
		DevDocURL:    "",
		Labels:       []string{},
		LogStoreName: "ams-app-framework",
		CreateTime:   time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime:   time.Now().Format(utils.DefaultTimeFormatLayout),
	}
	testTeam = &resp.TeamDetailResp{
		ID:           primitive.NewObjectID().Hex(),
		Name:         "基础架构小组",
		DingHook:     "https://oapi.dingtalk.com/robot/send?access_token=7ee5bd28bd25cbefa70d928a2031c63f91fa553fee32f2996d758c87934c6f9f",
		Label:        "infra",
		AliAlarmName: "infra",
		ExtraDingHooks: map[string]string{
			"ci": "https://oapi.dingtalk.com/robot/send?access_token=7ee5bd28bd25cbefa70d928a2031c63f91fa553fee32f2996d758c87934c6f9f",
		},
	}

	testRestfulServiceApp = &resp.AppDetailResp{
		ID:               primitive.NewObjectID().Hex(),
		Name:             "http",
		Type:             entity.AppTypeService,
		ServiceType:      entity.AppServiceTypeRestful,
		ServiceName:      "ams-app-framework-http",
		AliLogConfigName: "ams-app-framework-http",
		ProjectID:        "1449",
		Env: map[entity.AppEnvName]resp.AppEnvDetailResp{
			"fat": {
				AliAlarmName:    "k8s-ams-app-framework-http-fat",
				LogStoreName:    "ams-app-framework-http-fat",
				ServiceProtocol: entity.LoadBalancerProtocolHTTP,
			},
			"stg": {
				AliAlarmName:    "k8s-ams-app-framework-http-stg",
				LogStoreName:    "ams-app-framework-http-stg",
				ServiceProtocol: entity.LoadBalancerProtocolHTTP,
			},
			"pre": {
				AliAlarmName:    "k8s-ams-app-framework-http-pre",
				LogStoreName:    "ams-app-framework-http-pre",
				ServiceProtocol: entity.LoadBalancerProtocolHTTP,
			},
			"prd": {
				AliAlarmName:    "k8s-ams-app-framework-http-prd",
				LogStoreName:    "ams-app-framework-http-prd",
				ServiceProtocol: entity.LoadBalancerProtocolHTTP,
			},
		},
		CreateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
	}
	testRestfulServiceStartTask = &resp.TaskDetailResp{
		ID:         primitive.NewObjectID().Hex(),
		Version:    "ams-app-framework-http",
		Action:     entity.TaskActionFullDeploy,
		Detail:     "",
		Status:     entity.TaskStatusInit,
		EnvName:    entity.AppEnvStg,
		AppID:      primitive.NewObjectID().Hex(),
		OperatorID: primitive.NewObjectID().Hex(),
		Param: &resp.TaskParamDetailResp{
			ImageVersion:     "crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/infra/app-framework:8befb721-v0.7.0",
			ConfigCommitID:   "c91da6fae5653c09b84e46d5438de3be234f1057",
			ConfigMountPath:  defaultConfigMountPath,
			HealthCheckURL:   "/health",
			IsAutoScale:      true,
			IsSupportMetrics: true,
			Vars: map[string]string{
				"env": string(entity.AppEnvStg),
			},
			PreStopCommand:                "",
			TerminationGracePeriodSeconds: 30,
			CoverCommand:                  "./http",
			TargetPort:                    80,
			CPULimit:                      "1",
			MemLimit:                      "1024Mi",
			CPURequest:                    "0.1",
			MemRequest:                    "256Mi",
			MinPodCount:                   2,
			MaxPodCount:                   4,
			CronCommand:                   "",
			CronParam:                     "",
			ConcurrencyPolicy:             "",
			RestartPolicy:                 "",
			SuccessfulHistoryLimit:        0,
			FailedHistoryLimit:            0,
		},
		CreateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
	}

	testGRPCServiceApp = &resp.AppDetailResp{
		ID:          primitive.NewObjectID().Hex(),
		Name:        "grpc",
		Type:        entity.AppTypeService,
		ServiceType: entity.AppServiceTypeGRPC,
		ServiceName: "ams-app-framework-grpc",
		ProjectID:   "1449",
		Env: map[entity.AppEnvName]resp.AppEnvDetailResp{
			"fat": {
				AliAlarmName:    "k8s-ams-app-framework-grpc-fat",
				LogStoreName:    "ams-app-framework-grpc-fat",
				ServiceProtocol: entity.LoadBalancerProtocolHTTP,
			},
			"stg": {
				AliAlarmName:    "k8s-ams-app-framework-grpc-stg",
				LogStoreName:    "ams-app-framework-grpc-stg",
				ServiceProtocol: entity.LoadBalancerProtocolHTTP,
			},
			"pre": {
				AliAlarmName:    "k8s-ams-app-framework-grpc-pre",
				LogStoreName:    "ams-app-framework-grpc-pre",
				ServiceProtocol: entity.LoadBalancerProtocolHTTP,
			},
			"prd": {
				AliAlarmName:    "k8s-ams-app-framework-grpc-prd",
				LogStoreName:    "ams-app-framework-grpc-prd",
				ServiceProtocol: entity.LoadBalancerProtocolHTTP,
			},
		},
		CreateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
	}
	testGRPCServiceStartTask = &resp.TaskDetailResp{
		ID:         primitive.NewObjectID().Hex(),
		Version:    "ams-app-framework-grpc",
		Action:     entity.TaskActionFullDeploy,
		Detail:     "",
		Status:     entity.TaskStatusInit,
		EnvName:    entity.AppEnvStg,
		AppID:      primitive.NewObjectID().Hex(),
		OperatorID: primitive.NewObjectID().Hex(),
		Param: &resp.TaskParamDetailResp{
			ImageVersion:     "crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/infra/app-framework:cdc53c94-v0.2.9",
			ConfigCommitID:   "2d139fb4dcb6430d34313d597b81d71bd4f294f1",
			ConfigMountPath:  defaultConfigMountPath,
			HealthCheckURL:   "",
			IsAutoScale:      true,
			IsSupportMetrics: true,
			Vars: map[string]string{
				"env":      string(entity.AppEnvStg),
				"APP_NAME": "grpc",
			},
			PreStopCommand:                "",
			TerminationGracePeriodSeconds: 30,
			CoverCommand:                  "",
			TargetPort:                    80,
			CPULimit:                      "1",
			MemLimit:                      "1024Mi",
			CPURequest:                    "0.5",
			MemRequest:                    "512Mi",
			MinPodCount:                   2,
			MaxPodCount:                   4,
			CronCommand:                   "",
			CronParam:                     "",
			ConcurrencyPolicy:             "",
			RestartPolicy:                 "",
			SuccessfulHistoryLimit:        0,
			FailedHistoryLimit:            0,
		},
		CreateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
	}

	testCronJobApp = &resp.AppDetailResp{
		ID:        primitive.NewObjectID().Hex(),
		Name:      "cronjob",
		Type:      entity.AppTypeCronJob,
		ProjectID: "1449",
		Env: map[entity.AppEnvName]resp.AppEnvDetailResp{
			"fat": {
				LogStoreName: "ams-app-framework-cronjob-fat",
			},
			"stg": {
				LogStoreName: "ams-app-framework-cronjob-stg",
			},
			"pre": {
				LogStoreName: "ams-app-framework-cronjob-pre",
			},
			"prd": {
				LogStoreName: "ams-app-framework-cronjob-prd",
			},
		},
		CreateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
	}
	testCronJobStartTask = &resp.TaskDetailResp{
		ID:         primitive.NewObjectID().Hex(),
		Version:    "ams-app-framework-cronjob",
		Action:     entity.TaskActionFullDeploy,
		Detail:     "",
		Status:     entity.TaskStatusInit,
		EnvName:    entity.AppEnvStg,
		AppID:      primitive.NewObjectID().Hex(),
		OperatorID: primitive.NewObjectID().Hex(),
		Param: &resp.TaskParamDetailResp{
			IsSupportMetrics: false,
			Vars: map[string]string{
				"env":      string(entity.AppEnvStg),
				"APP_NAME": "job",
			},
			ImageVersion:           "crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/infra/app-framework:ae4b434f-v0.7.0",
			ConfigCommitID:         "37ef7150093cad8d3d89ab600cab5f1b737b9b53",
			ConfigMountPath:        defaultConfigMountPath,
			CPULimit:               "1",
			MemLimit:               "1024Mi",
			CPURequest:             "0.5",
			MemRequest:             "512Mi",
			CronCommand:            "./job",
			CronParam:              "0/1 * * * *",
			ConcurrencyPolicy:      batchV1.ForbidConcurrent,
			RestartPolicy:          v1.RestartPolicyOnFailure,
			SuccessfulHistoryLimit: 10,
			FailedHistoryLimit:     5,
			ActiveDeadlineSeconds:  60,
		},
		CreateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
	}

	testJobApp = &resp.AppDetailResp{
		ID:        primitive.NewObjectID().Hex(),
		Name:      "job",
		Type:      entity.AppTypeOneTimeJob,
		ProjectID: "1449",
		Env: map[entity.AppEnvName]resp.AppEnvDetailResp{
			"fat": {
				LogStoreName: "ams-app-framework-job-fat",
			},
			"stg": {
				LogStoreName: "ams-app-framework-job-stg",
			},
			"pre": {
				LogStoreName: "ams-app-framework-job-pre",
			},
			"prd": {
				LogStoreName: "ams-app-framework-job-prd",
			},
		},
		CreateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
	}
	testJobStartTask = &resp.TaskDetailResp{
		ID:         primitive.NewObjectID().Hex(),
		Version:    "ams-app-framework-job",
		Action:     entity.TaskActionFullDeploy,
		Detail:     "",
		Status:     entity.TaskStatusInit,
		EnvName:    entity.AppEnvStg,
		AppID:      primitive.NewObjectID().Hex(),
		OperatorID: primitive.NewObjectID().Hex(),
		Param: &resp.TaskParamDetailResp{
			IsSupportMetrics: false,
			Vars: map[string]string{
				"env":      string(entity.AppEnvStg),
				"APP_NAME": "job",
			},
			ImageVersion:          "crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/infra/app-framework:ae4b434f-v0.7.0",
			ConfigCommitID:        "37ef7150093cad8d3d89ab600cab5f1b737b9b53",
			ConfigMountPath:       defaultConfigMountPath,
			CPULimit:              "1",
			MemLimit:              "1024Mi",
			CPURequest:            "0.5",
			MemRequest:            "512Mi",
			JobCommand:            "./job",
			RestartPolicy:         v1.RestartPolicyNever,
			BackoffLimit:          2,
			ActiveDeadlineSeconds: 60,
		},
		CreateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
		UpdateTime: time.Now().Format(utils.DefaultTimeFormatLayout),
	}
)

// ====================
// >>>请勿删除<<<
//
// 测试前准备工作
//
// 常用于初始化操作
// ====================
func beforeTest() {
	// 读取配置
	config.Read("./config/config.yaml")
	// 新建测试所用的数据层
	s = New()
}

// ====================
// >>>请勿删除<<<
//
// 测试后清理工作
//
// 常用于删除测试数据
// ====================
func afterTest() {
}

// ====================
// >>>请勿删除<<<
//
// 进行任意测试时，都会最先进行的测试主函数
// ====================
func TestMain(m *testing.M) {
	// 测试前准备工作
	beforeTest()
	// 进行测试
	code := m.Run()
	// 测试后清洗
	afterTest()
	// 退出
	os.Exit(code)
}
