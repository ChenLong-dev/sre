package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const TaskCanaryVersionSuffix = "-canary"

// TaskAction 任务行为
type TaskAction string

// 任务行为枚举值
const (
	// 全量部署
	TaskActionFullDeploy TaskAction = "full_deploy"
	// 金丝雀部署
	TaskActionCanaryDeploy TaskAction = "canary_deploy"
	// 基于金丝雀的全量部署
	TaskActionFullCanaryDeploy TaskAction = "full_canary_deploy"
	// 停止
	TaskActionStop TaskAction = "stop"
	// 重启
	TaskActionRestart TaskAction = "restart"
	// 恢复
	TaskActionResume TaskAction = "resume"
	// 删除
	TaskActionDelete TaskAction = "delete"
	// 清理
	TaskActionClean TaskAction = "clean"
	// 手动启动cronjob
	TaskActionManualLaunch TaskAction = "manual_launch"
	// 更新hpa
	TaskActionUpdateHPA TaskAction = "update_hpa"
	// 热加载配置
	TaskActionReloadConfig TaskAction = "reload_config"
	// 暂停集群内DNS解析
	TaskActionDisableInClusterDNS TaskAction = "disable_in_cluster_dns"
	// 恢复集群内DNS解析
	TaskActionEnableInClusterDNS TaskAction = "enable_in_cluster_dns"
)

// LaunchType specifies the launch type of a pod
type LaunchType string

const (
	// LaunchTypeManual 手动启动job
	LaunchTypeManual LaunchType = "manual"
)

const (
	// LabelKeyLaunchType 启动方式的label键值
	LabelKeyLaunchType string = "launchType"

	// AnnoatationCronJobInstantiateType CronJob的实例化方式注释
	AnnoatationCronJobInstantiateType string = "cronjob.kubernetes.io/instantiate"
)

// TaskStatus 任务状态
type TaskStatus string

