package req

import (
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	"gitlab.shanhai.int/sre/library/base/null"

	"rulai/models"
	"rulai/models/entity"
)

// CreateTaskReq : 创建任务请求
type CreateTaskReq struct {
	AppID        string                `json:"app_id" binding:"required"`
	Version      string                `json:"version"`
	OperatorID   string                `json:"operator_id"`
	Description  string                `json:"description"`
	ClusterName  entity.ClusterName    `json:"cluster_name" binding:"required"`
	EnvName      entity.AppEnvName     `json:"env_name" binding:"required"`
	Action       entity.TaskAction     `json:"action" binding:"required"`
	DeployType   entity.TaskDeployType `json:"deploy_type"`
	ScheduleTime int64                 `json:"schedule_time"`
	Param        *CreateTaskParamReq   `json:"param" binding:"required"`
	Approval     *ApprovalReq          `json:"approval"`
	// 部署时忽略分支不匹配异常（强制部署）
	IgnoreExpectedBranch bool   `json:"ignore_expected_branch"`
	Namespace            string `json:"namespace" default:"stg"` // 创建任务时,携带命名空间
}

// CreateTaskParamReq : 创建任务参数请求
type CreateTaskParamReq struct {
	// 镜像地址
	ImageVersion string `json:"image_version"`
	// 配置文件CommitID
	ConfigCommitID string `json:"config_commit_id"`
	// ConfigRenamePrefix 特殊配置重命名前缀
	ConfigRenamePrefix string `json:"config_rename_prefix"`
	// ConfigRenameMode 特殊配置重命名模式, binding 注意与 entity 的枚举值同步
	ConfigRenameMode entity.ConfigRenameMode `json:"config_rename_mode" binding:"omitempty"`
	// 配置文件挂载路径
	ConfigMountPath string `json:"config_mount_path"`
	// create oss storage to store log no time limits
	OpenColdStorage bool `json:"open_cold_storage"`
	// 需要的CPU
	CPURequest entity.CPUResourceType `json:"cpu_request"`
	// 需要的内存
	MemRequest entity.MemResourceType `json:"mem_request"`
	// 限制的最大CPU
	CPULimit entity.CPUResourceType `json:"cpu_limit"`
	// 限制的最大内存
	MemLimit entity.MemResourceType `json:"mem_limit"`
	// 环境变量
	Vars map[string]string `json:"vars"`
	// 是否支持监控
	IsSupportMetrics bool `json:"is_support_metrics"`
	// 监控端口
	MetricsPort int `json:"metrics_port"`
	// 宽限终止时长
	TerminationGracePeriodSeconds entity.TerminationGracePeriodSpan `json:"termination_grace_period_sec"`
	// 存活探针初始化延迟时长
	LivenessProbeInitialDelaySeconds entity.ProbeDelaySpan `json:"liveness_probe_initial_delay_seconds"`
	// 可读探针初始化延迟时长
	ReadinessProbeInitialDelaySeconds entity.ProbeDelaySpan `json:"readiness_probe_initial_delay_seconds"`

	// Service&Worker类型需要
	// 是否自动扩缩容
	IsAutoScale bool `json:"is_auto_scale"`
	// 预停止命令
	PreStopCommand string `json:"pre_stop_command"`
	// 覆盖命令
	CoverCommand string `json:"cover_command"`
	// 最小实例数
	MinPodCount int `json:"min_pod_count"`
	// 最大实例数
	MaxPodCount int `json:"max_pod_count"`
	// Service类型需要以下参数
	// 健康检查地址
	HealthCheckURL string `json:"health_check_url"`
	// 暴露的主要端口号
	TargetPort int `json:"target_port"`
	// 暴露的额外端口
	ExposedPorts map[string]int `json:"exposed_ports"`
	// 节点亲和性标签配置
	NodeAffinityLabelConfig NodeAffinityLabelConfig `json:"node_affinity_label_config" binding:"omitempty"`
	// Deprecated: 节点选择器(nodeSelector)按照官方建议弃用，但先保留字段做向后兼容
	NodeSelector map[string]string `json:"node_selector" binding:"omitempty"`
	// 是否支持会话保持
	IsSupportStickySession bool `json:"is_support_sticky_session"`
	// 会话保持cookie过期时间 单位秒
	SessionCookieMaxAge int `json:"session_cookie_max_age"`
	// 关闭pod反亲和性
	DisableHighAvailability bool `json:"disable_high_availability"`
	// 关闭金丝雀发布(仅存储)
	DisableCanary bool `json:"disable_canary"`

	// 如果应用服务类型是 Restful，且服务暴露方式是 LB，该字段才会有值
	LoadBalancerID string `json:"load_balancer_id"`
	// 7层LB的https证书ID
	LoadBalancerCertID string `json:"load_balancer_cert_id"`

	// CronJob需要以下参数
	// 调度命令
	CronCommand string `json:"cron_command"`
	// 调度周期，Cron表达式
	CronParam string `json:"cron_param"`
	// 并发策略
	ConcurrencyPolicy batchV1.ConcurrencyPolicy `json:"concurrency_policy"`
	// 重启策略
	RestartPolicy v1.RestartPolicy `json:"restart_policy"`
	// 成功的历史限制
	SuccessfulHistoryLimit int `json:"successful_history_limit"`
	// 失败的历史限制
	FailedHistoryLimit int `json:"failed_history_limit"`
	// 任务超时时间（job也需要）
	ActiveDeadlineSeconds int `json:"active_deadline_seconds"`

	// Job需要以下参数
	// CronJob需要以下参数
	// 任务失败重试次数
	BackoffLimit int `json:"backoff_limit"`
	// 脚本执行命令
	JobCommand string `json:"job_command"`

	// 定时扩缩容任务组列表
	CronScaleJobGroups []*entity.CronScaleJobGroup `json:"cron_scale_job_groups"`
	// ScaleJobExcludeDate 定时扩缩容排除日期，五位时间模板，最小粒度为"天"，更小粒度填充"*"
	CronScaleJobExcludeDates []string `json:"cron_scale_job_exclude_dates"`

	// 用于清理工作
	CleanedProjectName          string                      `json:"cleaned_project_name,omitempty"`
	CleanedAppName              string                      `json:"cleaned_app_name,omitempty"`
	CleanedAppType              entity.AppType              `json:"cleaned_app_type,omitempty"`
	CleanedAppServiceType       entity.AppServiceType       `json:"cleaned_app_service_type,omitempty"`
	CleanedAppServiceExposeType entity.AppServiceExposeType `json:"cleaned_app_service_expose_type,omitempty"`
	CleanedServiceName          string                      `json:"cleaned_service_name,omitempty"`
	CleanedAliAlarmName         string                      `json:"cleaned_ali_alarm_name,omitempty"`
	CleanedAliLogConfigName     string                      `json:"cleaned_ali_log_config_name,omitempty"`
	CleanedAliLogStoreName      string                      `json:"cleaned_ali_log_store_name,omitempty"`
}

