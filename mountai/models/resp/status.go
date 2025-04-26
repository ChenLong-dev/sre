package resp

import (
	"rulai/models/entity"

	"github.com/AliyunContainerService/kubernetes-cronhpa-controller/pkg/apis/autoscaling/v1beta1"
	v1 "k8s.io/api/core/v1"
)

type RunningStatusDetailResp struct {
	Version           string                  `json:"version"`
	CreateTime        string                  `json:"create_time"`
	TaskID            string                  `json:"task_id"`
	TaskStatus        string                  `json:"task_status"`
	TaskStatusDisplay string                  `json:"task_status_display"`
	TaskDisplayIcon   string                  `json:"task_display_icon"`
	TaskDetail        string                  `json:"task_detail"`
	TaskRetryCount    int                     `json:"task_retry_count"`
	TaskSuspend       bool                    `json:"task_suspend"`
	ImageVersion      string                  `json:"image_version"`
	ConfigURL         string                  `json:"config_url"`
	ApprovalType      entity.TaskApprovalType `json:"approval_type"`
	DeployType        entity.TaskDeployType   `json:"deploy_type"`
	ScheduleTime      string                  `json:"schedule_time"`
	Approval          *ApprovalResp           `json:"approval"`

	ReadyPodCount int    `json:"ready_pod_count"`
	TotalPodCount int    `json:"total_pod_count"`
	PodMonitorURL string `json:"pod_monitor_url"`

	CronParam        string `json:"cron_param"`
	LastScheduleTime string `json:"last_schedule_time"`
	IsSuspend        bool   `json:"is_suspend"`
	NextScheduleTime string `json:"next_schedule_time"`

	DeploymentPods []*RunningStatusPodDetailResp `json:"deployment_pods"`

	Jobs []*RunningStatusJobDetailResp `json:"jobs"`

	AllowedActions map[entity.TaskAction]bool `json:"allowed_actions"`

	CronAutoScaleJobs []v1beta1.Condition `json:"cron_auto_scale_jobs,omitempty"`
}

type RunningStatusPodDetailResp struct {
	Name         string      `json:"name"`
	RestartCount int         `json:"restart_count"`
	Phase        v1.PodPhase `json:"phase"`
	Age          string      `json:"age"`
	NodeIP       string      `json:"node_ip"`
	PodIP        string      `json:"pod_ip"`
	CreateTime   string      `json:"create_time"`
	ShellURL     string      `json:"shell_url"`
	Namespace    string      `json:"namespace"`
}

type RunningStatusJobDetailResp struct {
	Name              string                       `json:"name"`
	StartTime         string                       `json:"start_time"`
	CompletionTime    string                       `json:"completion_time"`
	SucceededCount    int                          `json:"succeeded_count"`
	FailedCount       int                          `json:"failed_count"`
	NeedCompleteCount int                          `json:"need_complete_count"`
	LaunchType        entity.LaunchType            `json:"launch_type"`
	Pods              []RunningStatusPodDetailResp `json:"pods"`
}

type RunningStatusListResp struct {
	Version           string `json:"version"`
	Namespace         string `json:"namespace"`
	CreateTime        string `json:"create_time"`
	TaskID            string `json:"task_id"`
	TaskStatus        string `json:"task_status"`
	TaskStatusDisplay string `json:"task_status_display"`
	TaskDisplayIcon   string `json:"task_display_icon"`
	TaskRetryCount    int    `json:"task_retry_count"`
	TaskSuspend       bool   `json:"task_suspend"`
	ImageVersion      string `json:"image_version"`
	ConfigURL         string `json:"config_url"`
	PodMonitorURL     string `json:"pod_monitor_url"`

	ReadyPodCount int `json:"ready_pod_count"`
	TotalPodCount int `json:"total_pod_count"`

	CronParam        string `json:"cron_param"`
	LastScheduleTime string `json:"last_schedule_time"`
	IsSuspend        bool   `json:"is_suspend"`
	NextScheduleTime string `json:"next_schedule_time"`
	NetworkTraffic   bool   `json:"network_traffic"`
}

type GetRunningStatusDescriptionResp struct {
	ServiceDesc        *DescribeServiceResp        `json:"service,omitempty"`
	CronJobDesc        *DescribeCronJobResp        `json:"cronjob,omitempty"`
	JobDesc            *DescribeJobResp            `json:"job,omitempty"`
	DeploymentDesc     *DescribeDeploymentResp     `json:"deployment,omitempty"`
	HPADesc            *DescribeHPAResp            `json:"hpa,omitempty"`
	IngressDesc        *DescribeIngressResp        `json:"ingress,omitempty"`
	VirtualServiceDesc *DescribeVirtualServiceResp `json:"virtualservice,omitempty"`
}

// GetCronHPAStatusResp 获取cronHPA状态响应体
type GetCronHPAStatusResp struct {
	Version string              `json:"version,omitempty"`
	EnvName string              `json:"env_name,omitempty"`
	Jobs    []v1beta1.Condition `json:"jobs,omitempty"`
}