// 任务状态枚举值
const (
	// 初始态
	TaskStatusInit TaskStatus = "init"

	// TaskStatusCreateConfigMapUnderway 创建工作
	// 创建K8s ConfigMap中
	TaskStatusCreateConfigMapUnderway TaskStatus = "create-config_map-underway"
	// TaskStatusCreateConfigMapFinish 创建K8s ConfigMap阶段完成
	TaskStatusCreateConfigMapFinish TaskStatus = "create-config_map-finish"
	// TaskStatusCreateCanaryDeploymentUnderway 创建K8s Canary Deployment中
	TaskStatusCreateCanaryDeploymentUnderway TaskStatus = "create-canary_deployment-underway"
	// TaskStatusCreateCanaryDeploymentFinish 创建K8s Canary Deployment阶段完成
	TaskStatusCreateCanaryDeploymentFinish TaskStatus = "create-canary_deployment-finish"
	// TaskStatusCreateFullDeploymentUnderway 创建K8s Full Deployment中
	TaskStatusCreateFullDeploymentUnderway TaskStatus = "create-full_deployment-underway"
	// TaskStatusCreateFullDeploymentFinish 创建K8s Full Deployment阶段完成
	TaskStatusCreateFullDeploymentFinish TaskStatus = "create-full_deployment-finish"
	// TaskStatusCreateFullCronJobUnderway 创建K8s Full CronJob中
	TaskStatusCreateFullCronJobUnderway TaskStatus = "create-full_cronjob-underway"
	// TaskStatusCreateFullCronJobFinish 创建K8s Full CronJob阶段完成
	TaskStatusCreateFullCronJobFinish TaskStatus = "create-full_cronjob-finish"
	// TaskStatusCreateJobUnderway 创建K8s Job中
	TaskStatusCreateJobUnderway TaskStatus = "create-job-underway"
	// TaskStatusCreateJobFinish 创建K8s Job阶段完成
	TaskStatusCreateJobFinish TaskStatus = "create-job-finish"
	// TaskStatusCreateHPAUnderway 创建K8s HPA中
	TaskStatusCreateHPAUnderway TaskStatus = "create-hpa-underway"
	// TaskStatusCreateHPAFinish 创建K8s HPA阶段完成
	TaskStatusCreateHPAFinish TaskStatus = "create-hpa-finish"
	// TaskStatusCreateCronHPAUnderway 创建定时HPA中
	TaskStatusCreateCronHPAUnderway TaskStatus = "create-cronhpa-underway"
	// TaskStatusCreateCronHPAFinish 创建定时HPA完成
	TaskStatusCreateCronHPAFinish TaskStatus = "create-cronhpa-finish"
	// TaskStatusCreateAliLogConfigUnderway 创建日志配置中
	TaskStatusCreateAliLogConfigUnderway TaskStatus = "create-ali_log_config-underway"
	// TaskStatusCreateAliLogConfigFinish 创建日志配置阶段完成
	TaskStatusCreateAliLogConfigFinish TaskStatus = "create-ali_log_config-finish"
	// TaskStatusCreateLogStoreIndexFinish 创建日志索引完成
	TaskStatusCreateLogStoreIndexFinish TaskStatus = "create-log_store_index-finish"
	// TaskStatusSyncColdStorageDeliverTaskFinish 投递任务同步完成
	TaskStatusSyncColdStorageDeliverTaskFinish TaskStatus = "sync-cold_storage-deliver-task-finish"
	// TaskStatusCreateK8sServiceUnderway 创建K8s Service中
	TaskStatusCreateK8sServiceUnderway TaskStatus = "create-k8s_service-underway"
	// TaskStatusCreateK8sServiceFinish 创建K8s Service阶段完成
	TaskStatusCreateK8sServiceFinish TaskStatus = "create-k8s_service-finish"
	// TaskStatusCreateK8sIngressUnderway 创建K8s Ingress中
	TaskStatusCreateK8sIngressUnderway TaskStatus = "create-k8s_ingress-underway"
	// TaskStatusCreateK8sIngressFinish 创建K8s Ingress阶段完成
	TaskStatusCreateK8sIngressFinish TaskStatus = "create-k8s_ingress-finish"
	// TaskStatusCreateAliRecordUnderway 创建阿里云云解析中
	TaskStatusCreateAliRecordUnderway TaskStatus = "create-ali_record-underway"
	// TaskStatusCreateAliRecordFinish 创建阿里云云解析完成
	TaskStatusCreateAliRecordFinish TaskStatus = "create-ali_record-finish"
	// TaskStatusCreateAliIngressRecordUnderway 创建阿里云ingress云解析中
	TaskStatusCreateAliIngressRecordUnderway TaskStatus = "create-ali_ingress_record-underway"
	// TaskStatusCreateKongObjectsFinish 创建Kong网关路由规则完成
	TaskStatusCreateKongObjectsFinish TaskStatus = "create-kong_objects-finish"
	// TaskStatusAllCreationPhasesFinish 所有创建阶段任务完成
	TaskStatusAllCreationPhasesFinish TaskStatus = "all-creation_phases-finish"
	// TaskStatusCreateVirtualServiceUnderway 创建 VirtualService 中
	TaskStatusCreateVirtualServiceUnderway TaskStatus = "create-k8s_virtualservice-underway"
	// TaskStatusCreateVirtualServiceFinish 创建 VirtualService 完成
	TaskStatusCreateVirtualServiceFinish TaskStatus = "create-k8s_virtualservice-finish"

	// 热加载配置
	// 更新k8s ConfigMap中
	TaskStatusUpdateConfigMapUnderway TaskStatus = "update-config_map-underway"
	// 更新k8s ConfigMap完成
	TaskStatusUpdateConfigMapFinish TaskStatus = "update-config_map-finish"
	// 更新k8s Deployment注解中
	TaskStatusUpdateDeploymentAnnotationUnderway TaskStatus = "update-deployment-annotation-underway"
	// 更新k8s Deployment注解完成
	TaskStatusUpdateDeploymentAnnotationFinish TaskStatus = "update-deployment-annotation-finish"
	// TaskStatusUpdateKongObjectsUnderway 更新 kong_objects 中
	TaskStatusUpdateKongUpstreamUnderway TaskStatus = "update-kong_objects-underway"
	// TaskStatusUpdateKongObjectsFinish 更新 kong_objects 完成
	TaskStatusUpdateKongUpstreamFinish TaskStatus = "update-kong_objects-finish"

	// TaskStatusCleanDeploymentUnderway 清理工作
	// 清理K8s Deployment中
	TaskStatusCleanDeploymentUnderway TaskStatus = "clean-deployment-underway"
	// TaskStatusCleanDeploymentFinish 清理K8s Deployment阶段完成
	TaskStatusCleanDeploymentFinish TaskStatus = "clean-deployment-finish"
	// TaskStatusCleanCronJobUnderway 清理K8s CronJob中
	TaskStatusCleanCronJobUnderway TaskStatus = "clean-cronjob-underway"
	// TaskStatusCleanCronJobFinish 清理K8s CronJob阶段完成
	TaskStatusCleanCronJobFinish TaskStatus = "clean-cronjob-finish"
	// TaskStatusCleanJobUnderway 清理K8s Job中
	TaskStatusCleanJobUnderway TaskStatus = "clean-job-underway"
	// TaskStatusCleanJobFinish 清理K8s Job阶段完成
	TaskStatusCleanJobFinish TaskStatus = "clean-job-finish"

	// TaskStatusCleanCronHPAUnderway 清理k8s cronHPA中
	TaskStatusCleanCronHPAUnderway TaskStatus = "clean-cronhpa-underway"
	// TaskStatusCleanCronHPAFinish 清理k8s cronHPA结束
	TaskStatusCleanCronHPAFinish TaskStatus = "clean-cronhpa-finish"

	// TaskStatusCleanHPAUnderway 清理K8s HPA中
	TaskStatusCleanHPAUnderway TaskStatus = "clean-hpa-underway"
	// TaskStatusCleanHPAFinish 清理K8s HPA阶段完成
	TaskStatusCleanHPAFinish TaskStatus = "clean-hpa-finish"
	// TaskStatusCleanK8sServiceUnderway 清理K8s Service中
	TaskStatusCleanK8sServiceUnderway TaskStatus = "clean-k8s_service-underway"
	// TaskStatusCleanK8sServiceFinish 清理K8s Service阶段完成
	TaskStatusCleanK8sServiceFinish TaskStatus = "clean-k8s_service-finish"
	// TaskStatusCleanConfigMapUnderway 清理K8s ConfigMap中
	TaskStatusCleanConfigMapUnderway TaskStatus = "clean-config_map-underway"
	// TaskStatusCleanConfigMapFinish 清理K8s ConfigMap阶段完成
	TaskStatusCleanConfigMapFinish TaskStatus = "clean-config_map-finish"
	// TaskStatusCleanK8sIngressUnderway 清理K8s Ingress中
	TaskStatusCleanK8sIngressUnderway TaskStatus = "clean-k8s_ingress-underway"
	// TaskStatusCleanK8sIngressFinish 清理K8s Ingress阶段完成
	TaskStatusCleanK8sIngressFinish TaskStatus = "clean-k8s_ingress-finish"
	// TaskStatusCleanAliServiceFinish 清理云服务相关阶段完成
	TaskStatusCleanAliServiceFinish TaskStatus = "clean-ali_service-finish"
	// TaskStatusCleanKongRecordFinish 清理 Kong 记录阶段完成
	TaskStatusCleanKongRecordFinish TaskStatus = "clean-kong_record-finish"
	// TaskStatusCleanAliLogConfigUnderway 清理日志配置中
	TaskStatusCleanAliLogConfigUnderway TaskStatus = "clean-ali_log_config-underway"
	// TaskStatusCleanAliLogConfigFinish 清理日志配置阶段完成
	TaskStatusCleanAliLogConfigFinish TaskStatus = "clean-ali_log_config-finish"
	// TaskStatusCleanAliLogStoreFinish 清理日志相关资源完成
	TaskStatusCleanAliLogStoreFinish TaskStatus = "clean-ali_log_store-finish"
	// TaskStatusCleanVirtualServiceUnderWay 清理 VirtualService资源
	TaskStatusCleanVirtualServiceFinish TaskStatus = "clean-istio-virtualservice"
	// TaskStatusCleanVirtualServiceUnderWay 清理 VirtualService 资源完成
	TaskStatusCleanVirtualServiceUnderWay TaskStatus = "clean-istio-virtualservice-underway"

	// TaskStatusUpdateDeploymentScaleUnderway 其他工作
	// 更新K8s Deployment实例数中
	TaskStatusUpdateDeploymentScaleUnderway TaskStatus = "update-deployment_scale-underway"
	// TaskStatusUpdateDeploymentScaleFinish 更新K8s Deployment实例数阶段完成
	TaskStatusUpdateDeploymentScaleFinish TaskStatus = "update-deployment_scale-finish"
	// TaskStatusUpdateCronJobSuspendUnderway 更新K8s CronJob暂停状态中
	TaskStatusUpdateCronJobSuspendUnderway TaskStatus = "update-cronjob_suspend-underway"
	// TaskStatusUpdateCronJobSuspendFinish 更新K8s CronJob暂停状态阶段完成
	TaskStatusUpdateCronJobSuspendFinish TaskStatus = "update-cronjob_suspend-finish"
	// TaskStatusRestartDeploymentUnderway 重启K8s Deployment中
	TaskStatusRestartDeploymentUnderway TaskStatus = "restart-deployment-underway"
	// TaskStatusRestartDeploymentFinish 重启K8s Deployment阶段完成
	TaskStatusRestartDeploymentFinish TaskStatus = "restart-deployment-finish"
	// TaskStatusUpdateHPAUnderway 更新K8s HPA中
	TaskStatusUpdateHPAUnderway TaskStatus = "update-hpa-underway"
	// TaskStatusUpdateHPAFinish 更新K8s HPA阶段完成
	TaskStatusUpdateHPAFinish TaskStatus = "update-hpa-finish"

	// TaskStatusDeleteAliRecordUnderway 删除阿里云云解析中
	TaskStatusDeleteAliRecordUnderway TaskStatus = "delete-ali_record-underway"
	// TaskStatusDeleteAliRecordFinish 删除阿里云云解析完成
	TaskStatusDeleteAliRecordFinish TaskStatus = "delete-ali_record-finish"

	// TaskStatusUpdateInClusterDNSUnderway 集群内DNS解析更新中
	// 1. 等待dns生效
	// 2. 等待ingress生效
	TaskStatusUpdateInClusterDNSUnderway TaskStatus = "update-incluster-dns-underway"
	// TaskStatusUpdateInClusterDNSFinish 集群内DNS解析更新完成
	TaskStatusUpdateInClusterDNSFinish TaskStatus = "update-incluster-dns-finish"
	// TaskStatusSuccess 最终态
	// 成功
	TaskStatusSuccess TaskStatus = "success"
	// TaskStatusFail 失败
	TaskStatusFail TaskStatus = "fail"
)