// NodeAffinityLabelConfig : 节点亲和性标签配置
type NodeAffinityLabelConfig struct {
	Importance string                    `json:"importance" binding:"omitempty,oneof=low medium high special"`
	CPU        entity.NodeLabelValueType `json:"cpu" binding:"omitempty"`
	Mem        entity.NodeLabelValueType `json:"mem" binding:"omitempty"`
	Exclusive  entity.NodeLabelValueType `json:"exclusive" binding:"omitempty"`
}

// GetTasksReq : 获取任务列表请求
type GetTasksReq struct {
	models.BaseListRequest
	OperatorID         string                      `form:"operator_id" json:"operator_id"`
	AppID              string                      `form:"app_id" json:"app_id"`
	AppIDList          []string                    `form:"app_id_list" json:"app_id_list"`
	EnvName            entity.AppEnvName           `form:"env_name" json:"env_name"`
	ClusterName        entity.ClusterName          `form:"cluster_name" json:"cluster_name" binding:"required"`
	Action             entity.TaskAction           `form:"action" json:"action"`
	ApprovalType       []entity.TaskApprovalType   `form:"approval_type" json:"approval_type"`
	DeployTypeList     []entity.TaskDeployType     `form:"deploy_type_list" json:"deploy_type_list"`
	ApprovalInstanceID string                      `form:"approval_instance_id" json:"approval_instance_id"`
	ApprovalStatusList []entity.TaskApprovalStatus `form:"approval_status_list" json:"approval_status_list"`
	MaxScheduleTime    int64                       `form:"max_schedule_time" json:"max_schedule_time"`
	MinScheduleTime    int64                       `form:"min_schedule_time" json:"min_schedule_time"`
	Version            string                      `form:"version" json:"version"`
	Detail             string                      `form:"detail" json:"detail"`
	// 行为列表
	ActionList []entity.TaskAction `form:"action_list" json:"action_list"`
	// 行为取反列表
	ActionInverseList []entity.TaskAction `form:"action_inverse_list" json:"action_inverse_list"`
	// 状态列表
	StatusList []entity.TaskStatus `form:"status_list" json:"status_list"`
	// 状态列表取反
	StatusInverseList []entity.TaskStatus `form:"status_inverse_list" json:"status_inverse_list"`
	// 最小时间戳
	MinTimestamp int `form:"min_timestamp" json:"min_timestamp"`
	// 最大时间戳
	MaxTimestamp int `form:"max_timestamp" json:"max_timestamp"`
	// 是否暂停
	Suspend null.Bool `form:"suspend" json:"suspend"`
	// 是否不需要查询
	NoNeedQuery bool `json:"no_need_query"`
}

