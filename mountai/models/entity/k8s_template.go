package entity

import (
	"strings"
	"time"

	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

const DefaultTemplateFileDir = "./template/k8s/"

// k8s 资源类型
const (
	K8sObjectKindAliyunLogConfig = "AliyunLogConfig"
	K8sObjectKindConfigMap       = "ConfigMap"
	K8sObjectKindCronHPA         = "CronHorizontalPodAutoscaler"
	K8sObjectKindCronJob         = "CronJob"
	K8sObjectKindDeployment      = "Deployment"
	K8sObjectKindHPA             = "HorizontalPodAutoscaler"
	K8sObjectKindIngress         = "Ingress"
	K8sObjectKindJob             = "Job"
	K8sObjectKindReplicaSet      = "ReplicaSet"
	K8sObjectKindService         = "Service"
	K8sObjectKindVirtualService  = "VirtualService"
)

// AMSSupportedK8sObjectKindsByAllVendors AMS 所有集群通用支持的 k8s 资源类型
var AMSSupportedK8sObjectKindsByAllVendors = map[string]struct{}{
	K8sObjectKindConfigMap:      {},
	K8sObjectKindCronHPA:        {}, // CronHPA 虽然是阿里云开源, 但目前所有集群均安装使用
	K8sObjectKindCronJob:        {},
	K8sObjectKindDeployment:     {},
	K8sObjectKindHPA:            {},
	K8sObjectKindIngress:        {},
	K8sObjectKindJob:            {},
	K8sObjectKindReplicaSet:     {},
	K8sObjectKindService:        {},
	K8sObjectKindVirtualService: {},
}

// AMSSupportedK8sObjectKindsByVendors AMS 各运营商集群分别支持的 k8s 资源类型
var AMSSupportedK8sObjectKindsByVendors = map[VendorName]map[string]struct{}{
	VendorAli: {
		K8sObjectKindAliyunLogConfig: {},
	},
	VendorHuawei: {},
}

const (
	// DeploymentMetricsLabelEnable 开启监控标签值
	DeploymentMetricsLabelEnable = "enable"
	// DeploymentMetricsLabelDisable 关闭监控标签值
	DeploymentMetricsLabelDisable = "disable"

	// DefaultStartingDeadlineSeconds 默认启动超时时间
	DefaultStartingDeadlineSeconds = 100

	// ServiceTCPProtocol 服务tcp协议
	ServiceTCPProtocol = "TCP"
	// ServiceDefaultInternalPort 服务默认内部端口
	ServiceDefaultInternalPort int32 = 80
	// ServiceDefaultInternalPort443 服务默认内部443端口
	ServiceDefaultInternalPort443 int32 = 443
	// ServiceGRPCDefaultInternalPort grpc服务默认内部端口
	ServiceGRPCDefaultInternalPort int32 = 443
	// ServiceDefaultHTTPName 服务默认名
	ServiceDefaultHTTPName = "http"
	// ServiceDefaultHTTPSName https服务默认名
	ServiceDefaultHTTPSName = "https"

	// hpa初始化延迟
	// 为防止prometheus中没有数据
	HPAInitDelayDuration = 30 * time.Second

	// ingress注释名
	// 亲和性
	IngressAnnotationNameAffinity = "nginx.ingress.kubernetes.io/affinity"
	// IngressAnnotationNameAffinityMode 亲和性模式
	IngressAnnotationNameAffinityMode = "nginx.ingress.kubernetes.io/affinity-mode"
	// IngressAnnotationNameSessionCookieMaxAge 会话cookie过期时间 单位秒
	IngressAnnotationNameSessionCookieMaxAge = "nginx.ingress.kubernetes.io/session-cookie-max-age"
	// IngressAnnotationNameBackendProtocol 后端协议
	IngressAnnotationNameBackendProtocol = "nginx.ingress.kubernetes.io/backend-protocol"

	// IngressAnnotationIngressClasss ingress class
	IngressAnnotationIngressClasss = "kubernetes.io/ingress.class"

	// ingress注释值
	// cookie模式
	IngressAnnotationValueAffinityCookie = "cookie"
	// 平衡模式
	IngressAnnotationValueAffinityModeBalanced = "balanced"

	// IngressAnnotationValueGRPCS 后端协议名
	IngressAnnotationValueGRPCS = "GRPCS"

	// ServiceInClusterDNSLable 是否禁用集群内DNS
	ServiceInClusterDNSLable = "shanhai.int/incluster_dns"
	// ServiceInClusterValueDisabled 禁用集群内解析
	ServiceInClusterValueDisabled = "disabled"
	// ShadowServicePrefix the prefix of "shadow" service
	// 目前应用名称长度< 50 所以len("shadow--" +appService) < 63
	ShadowServicePrefix = "shadow--"
)

// K8sObjectTemplate k8s 资源模板接口类型
type K8sObjectTemplate interface {
	Kind() string
	SetAPIVersion(ver string)
}

// K8sWorkloadObjectTemplate k8s 工作负载类型资源模板接口类型
// TODO: 过渡方案, 在无法区分多云镜像仓库的阶段进行兼容性处理
type K8sWorkloadObjectTemplate interface {
	K8sObjectTemplate
	UnifyImageName(imageRegistryHostWithNamespace string)
}

// ImageName 镜像名称, 实现 K8sWorkloadObjectTemplate 接口类型的 UnifyImageName 方法
type ImageName string

func (in *ImageName) UnifyImageName(imageRegistryHostWithNamespace string) {
	*in = ImageName(strings.Replace(string(*in), "crpi-592g7buyguepbrqd-vpc.cn-shanghai.personal.cr.aliyuncs.com/infra", imageRegistryHostWithNamespace, 1))
}

// Deployment渲染模版
type DeploymentTemplate struct {
	// k8s API 版本
	APIVersion string
	// 版本
	DeploymentVersion string
	// 命名空间
	Namespace string
	// 项目名
	ProjectName string
	// 应用名
	AppName string
	// 应用服务类型
	AppServiceType AppServiceType
	// 启用 grpc-health-probe 方式的健康检查端口
	GRPCHealthProbePort string
	// 启用 grpc-health-probe 方式的健康检查是否需要 TLS 加密
	GRPCHealthProbeUseTLS bool
	// 标签
	Labels map[string]string

	// 副本数
	Replicas int32
	// 镜像名
	ImageName
	// 容器名
	ContainerName string
	// 环境变量
	Env []*EnvTemplate
	// 日志仓库名
	LogStoreName string
	// 配置文件名
	ConfigName string
	// 配置文件路径
	ConfigMountPath string
	// pod注释
	PodAnnotations map[string]string

	// 实际覆盖的运行指令，用于覆盖entrypoint
	CoverCommand string
	// 预停止指令
	PreStopCommand string
	// 宽限终止时长
	TerminationGracePeriodSeconds TerminationGracePeriodSpan
	// 是否开启健康检查
	EnableHealth bool
	// 健康检查地址
	HealthCheckURL string
	// 目标暴露端口
	TargetPort int32
	// metrics端口
	MetricsPort int32
	// 污点容忍
	Tolerations map[string]string

	// 最大CPU
	CPULimit CPUResourceType
	// 最大内存
	MemoryLimit MemResourceType
	// 请求CPU
	CPURequest CPUResourceType
	// 请求内存
	MemoryRequest MemResourceType
	// 存活探针初始化延迟时长
	LivenessProbeInitialDelaySeconds ProbeDelaySpan
	// 可读探针初始化延迟时长
	ReadinessProbeInitialDelaySeconds ProbeDelaySpan

	// 节点亲和性模板(功能性字段，从app配置中同步)
	// 取消节点选择模版(nodeSelector)
	NodeAffinity NodeAffinityTemplate
	// 关闭pod反亲和性
	DisableHighAvailability bool

	LocalDNS                string
	ProgressDeadlineSeconds int
}

func (tpl *DeploymentTemplate) Kind() string { return K8sObjectKindDeployment }

func (tpl *DeploymentTemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }

// 环境变量模版
type EnvTemplate struct {
	// 变量名
	Name string
	// 变量值
	Value string
}

// Service渲染模版
type ServiceTemplate struct {
	// k8s API 版本
	APIVersion string
	// 服务名
	Name string
	// K8s服务类型
	Type v1.ServiceType
	// 命名空间
	Namespace string
	// 项目名
	ProjectName string
	// 应用名
	AppName string
	// 应用服务类型
	AppServiceType AppServiceType

	// 映射端口
	Ports []*ServicePortTemplate
	// 负载均衡协议
	Protocol string

	// 是否使用 lb
	WithLB bool
	// 注解(一般用于适配各服务商的配置)
	Annotations map[string]string
}

func (tpl *ServiceTemplate) Kind() string { return K8sObjectKindService }

func (tpl *ServiceTemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }

// 服务端口模版
type ServicePortTemplate struct {
	// 名称
	Name string
	// 协议
	Protocol string
	// 对外端口号
	ExternalPort int32
	// 目标端口号
	TargetPort int32
}

// CronJob渲染模版
type CronJobTemplate struct {
	// k8s API 版本
	APIVersion string
	// 版本
	CronJobVersion string
	// 命名空间
	Namespace string
	// 项目名
	ProjectName string
	// 应用名
	AppName string
	// 标签
	Labels map[string]string

	// 镜像名
	ImageName
	// 容器名
	ContainerName string
	// 环境变量
	Env []*EnvTemplate
	// 日志仓库名
	LogStoreName string
	// 配置文件名
	ConfigName string
	// 配置文件路径
	ConfigMountPath string
	// pod注释
	PodAnnotations map[string]string

	// 是否停止
	Suspend bool
	// 定时参数
	Schedule string
	// 执行命令
	CronCommand string
	// 预停止指令
	PreStopCommand string
	// 重启策略
	RestartPolicy v1.RestartPolicy
	// 并发策略
	ConcurrencyPolicy batchV1.ConcurrencyPolicy
	// 启动超时时间
	StartingDeadlineSeconds int64
	// 宽限终止时长
	TerminationGracePeriodSeconds TerminationGracePeriodSpan
	// 成功的历史最大记录
	SuccessfulJobsHistoryLimit int32
	// 失败的历史最大记录
	FailedJobsHistoryLimit int32
	// 活跃超时时间
	ActiveDeadlineSeconds int
	// 污点容忍
	Tolerations map[string]string

	// 最大CPU
	CPULimit CPUResourceType
	// 最大内存
	MemoryLimit MemResourceType
	// 请求CPU
	CPURequest CPUResourceType
	// 请求内存
	MemoryRequest MemResourceType

	// 节点亲和性模板(功能性字段，从app配置中同步)
	// 取消节点选择模版(nodeSelector)
	NodeAffinity NodeAffinityTemplate

	// 重试次数
	BackoffLimit int
}

func (tpl *CronJobTemplate) Kind() string { return K8sObjectKindCronJob }

func (tpl *CronJobTemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }

// 应用配置ConfigMap渲染模版
type AppConfigMapTemplate struct {
	// k8s API 版本
	APIVersion string
	// 配置名称
	Name string
	// 命名空间
	Namespace string
	// 项目名
	ProjectName string
	// 应用名
	AppName string
	// 团队标签
	TeamLabel string
	// 标签
	Labels map[string]string

	// 配置数据
	Data interface{}
}

func (tpl *AppConfigMapTemplate) Kind() string { return K8sObjectKindConfigMap }

func (tpl *AppConfigMapTemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }

// 阿里云日志采集CRD渲染模版
type AliyunLogConfigTemplate struct {
	// k8s API 版本
	APIVersion string
	// 采集器名称
	Name string
	// 命名空间
	Namespace string
	// 集群名
	ClusterName ClusterName
	// 项目名
	ProjectName string
	// 应用名
	AppName string
	// 团队标签
	TeamLabel string

	// 日志仓库名
	LogStoreName string
	// logtail名
	LogTailName string
	// 容器名
	ContainerName string
	// 过滤环境标签
	ExcludeLabel string
	// create oss storage to store log no time limits
	OpenColdStorage string
}

func (tpl *AliyunLogConfigTemplate) Kind() string { return K8sObjectKindAliyunLogConfig }

func (tpl *AliyunLogConfigTemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }

// HPATemplate HPA渲染模版
type HPATemplate struct {
	// k8s API 版本
	APIVersion string
	// 采集器名称
	Name string
	// 命名空间
	Namespace string
	// 项目名
	ProjectName string
	// 应用名
	AppName string

	// HPA 控制的目标
	ScaleTargetRef ScaleTargetRefTemplate

	// 最小实例数
	MinReplicas int32
	// 最大实例数
	MaxReplicas int32
	// 扩容的Cpu限制
	CPUTarget int32
	// 扩容的内存限制
	MemTarget int32
}

func (tpl *HPATemplate) Kind() string { return K8sObjectKindHPA }

func (tpl *HPATemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }

// ScaleTargetRefTemplate 控制目标渲染模板
type ScaleTargetRefTemplate struct {
	Kind       string
	Name       string
	APIVersion string
}

// CronHPATemplate CronHPA渲染模板
type CronHPATemplate struct {
	// k8s API 版本
	APIVersion  string
	Name        string
	Namespace   string
	ProjectName string
	AppName     string
	// CronHPA 控制的目标
	ScaleTargetRef ScaleTargetRefTemplate
	// 标签
	Labels                   map[string]string
	CronScaleJobs            []*CronScaleJob
	CronScaleJobExcludeDates []string
}

func (tpl *CronHPATemplate) Kind() string { return K8sObjectKindCronHPA }

func (tpl *CronHPATemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }

// CronScaleJobPair 扩缩容任务对
type CronScaleJobPair struct {
	Name      string
	ScaleUp   *CronScaleJob
	ScaleDown *CronScaleJob
}

// CronScaleJob 扩/缩容任务
type CronScaleJob struct {
	Name string
	// 扩/缩容时间模板，六位模板, 最小粒度为"分"，"秒"字段填充"0"
	Schedule   string
	TargetSize int
	RunOnce    bool
}

// JobTemplate Job渲染模版
type JobTemplate struct {
	// k8s API 版本
	APIVersion string
	// 版本
	JobVersion string
	// 命名空间
	Namespace string
	// 项目名
	ProjectName string
	// 应用名
	AppName string
	// 标签
	Labels map[string]string

	// 镜像名
	ImageName
	// 容器名
	ContainerName string
	// 环境变量
	Env []*EnvTemplate
	// 日志仓库名
	LogStoreName string
	// 配置文件名
	ConfigName string
	// 配置文件路径
	ConfigMountPath string
	// pod注释
	PodAnnotations map[string]string

	// 执行命令
	JobCommand string
	// 活跃超时时间
	ActiveDeadlineSeconds int
	// 宽限终止时长
	TerminationGracePeriodSeconds TerminationGracePeriodSpan
	// 重试次数
	BackoffLimit int
	// 污点容忍
	Tolerations map[string]string

	// 最大CPU
	CPULimit CPUResourceType
	// 最大内存
	MemoryLimit MemResourceType
	// 请求CPU
	CPURequest CPUResourceType
	// 请求内存
	MemoryRequest MemResourceType

	// 节点亲和性模板(功能性字段，从app配置中同步)
	// 取消节点选择模版(nodeSelector)
	NodeAffinity NodeAffinityTemplate
}

func (tpl *JobTemplate) Kind() string { return K8sObjectKindJob }

func (tpl *JobTemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }

// IngressTemplate 渲染模版
type IngressTemplate struct {
	// k8s API 版本
	APIVersion string
	// ingress名
	Name string
	// 命名空间
	Namespace string
	// 项目名
	ProjectName string
	// 应用名
	AppName string
	// 服务名
	ServiceName string
	// 服务host
	ServiceHost string
	// 集群专用的服务host
	ServiceHostWithCluster string
	// 所有集群的域名, ingress 兼容kong重试产生的 404
	ServiceHostsWithCluster []string
	// 注释
	Annotations map[string]string
	// 存放tls证书/密钥的k8s的secret资源名
	SecretName string
	// 服务端口
	ServicePort int32
	// Ingress Class Name
	IngressClass string
	// 服务对外host
	ServiceHostPublic string
	// 证书secret
	TlsSecretName string
}

func (tpl *IngressTemplate) Kind() string { return K8sObjectKindIngress }

func (tpl *IngressTemplate) SetAPIVersion(ver string) { tpl.APIVersion = ver }