// TaskDisplayIcon 任务展示的icon
type TaskDisplayIcon string

// 任务展示 icon 枚举值
const (
	// TaskDisplayIconSuccess 成功
	TaskDisplayIconSuccess TaskDisplayIcon = "success"
	// TaskDisplayIconFail 失败
	TaskDisplayIconFail TaskDisplayIcon = "fail"
	// TaskDisplayIconUnderway 进行中
	TaskDisplayIconUnderway TaskDisplayIcon = "underway"
	// TaskDisplayIconCanaryContinue 金丝雀继续
	TaskDisplayIconCanaryContinue TaskDisplayIcon = "canary_continue"
	// TaskDisplayIconResume 恢复部署
	TaskDisplayIconResume TaskDisplayIcon = "resume"
)

// 限制类常量
const (
	// TaskMaxRetryCount 任务最大的重试次数
	TaskMaxRetryCount = 10

	// TaskExecuteTimeout 任务执行超时时间
	TaskExecuteTimeout = time.Minute * 10

	// MaxBackOffLimit CronJob最大重试次数
	MaxBackOffLimit = 10
)

var (
	// TaskActionInitDeployList 任务初始部署的行为列表
	TaskActionInitDeployList = []TaskAction{
		TaskActionFullDeploy,
		TaskActionCanaryDeploy,
	}
	// TaskActionSystemList 系统行为列表
	TaskActionSystemList = []TaskAction{
		TaskActionClean,
	}
	// TaskStatusSuccessStateList 任务成功状态列表
	TaskStatusSuccessStateList = []TaskStatus{
		TaskStatusSuccess,
	}
	// TaskStatusFinalStateList 任务终态列表
	TaskStatusFinalStateList = []TaskStatus{
		TaskStatusSuccess,
		TaskStatusFail,
	}
	// TaskStatusFinalAndInitList 任务终止和初始态列表
	TaskStatusFinalAndInitList = []TaskStatus{
		TaskStatusSuccess,
		TaskStatusFail,
		TaskStatusInit,
	}
	// TaskActionUpdateHPADeployList 任务更新部署hpa的行为列表
	TaskActionUpdateHPADeployList = []TaskAction{
		TaskActionUpdateHPA,
	}
	// TaskActionPodNumberRelatedList 与 pod 数量变更相关的行为列表
	TaskActionPodNumberRelatedList = []TaskAction{
		TaskActionCanaryDeploy,
		TaskActionFullDeploy,
		TaskActionUpdateHPA,
	}
	// TaskActionServiceBatchList Service批量操作相关行为列表
	TaskActionServiceBatchList = []TaskAction{
		TaskActionFullDeploy,
		TaskActionCanaryDeploy,
		TaskActionStop,
		TaskActionRestart,
		TaskActionDelete,
		TaskActionResume,
	}
	// TaskActionWorkerBatchList Worker批量操作相关行为列表
	TaskActionWorkerBatchList = []TaskAction{
		TaskActionFullDeploy,
		TaskActionCanaryDeploy,
		TaskActionStop,
		TaskActionRestart,
		TaskActionDelete,
		TaskActionResume,
	}
	// TaskActionCronJobBatchList CronJob批量操作相关行为列表
	TaskActionCronJobBatchList = []TaskAction{
		TaskActionFullDeploy,
		TaskActionStop,
		TaskActionDelete,
		TaskActionResume,
	}
	// TaskActionOneTimeJobBatchList OneTimeJob批量操作相关行为列表
	TaskActionOneTimeJobBatchList = []TaskAction{
		TaskActionFullDeploy,
		TaskActionDelete,
	}
)