// GetLatestTaskReq : 获取最新任务请求
type GetLatestTaskReq struct {
	AppID        string             `form:"app_id" json:"app_id"`
	EnvName      entity.AppEnvName  `form:"env_name" json:"env_name"`
	ClusterName  entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
	Version      string             `form:"version" json:"version"`
	IgnoreStatus bool               `form:"ignore_status" json:"ignore_status"`
	ActionList   []entity.TaskAction
}

// GetTaskDetailReq : 获取任务详情请求
type GetTaskDetailReq struct {
}

// GetActivitiesReq : 获取活跃任务请求
type GetActivitiesReq struct {
	models.BaseListRequest
	EnvName     entity.AppEnvName `form:"env_name" json:"env_name"`
	ProjectName string            `form:"project_name" json:"project_name"`
	ProjectID   string            `form:"project_id" json:"project_id"`
	AppName     string            `form:"app_name" json:"app_name"`
	AppType     entity.AppType    `form:"app_type" json:"app_type"`
	Action      entity.TaskAction `form:"action" json:"action"`
	OperatorID  string            `form:"operator_id" json:"operator_id"`
	TeamID      string            `form:"team_id" json:"team_id"`
	// 最小时间戳
	MinTimestamp int `form:"min_timestamp" json:"min_timestamp"`
	// 最大时间戳
	MaxTimestamp int  `form:"max_timestamp" json:"max_timestamp"`
	IsFav        bool `form:"is_fav" json:"is_fav"`
}

// UpdateTaskReq : 更新任务请求
type UpdateTaskReq struct {
	Suspend      null.Bool             `form:"suspend" json:"suspend"`
	DeployType   entity.TaskDeployType `form:"deploy_type" json:"deploy_type"`
	ScheduleTime null.Int64            `form:"schedule_time" json:"schedule_time"`

	OperatorID         string                    `json:"-"`
	ApprovalInstanceID string                    `json:"-"`
	ApprovalStatus     entity.TaskApprovalStatus `json:"-"`
	Status             entity.TaskStatus         `json:"-"`
}

// BatchCreateTaskReq : 批量创建任务请求
type BatchCreateTaskReq struct {
	ProjectID             string              `json:"project_id" binding:"required"`
	AppIDs                []string            `json:"app_ids"`
	EnvName               entity.AppEnvName   `json:"env_name" binding:"required"`
	ClusterName           entity.ClusterName  `json:"cluster_name" binding:"required"`
	Action                entity.TaskAction   `json:"action" binding:"required"`
	IgnoreApprovalProcess null.Bool           `json:"ignore_approval_process"`
	Param                 *CreateTaskParamReq `json:"param" binding:"required"`
}

// Approval information.
type ApprovalReq struct {
	Type               entity.TaskApprovalType      `json:"type"`
	QAEngineers        []*entity.DingDingUserDetail `json:"qa_engineers"`
	OperationEngineers []*entity.DingDingUserDetail `json:"operation_engineers"`
	ProductManagers    []*entity.DingDingUserDetail `json:"product_managers"`
}
