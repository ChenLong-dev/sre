package resp

import (
	"strings"

	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"

	"rulai/models/entity"
)

// TaskDetailResp 任务详情返回值
type TaskDetailResp struct {
	ID            string                  `json:"id" deepcopy:"method:GenerateObjectIDString"`
	Version       string                  `json:"version"`
	Action        entity.TaskAction       `json:"action"`
	ActionDisplay string                  `json:"action_display" deepcopy:"method:ActionDisplay"`
	ApprovalType  entity.TaskApprovalType `json:"approval_type"`
	DeployType    entity.TaskDeployType   `json:"deploy_type"`
	ScheduleTime  string                  `json:"schedule_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	Approval      *ApprovalResp           `json:"approval"`
	Detail        string                  `json:"detail"`
	Description   string                  `json:"description"`
	RetryCount    int                     `json:"retry_count"`
	ClusterName   entity.ClusterName      `json:"cluster_name"`
	Status        entity.TaskStatus       `json:"status"`
	StatusDisplay string                  `json:"status_display" deepcopy:"method:StatusDisplay"`
	DisplayIcon   string                  `json:"display_icon" deepcopy:"method:DisplayIcon"`
	EnvName       entity.AppEnvName       `json:"env_name"`
	AppID         string                  `json:"app_id"`
	OperatorID    string                  `json:"operator_id"`
	Suspend       bool                    `json:"suspend"`
	Param         *TaskParamDetailResp    `json:"param"`
	Namespace     string                  `json:"namespace"`
	CreateTime    string                  `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime    string                  `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
}

func (t *TaskDetailResp) GetNamespace(enableIstio bool) string {
	if enableIstio {
		return entity.IstioNamespacePrefix + string(t.EnvName)
	}
	return string(t.EnvName)
}

// GetEnvName 这里使用 namespace 代替 envName 以提供集群私有域名生成使用, 如果 namespace 为空则默认使用 task 的 envName 字段
func (t *TaskDetailResp) GetEnvName() entity.AppEnvName {
	if t.Namespace == "" {
		return t.EnvName
	}
	return entity.AppEnvName(t.Namespace)
}

func (t *TaskDetailResp) GetNamespaceAppEnv(enableIstio bool) entity.AppEnvName {
	return entity.AppEnvName(t.GetNamespace(enableIstio))
}

// "image_version" : "crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/shanhaii/user-xingya:393d988d-feat_sockpuppet-test"
func (t *TaskDetailResp) GetDeployBranch() string {
	segments := strings.Split(t.Param.ImageVersion, ":")
	if len(segments) == 2 {
		hashAndBranch := segments[1]
		if idx := strings.Index(hashAndBranch, "-"); idx != -1 {
			return hashAndBranch[idx+1:]
		}
	}
	return ""
}

// TaskParamDetailResp 任务参数详情返回值
type TaskParamDetailResp struct {
	ImageVersion       string                  `json:"image_version"`
	ConfigCommitID     string                  `json:"config_commit_id"`
	ConfigRenamePrefix string                  `json:"config_rename_prefix"`
	ConfigRenameMode   entity.ConfigRenameMode `json:"config_rename_mode"`
	ConfigURL          string                  `json:"config_url" deepcopy:"method:ConfigURL"`
	ConfigMountPath    string                  `json:"config_mount_path"`
	OpenColdStorage    bool                    `json:"open_cold_storage"`

	HealthCheckURL                string                            `json:"health_check_url"`
	IsAutoScale                   bool                              `json:"is_auto_scale"`
	IsSupportMetrics              bool                              `json:"is_support_metrics"`
	MetricsPort                   int                               `json:"metrics_port"`
	Vars                          map[string]string                 `json:"vars"`
	PreStopCommand                string                            `json:"pre_stop_command"`
	TerminationGracePeriodSeconds entity.TerminationGracePeriodSpan `json:"termination_grace_period_sec"`
	CoverCommand                  string                            `json:"cover_command"`
	TargetPort                    int                               `json:"target_port"`
	ExposedPorts                  map[string]int                    `json:"exposed_ports"`
	// Deprecated: 节点选择器向后兼容保留字段
	NodeSelector map[string]string `json:"node_selector"`
	// 节点亲和性标签设置
	NodeAffinityLabelConfig entity.NodeAffinityLabelConfig `json:"node_affinity_label_config"`
	// 是否支持会话保持
	IsSupportStickySession bool `json:"is_support_sticky_session"`
	// 会话保持cookie过期时间 单位秒
	SessionCookieMaxAge int `json:"session_cookie_max_age"`
	// 关闭pod反亲和性
	DisableHighAvailability bool `json:"disable_high_availability"`
	// 关闭金丝雀发布
	DisableCanary bool `json:"disable_canary"`

	CPURequest entity.CPUResourceType `json:"cpu_request"`
	MemRequest entity.MemResourceType `json:"mem_request"`
	CPULimit   entity.CPUResourceType `json:"cpu_limit"`
	MemLimit   entity.MemResourceType `json:"mem_limit"`

	// 存活探针初始化延迟时长
	LivenessProbeInitialDelaySeconds entity.ProbeDelaySpan `json:"liveness_probe_initial_delay_seconds"`
	// 可读探针初始化延迟时长
	ReadinessProbeInitialDelaySeconds entity.ProbeDelaySpan `json:"readiness_probe_initial_delay_seconds"`

	// 如果应用服务类型是 Restful，且服务暴露方式是 LB，该字段才会有值
	LoadBalancerID string `json:"load_balancer_id"`

	MinPodCount int `json:"min_pod_count"`
	MaxPodCount int `json:"max_pod_count"`

	// 定时扩缩容任务组列表
	CronScaleJobGroups []*entity.CronScaleJobGroup `json:"cron_scale_job_groups"`
	// 定时扩缩容排除日期，五位时间模板，最小粒度为"天"，更小粒度填充"*"
	CronScaleJobExcludeDates []string `json:"cron_scale_job_exclude_dates"`

	CronCommand            string                    `json:"cron_command"`
	CronParam              string                    `json:"cron_param"`
	ConcurrencyPolicy      batchV1.ConcurrencyPolicy `json:"concurrency_policy"`
	RestartPolicy          v1.RestartPolicy          `json:"restart_policy"`
	SuccessfulHistoryLimit int                       `json:"successful_history_limit"`
	FailedHistoryLimit     int                       `json:"failed_history_limit"`
	ActiveDeadlineSeconds  int                       `json:"active_deadline_seconds"`
	BackoffLimit           int                       `json:"backoff_limit"`
	JobCommand             string                    `json:"job_command"`
	ManualJobName          string                    `json:"manual_job_name"`

	// 用于清理工作的字段
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

// ActiveTaskResp 活跃任务返回值
type ActiveTaskResp struct {
	ID            string             `json:"id"`
	Action        entity.TaskAction  `json:"action"`
	Status        entity.TaskStatus  `json:"status"`
	ActionDisplay string             `json:"action_display"`
	StatusDisplay string             `json:"status_display"`
	EnvName       entity.AppEnvName  `json:"env_name"`
	ClusterName   entity.ClusterName `json:"cluster_name"`
	CreateTime    string             `json:"create_time"`
	Version       string             `json:"version"`
	RetryCount    int                `json:"retry_count"`
	DisplayIcon   string             `json:"display_icon"`
	Detail        string             `json:"detail"`
	Description   string             `json:"description"`

	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	ProjectDesc string `json:"project_desc"`

	TeamID   string `json:"team_id"`
	TeamName string `json:"team_name"`

	AppID          string `json:"app_id"`
	AppName        string `json:"app_name"`
	AppType        string `json:"app_type"`
	AppServiceType string `json:"app_service_type"`

	OperatorID        string `json:"operator_id"`
	OperatorName      string `json:"operator_name"`
	OperatorAvatarURL string `json:"operator_avatar_url"`

	ImageVersion string `json:"image_version"`
}

// ApprovalResp information.
type ApprovalResp struct {
	Type               entity.TaskApprovalType      `json:"type"`
	Status             entity.TaskApprovalStatus    `json:"status"`
	QAEngineers        []*entity.DingDingUserDetail `json:"qa_engineers"`
	ProductManagers    []*entity.DingDingUserDetail `json:"product_managers"`
	OperationEngineers []*entity.DingDingUserDetail `json:"operation_engineers"`
	InstanceID         string                       `json:"instance_id"`
}