// TaskActionList the list of action
type TaskActionList []TaskAction

// Contains wether the list contains specified action
func (list *TaskActionList) Contains(action TaskAction) bool {
	for _, item := range *list {
		if item == action {
			return true
		}
	}

	return false
}

var (
	// TaskActionInClusterDNSList DNS相关的操作
	TaskActionInClusterDNSList = &TaskActionList{
		TaskActionEnableInClusterDNS,
		TaskActionDisableInClusterDNS,
	}
)

// Task 任务
type Task struct {
	// 任务id
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// 版本
	Version string `bson:"version" json:"version"`
	// 任务类型
	Action TaskAction `bson:"action" json:"action"`
	// 部署类型
	DeployType TaskDeployType `bson:"deploy_type" json:"deploy_type"`
	// 自动部署时间
	ScheduleTime *time.Time `bson:"schedule_time" json:"schedule_time"`
	// 审批信息
	Approval *Approval `bson:"approval" json:"approval"`
	// 详情
	Detail string `bson:"detail" json:"detail"`
	// 描述信息
	Description string `bson:"description" json:"description"`
	// 状态
	Status TaskStatus `bson:"status" json:"status"`
	// 重试次数
	RetryCount int `bson:"retry_count" json:"retry_count"`
	// 集群名
	ClusterName ClusterName `bson:"cluster_name" json:"cluster_name"`
	// 环境名
	EnvName AppEnvName `bson:"env_name" json:"env_name"`
	// 应用id
	AppID string `bson:"app_id" json:"app_id"`
	// 操作人id
	OperatorID string `bson:"operator_id" json:"operator_id"`
	// 任务参数
	Param *TaskParam `bson:"param" json:"param"`
	// 命名空间
	Namespace string `bson:"namespace" json:"namespace"`
	// 是否暂停
	Suspend    bool       `bson:"suspend" json:"suspend"`
	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
	// 软删除
	DeleteTime *time.Time `bson:"delete_time" json:"delete_time"`
}

