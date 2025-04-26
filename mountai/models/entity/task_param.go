package entity

import (
	"rulai/config"

	"fmt"

	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

// cpu资源类型
type CPUResourceType string

const (
	CPUResourcePico   = "0.01"
	CPUResourceNano   = "0.1"
	CPUResourceTiny   = "0.25"
	CPUResourceSmall  = "0.5"
	CPUResourceMedium = "1"
	CPUResourceLarge  = "2"
)

// 内存资源类型
type MemResourceType string

const (
	MemResourcePico   = "0.1Gi"
	MemResourceNano   = "0.25Gi"
	MemResourceTiny   = "0.5Gi"
	MemResourceSmall  = "1Gi"
	MemResourceMedium = "2Gi"
	MemResourceLarge  = "4Gi"

	MemResourceNanoBytes = "268435456"
)

// TerminationGracePeriodSpan 宽限终止时长类型
type TerminationGracePeriodSpan int

const (

	// TerminationGracePeriodSpanTiny 默认 最小宽限时长 30秒
	TerminationGracePeriodSpanTiny TerminationGracePeriodSpan = 30
	// TerminationGracePeriodSpanSmall 较小宽限时长 60秒
	TerminationGracePeriodSpanSmall TerminationGracePeriodSpan = 60
	// TerminationGracePeriodSpanMedium 中级宽限时长 90秒
	TerminationGracePeriodSpanMedium TerminationGracePeriodSpan = 90
	// TerminationGracePeriodSpanLarge 较长宽限时长 120秒
	TerminationGracePeriodSpanLarge TerminationGracePeriodSpan = 120

	// TerminationGracePeriodSpanUnknown 未知宽限时长 0秒
	TerminationGracePeriodSpanUnknown TerminationGracePeriodSpan = 0
)

// ProbeDelaySpan 探针延迟时长类型
type ProbeDelaySpan int

const (
	// DefaultProbeDelaySeconds 探针默认延迟时长
	DefaultProbeDelaySeconds ProbeDelaySpan = 10
)

// 以下资源配置按照推荐度排序
var (
	// CPUStgLimitResourceList 测试环境cpu请求资源列表
	CPUStgLimitResourceList = []CPUResourceType{
		CPUResourceLarge,
	}
	// CPUPrdLimitResourceList 线上环境cpu请求资源列表
	CPUPrdLimitResourceList = []CPUResourceType{
		CPUResourceLarge,
	}
	// MemStgLimitResourceList 测试环境内存请求资源列表
	MemStgLimitResourceList = []MemResourceType{
		MemResourceLarge,
	}
	// MemPrdLimitResourceList 线上环境内存请求资源列表
	MemPrdLimitResourceList = []MemResourceType{
		MemResourceLarge,
	}
	// CPUStgRequestResourceList 测试环境cpu请求资源列表
	CPUStgRequestResourceList = []CPUResourceType{
		CPUResourceNano, CPUResourcePico,
	}
	// CPUPrdRequestResourceList 线上环境cpu请求资源列表
	CPUPrdRequestResourceList = []CPUResourceType{
		CPUResourceMedium, CPUResourceSmall, CPUResourceTiny, CPUResourceNano, CPUResourceLarge,
	}
	// MemStgRequestResourceList 测试环境内存请求资源列表
	MemStgRequestResourceList = []MemResourceType{
		MemResourceNano, MemResourcePico,
	}
	// MemPrdRequestResourceList 线上环境内存请求资源列表
	MemPrdRequestResourceList = []MemResourceType{
		MemResourceMedium, MemResourceSmall, MemResourceTiny, MemResourceNano, MemResourceLarge,
	}
)

// TaskParam 任务参数
type TaskParam struct {
	// 镜像版本
	ImageVersion string `bson:"image_version" json:"image_version"`
	// 应用配置的commit id
	ConfigCommitID string `bson:"config_commit_id" json:"config_commit_id"`
	// ConfigRenamePrefix 特殊配置重命名前缀
	ConfigRenamePrefix string `bson:"config_rename_prefix" json:"config_rename_prefix"`
	// ConfigRenameMode 特殊配置重命名模式, binding 注意与 entity 的枚举值同步
	ConfigRenameMode ConfigRenameMode `bson:"config_rename_mode" json:"config_rename_mode"`
	// 环境变量
	Vars map[string]string `bson:"vars" json:"vars"`
	// 应用配置挂载路径
	ConfigMountPath string `bson:"config_mount_path" json:"config_mount_path"`
	// create oss storage to store log no time limits
	OpenColdStorage bool `bson:"open_cold_storage" json:"open_cold_storage"`

	// 最大CPU
	CPULimit CPUResourceType `bson:"cpu_limit" json:"cpu_limit"`
	// 最大内存
	MemLimit MemResourceType `bson:"mem_limit" json:"mem_limit"`
	// 请求CPU
	CPURequest CPUResourceType `bson:"cpu_request" json:"cpu_request"`
	// 请求内存
	MemRequest MemResourceType `bson:"mem_request" json:"mem_request"`
	// 存活探针初始化延迟时长
	LivenessProbeInitialDelaySeconds ProbeDelaySpan `bson:"liveness_probe_initial_delay_seconds" json:"liveness_probe_initial_delay_seconds"`
	// 可读探针初始化延迟时长
	ReadinessProbeInitialDelaySeconds ProbeDelaySpan `bson:"readiness_probe_initial_delay_seconds" json:"readiness_probe_initial_delay_seconds"`

	// 如果应用服务类型是 Restful，且服务暴露方式是 LB，该字段才会有值
	LoadBalancerID string `bson:"load_balancer_id" json:"load_balancer_id"`

	// 预停止指令
	PreStopCommand string `bson:"pre_stop_command" json:"pre_stop_command"`
	// 宽限终止时长
	TerminationGracePeriodSeconds TerminationGracePeriodSpan `bson:"termination_grace_period_sec" json:"termination_grace_period_sec"`
	// 实际覆盖的运行指令，用于覆盖entrypoint
	CoverCommand string `bson:"cover_command" json:"cover_command"`
	// 是否自动扩缩容
	IsAutoScale bool `bson:"is_auto_scale" json:"is_auto_scale"`
	// 最小实例数
	MinPodCount int `bson:"min_pod_count" json:"min_pod_count"`
	// 最大实例数
	MaxPodCount int `bson:"max_pod_count" json:"max_pod_count"`
	// Deprecated: 节点选择器按照官方建议应当弃用，但需要做向后兼容
	NodeSelector map[string]string `bson:"node_selector" json:"node_selector"`
	// 节点亲和性标签
	// 因为所有k8s节点至少都会标记有 importance 标签
	// 所以亲和性标签至少必须设置 importance，且必须为允许的枚举值
	NodeAffinityLabelConfig NodeAffinityLabelConfig `bson:"node_affinity_label_config" json:"node_affinity_label_config"`
	// 关闭pod反亲和性
	DisableHighAvailability bool `bson:"disable_high_availability" json:"disable_high_availability"`
	// 关闭金丝雀发布
	DisableCanary bool `bson:"disable_canary" json:"disable_canary"`
	// 目标暴露的主要端口
	// 用于健康检查
	TargetPort int `bson:"target_port" json:"target_port"`
	// 暴露的额外端口
	ExposedPorts map[string]int `bson:"exposed_ports" json:"exposed_ports"`
	// 健康检查地址
	HealthCheckURL string `bson:"health_check_url" json:"health_check_url"`
	// 是否支持监控
	IsSupportMetrics bool `bson:"is_support_metrics" json:"is_support_metrics"`
	// 监控端口
	MetricsPort int `bson:"metrics_port" json:"metrics_port"`
	// 是否支持会话保持
	IsSupportStickySession bool `bson:"is_support_sticky_session" json:"is_support_sticky_session"`
	// 会话保持cookie过期时间 单位秒
	SessionCookieMaxAge int `bson:"session_cookie_max_age" json:"session_cookie_max_age"`

	// 定时扩缩容任务组列表
	CronScaleJobGroups []*CronScaleJobGroup `bson:"cron_scale_job_groups" json:"cron_scale_job_groups"`
	// 定时扩缩容排除日期，五位时间模板，最小粒度为"天"，更小粒度填充"*"
	CronScaleJobExcludeDates []string `bson:"cron_scale_job_exclude_dates" json:"cron_scale_job_exclude_dates"`

	// 执行命令
	CronCommand string `bson:"cron_command" json:"cron_command"`
	// 定时参数
	CronParam string `bson:"cron_param" json:"cron_param"`
	// 并发策略
	ConcurrencyPolicy batchV1.ConcurrencyPolicy `bson:"concurrency_policy" json:"concurrency_policy"`
	// 重启策略
	RestartPolicy v1.RestartPolicy `bson:"restart_policy" json:"restart_policy"`
	// 成功的历史最大记录
	SuccessfulHistoryLimit int `bson:"successful_history_limit" json:"successful_history_limit"`
	// 失败的历史最大记录
	FailedHistoryLimit int `bson:"failed_history_limit" json:"failed_history_limit"`
	// 活跃超时时间
	ActiveDeadlineSeconds int `bson:"active_deadline_seconds" json:"active_deadline_seconds"`
	// 手动执行Job名称
	ManualJobName string `bson:"manual_job_name" json:"manual_job_name"`

	// 脚本执行命令
	JobCommand string `bson:"job_command" json:"job_command"`
	// 任务失败重试次数
	BackoffLimit int `bson:"backoff_limit" json:"backoff_limit"`

	// 用于清理工作的字段
	// 清理的项目名
	CleanedProjectName string `bson:"cleaned_project_name,omitempty" json:"cleaned_project_name,omitempty"`
	// 清理的应用名
	CleanedAppName string `bson:"cleaned_app_name,omitempty" json:"cleaned_app_name,omitempty"`
	// 清理的应用类型
	CleanedAppType AppType `bson:"cleaned_app_type,omitempty" json:"cleaned_app_type,omitempty"`
	// 清理的应用服务类型
	CleanedAppServiceType AppServiceType `bson:"cleaned_app_service_type,omitempty" json:"cleaned_app_service_type,omitempty"`
	// 清理的应用服务暴露方式
	CleanedAppServiceExposeType AppServiceExposeType `bson:"cleaned_app_service_expose_type,omitempty"  json:"cleaned_app_service_expose_type,omitempty"`
	// 清理的服务名
	CleanedServiceName string `bson:"cleaned_service_name,omitempty" json:"cleaned_service_name,omitempty"`
	// 清理的阿里云告警名
	CleanedAliAlarmName string `bson:"cleaned_ali_alarm_name,omitempty" json:"cleaned_ali_alarm_name,omitempty"`
	// 清理的阿里云日志采集器名
	CleanedAliLogConfigName string `bson:"cleaned_ali_log_config_name,omitempty" json:"cleaned_ali_log_config_name,omitempty"`
	// 清理的阿里云日志仓库名
	CleanedAliLogStoreName string `bson:"cleaned_ali_log_store_name,omitempty" json:"cleaned_ali_log_store_name,omitempty"`
}

// CronScaleJobGroup 扩缩容任务组
type CronScaleJobGroup struct {
	Name string `bson:"name" json:"name"`
	// TargetSize 扩容pod数量下限
	TargetSize int `bson:"target_size" json:"target_size"`
	// UpSchedule 和 DownSchedule 扩/缩容时间模板，五位模板。
	UpSchedule   string `bson:"up_schedule" json:"up_schedule"`
	DownSchedule string `bson:"down_schedule" json:"down_schedule"`
	// RunOnce 一次性扩缩容任务对
	RunOnce bool `bson:"run_once" json:"run_once"`
}

func (p *TaskParam) ConfigURL(args map[string]interface{}) string {
	if p.ConfigCommitID == "" {
		return ""
	}
	return GetAppConfigURL(p.ConfigCommitID)
}

// GetAppConfigURL 获取应用配置的跳转url
func GetAppConfigURL(commitID string) string {
	return fmt.Sprintf("%s/infra/config-center/tree/%s", config.Conf.Git.Host, commitID)
}