// TableName 获取任务表名
func (*Task) TableName() string {
	return "task"
}

// GenerateObjectIDString 生成任务主键
func (t *Task) GenerateObjectIDString(args map[string]interface{}) string {
	return t.ID.Hex()
}

// ActionDisplay 获取任务行为展示信息
func (t *Task) ActionDisplay(args map[string]interface{}) string {
	return GetTaskActionDisplay(t.Action)
}

// StatusDisplay 获取任务状态展示信息
func (t *Task) StatusDisplay(args map[string]interface{}) string {
	return GetTaskStatusDisplay(t.Action, t.Status, t.Suspend)
}

// DisplayIcon 获取任务 icon 展示信息
func (t *Task) DisplayIcon(args map[string]interface{}) string {
	return string(GetTaskDisplayIcon(t.Action, t.Status, t.Suspend))
}

// 审批信息
type Approval struct {
	// 审批流类型
	Type TaskApprovalType `bson:"type" json:"type"`
	// 审批状态
	Status TaskApprovalStatus `bson:"status" json:"status"`
	// 测试工程师
	QAEngineers []*DingDingUserDetail `bson:"qa_engineers" json:"qa_engineers"`
	// 运维工程师
	OperationEngineers []*DingDingUserDetail `bson:"operation_engineers" json:"operation_engineers"`
	// 产品经理
	ProductManagers []*DingDingUserDetail `bson:"product_managers" json:"product_managers"`
	// 审批实例id
	InstanceID string `bson:"instance_id" json:"instance_id"`
}

// GetTaskActionDisplay 获取任务用于展示的行为名
func GetTaskActionDisplay(action TaskAction) string {
	switch action {
	case TaskActionFullDeploy:
		return "全量部署"
	case TaskActionFullCanaryDeploy:
		return "基于金丝雀的全量部署"
	case TaskActionCanaryDeploy:
		return "金丝雀部署"
	case TaskActionStop:
		return "停止"
	case TaskActionRestart:
		return "重启"
	case TaskActionResume:
		return "恢复"
	case TaskActionDelete:
		return "删除"
	case TaskActionClean:
		return "清理"
	case TaskActionManualLaunch:
		return "手动启动"
	case TaskActionUpdateHPA:
		return "弹性伸缩"
	case TaskActionReloadConfig:
		return "热加载配置"
	case TaskActionDisableInClusterDNS:
		return "禁用集群DNS解析"
	case TaskActionEnableInClusterDNS:
		return "启用集群DNS解析"
	default:
		return "未知"
	}
}

// GetTaskStatusDisplay 获取任务用于展示的状态名
func GetTaskStatusDisplay(action TaskAction, status TaskStatus, suspend bool) string {
	if suspend {
		return "已暂停"
	}
	switch status {
	case TaskStatusInit:
		return "初始化中"
	case TaskStatusCreateConfigMapUnderway:
		return "应用配置创建中"
	case TaskStatusCreateConfigMapFinish:
		return "应用配置创建阶段结束"
	case TaskStatusCreateCanaryDeploymentUnderway:
		return "金丝雀部署创建中"
	case TaskStatusCreateCanaryDeploymentFinish:
		return "金丝雀部署创建阶段结束"
	case TaskStatusCreateFullDeploymentUnderway:
		return "全量部署创建中"
	case TaskStatusCreateFullDeploymentFinish:
		return "全量部署创建阶段结束"
	case TaskStatusCreateFullCronJobUnderway:
		return "全量部署创建中"
	case TaskStatusCreateFullCronJobFinish:
		return "全量部署创建阶段结束"
	case TaskStatusCreateJobUnderway:
		return "全量部署创建中"
	case TaskStatusCreateJobFinish:
		return "全量部署创建阶段结束"
	case TaskStatusCreateHPAUnderway:
		return "HPA创建中"
	case TaskStatusCreateHPAFinish:
		return "HPA创建阶段结束"
	case TaskStatusCreateK8sIngressUnderway:
		return "K8sIngress创建中"
	case TaskStatusCreateK8sIngressFinish:
		return "K8sIngress创建阶段完成"
	case TaskStatusCreateAliLogConfigUnderway:
		return "日志配置创建中"
	case TaskStatusCreateAliLogConfigFinish:
		return "日志配置创建阶段结束"
	case TaskStatusCreateLogStoreIndexFinish:
		return "日志索引创建完成"
	case TaskStatusSyncColdStorageDeliverTaskFinish:
		return "日志冷存投递任务同步完成"
	case TaskStatusCreateK8sServiceUnderway:
		return "K8s服务创建中"
	case TaskStatusCreateK8sServiceFinish:
		return "K8s服务创建阶段完成"
	case TaskStatusCleanDeploymentUnderway:
		return "清理部署中"
	case TaskStatusCleanDeploymentFinish:
		return "清理部署阶段完成"
	case TaskStatusCleanCronJobUnderway:
		return "清理任务中"
	case TaskStatusCleanCronJobFinish:
		return "清理任务阶段完成"
	case TaskStatusCleanJobUnderway:
		return "清理脚本中"
	case TaskStatusCleanJobFinish:
		return "清理脚本阶段完成"
	case TaskStatusCleanHPAUnderway:
		return "清理HPA中"
	case TaskStatusCleanHPAFinish:
		return "清理HPA阶段完成"
	case TaskStatusCreateCronHPAUnderway:
		return "cronHPA创建中"
	case TaskStatusCreateCronHPAFinish:
		return "cronHPA创建阶段结束"
	case TaskStatusCleanCronHPAUnderway:
		return "cronHPA清理中"
	case TaskStatusCleanCronHPAFinish:
		return "cronHPA清理阶段结束"
	case TaskStatusCleanK8sServiceUnderway:
		return "清理K8s服务中"
	case TaskStatusCleanK8sServiceFinish:
		return "清理K8s服务阶段完成"
	case TaskStatusCleanAliServiceFinish:
		return "清理云服务阶段完成"
	case TaskStatusCleanKongRecordFinish:
		return "清理Kong记录阶段完成"
	case TaskStatusCleanK8sIngressUnderway:
		return "清理K8sIngress阶段中"
	case TaskStatusCleanK8sIngressFinish:
		return "清理K8sIngress阶段完成"
	case TaskStatusCleanVirtualServiceUnderWay:
		return "清理 VirtualService 阶段中"
	case TaskStatusCleanVirtualServiceFinish:
		return "清理 VirtualService 阶段完成"
	case TaskStatusCleanConfigMapUnderway:
		return "清理应用配置中"
	case TaskStatusCleanConfigMapFinish:
		return "清理应用配置阶段完成"
	case TaskStatusCleanAliLogConfigUnderway:
		return "清理日志配置中"
	case TaskStatusCleanAliLogConfigFinish:
		return "清理日志配置完成"
	case TaskStatusCleanAliLogStoreFinish:
		return "清理日志相关资源完成"
	case TaskStatusUpdateDeploymentScaleUnderway:
		return "更新部署实例数中"
	case TaskStatusUpdateDeploymentScaleFinish:
		return "更新部署实例数阶段完成"
	case TaskStatusUpdateCronJobSuspendUnderway:
		return "更新任务暂停状态中"
	case TaskStatusUpdateHPAUnderway:
		return "HPA更新中"
	case TaskStatusUpdateHPAFinish:
		return "HPA更新阶段结束"
	case TaskStatusUpdateCronJobSuspendFinish:
		return "更新任务暂停状态阶段完成"
	case TaskStatusAllCreationPhasesFinish:
		return "所有创建阶段任务完成"
	case TaskStatusCreateAliRecordUnderway:
		return "创建阿里云云解析中"
	case TaskStatusCreateAliRecordFinish:
		return "创建阿里云云解析完成"
	case TaskStatusDeleteAliRecordUnderway:
		return "删除阿里云云解析中"
	case TaskStatusDeleteAliRecordFinish:
		return "删除阿里云云解析完成"
	case TaskStatusCreateAliIngressRecordUnderway:
		return "创建阿里云ingress云解析中"
	case TaskStatusCreateKongObjectsFinish:
		return "创建Kong网关路由规则完成"
	case TaskStatusRestartDeploymentUnderway:
		return "重启部署中"
	case TaskStatusRestartDeploymentFinish:
		return "重启部署阶段完成"
	case TaskStatusSuccess:
		if action == TaskActionStop {
			return "暂停成功"
		} else if action == TaskActionCanaryDeploy {
			return "金丝雀成功"
		}
		return "成功"
	case TaskStatusFail:
		return "失败"
	case TaskStatusUpdateConfigMapUnderway:
		return "更新应用配置中"
	case TaskStatusUpdateConfigMapFinish:
		return "应用配置更新完成"
	case TaskStatusUpdateDeploymentAnnotationUnderway:
		return "更新部署实例中"
	case TaskStatusUpdateDeploymentAnnotationFinish:
		return "部署实例更新完成"
	case TaskStatusUpdateInClusterDNSUnderway:
		return "集群DNS更新中"
	case TaskStatusUpdateInClusterDNSFinish:
		return "集群DNS完成"
	case TaskStatusCreateVirtualServiceUnderway:
		return "创建 VirtualService 中"
	case TaskStatusCreateVirtualServiceFinish:
		return "创建 VirtualService 完成"
	case TaskStatusUpdateKongUpstreamUnderway:
		return "更新 Kong upstream 中"
	case TaskStatusUpdateKongUpstreamFinish:
		return "更新 Kong upstream 完成"
	default:
		return "未知"
	}
}

// GetTaskDisplayIcon 获取任务用于展示的icon
func GetTaskDisplayIcon(action TaskAction, status TaskStatus, suspend bool) TaskDisplayIcon {
	if suspend {
		return TaskDisplayIconSuccess
	}
	if status == TaskStatusSuccess {
		if action == TaskActionStop {
			return TaskDisplayIconResume
		} else if action == TaskActionCanaryDeploy {
			return TaskDisplayIconCanaryContinue
		}
		return TaskDisplayIconSuccess
	} else if status == TaskStatusFail {
		return TaskDisplayIconFail
	} else {
		return TaskDisplayIconUnderway
	}
}

// Task deploy type.
type TaskDeployType string

const (
	// Manual deploy type.
	ManualTaskDeployType TaskDeployType = "manual"
	// Scheduled deploy type.
	ScheduledTaskDeployType TaskDeployType = "scheduled"
	// immediate deploy type.
	ImmediateTaskDeployType TaskDeployType = "immediate"
)

var (
	TaskDeployTypeImmediateList = []TaskDeployType{ImmediateTaskDeployType}
	TaskDeployTypeScheduledList = []TaskDeployType{ScheduledTaskDeployType}
)

// Task deploy type display.
func GetTaskDeployTypeDisplay(deployType TaskDeployType) string {
	switch deployType {
	case ManualTaskDeployType:
		return "手动部署"
	case ImmediateTaskDeployType:
		return "立即部署"
	case ScheduledTaskDeployType:
		return "定时部署"
	default:
		return "未知"
	}
}

// Task approval status.
type TaskApprovalStatus string

const (
	// Approving approval status.
	ApprovingTaskApprovalStatus TaskApprovalStatus = "approving"
	// Approved approval status.
	ApprovedTaskApprovalStatus TaskApprovalStatus = "approved"
	// Refused approval status.
	RefusedTaskApprovalStatus TaskApprovalStatus = "refused"
)

var (
	TaskApprovalStatusApprovingList = []TaskApprovalStatus{ApprovingTaskApprovalStatus}
	TaskApprovalStatusApprovedList  = []TaskApprovalStatus{ApprovedTaskApprovalStatus}
)

// Approval type.
type TaskApprovalType string

const (
	// Default approval type.
	DefaultTaskApprovalType TaskApprovalType = "default"
	// Skip approval type.
	SkipTaskApprovalType TaskApprovalType = "skip"
)
