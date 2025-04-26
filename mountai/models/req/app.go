package req

import (
	"fmt"

	"gitlab.shanhai.int/sre/library/base/null"

	"rulai/models"
	"rulai/models/entity"
)

// AppDetailDataType 应用详情数据类型
type AppDetailDataType string

// 已支持的应用详情数据类型枚举值
const (
	AppDetailDataTypeGeneral AppDetailDataType = "general"
)

// MaxDescriptionLength 应用描述最大长度(字节数，暂时不必考虑不同编码容量问题)
const MaxDescriptionLength = 1000

// CreateAppReq 创建应用请求参数
type CreateAppReq struct {
	Name                           string                      `json:"name" binding:"required"`
	Type                           entity.AppType              `json:"type" binding:"required,oneof=Service Worker CronJob OneTimeJob"`
	ServiceType                    entity.AppServiceType       `json:"service_type" binding:"omitempty,oneof=Restful GRPC"`
	ServiceExposeType              entity.AppServiceExposeType `json:"service_expose_type" binding:"omitempty,oneof=Ingress LB"`
	ProjectID                      string                      `json:"project_id" binding:"required"`
	SentryProjectSlug              string                      `json:"sentry_project_slug"`
	Description                    string                      `json:"description" binding:"omitempty"`
	LogTTLInDays                   int                         `json:"log_ttl_in_days" binding:"omitempty,min=1,max=365"`
	EnableBranchChangeNotification bool                        `json:"enable_branch_change_notification"`
	EnableIstio                    null.Bool                   `json:"enable_istio" binding:"omitempty"`
}

// GetLogStoreName 获取日志存储名称
func (r *CreateAppReq) GetLogStoreName(projectName, envName string) string {
	return fmt.Sprintf("%s-%s-%s", projectName, r.Name, envName)
}

func (r *CreateAppReq) GetLogTailName(projectName, envName string) string {
	return fmt.Sprintf("%s-%s-%s", projectName, r.Name, envName)
}

// GetServiceName 获取服务名(kubernetes ServiceName)
func (r *CreateAppReq) GetServiceName(projectName string) string {
	if r.Type != entity.AppTypeService {
		return ""
	}
	return fmt.Sprintf("%s-%s", projectName, r.Name)
}

// GetAliAlarmName 获取阿里云报警名称
func (r *CreateAppReq) GetAliAlarmName(projectName, envName string) string {
	return fmt.Sprintf("k8s-%s-%s-%s", projectName, r.Name, envName)
}

// GetAliLogConfigName 获取阿里云日志配置名称
func (r *CreateAppReq) GetAliLogConfigName(projectName string) string {
	return fmt.Sprintf("%s-%s", projectName, r.Name)
}

// GetDefaultEnv 获取默认环境信息
func (r *CreateAppReq) GetDefaultEnv(projectName string, enableBranchChangeNotification bool) map[entity.AppEnvName]entity.AppEnv {
	desiredEnv := entity.AppNormalEnvNames

	env := make(map[entity.AppEnvName]entity.AppEnv)
	for _, envName := range desiredEnv {
		env[envName] = entity.AppEnv{
			AliAlarmName: r.GetAliAlarmName(projectName, string(envName)),
			LogStoreName: r.GetLogStoreName(projectName, string(envName)),
			LogTailName:  r.GetLogTailName(projectName, string(envName)),
			// 默认七层负载均衡
			ServiceProtocol:                entity.LoadBalancerProtocolHTTP,
			EnableBranchChangeNotification: enableBranchChangeNotification,
		}
	}

	return env
}

// UpdateAppReq 更新应用请求参数
type UpdateAppReq struct {
	Name                   string                                `json:"-"`
	Env                    map[entity.AppEnvName]UpdateAppEnvReq `json:"env"`
	SentryProjectSlug      string                                `json:"sentry_project_slug"`
	SentryProjectPublicDsn string                                `json:"sentry_project_public_dsn"`
	Description            null.String                           `json:"description"`
	LoadBalancerInfo       []entity.ServiceLoadBalancer          `json:"load_balancer_info"`
	EnableIstio            null.Bool                             `json:"enable_istio" binding:"omitempty"`
}

// UpdateAppEnvReq 更新应用环境信息请求参数
type UpdateAppEnvReq struct {
	AliAlarmName                   string    `json:"ali_alarm_name"`
	ServiceProtocol                string    `json:"service_protocol"`
	LogStoreName                   string    `json:"log_store_name"`
	LogTailName                    string    `json:"log_tail_name"`
	EnableBranchChangeNotification null.Bool `json:"enable_branch_change_notification"`
	EnableHotReload                null.Bool `json:"enable_hot_reload"`
}

// GetAppsReq 获取应用列表请求参数
type GetAppsReq struct {
	Limit            int                   `form:"limit" json:"limit" binding:"max=50"`
	Page             int                   `form:"page" json:"page" binding:"min=1"`
	IDs              []string              `form:"ids" json:"ids"`
	ProjectID        string                `form:"project_id" json:"project_id"`
	Name             string                `form:"name" json:"name"`
	Type             entity.AppType        `form:"type" json:"type"`
	ServiceType      entity.AppServiceType `form:"service_type" json:"service_type"`
	ServiceName      string                `form:"service_name" json:"service_name"`
	AliLogConfigName string                `form:"ali_log_config_name" json:"ali_log_config_name"`
	Keyword          string                `form:"keyword" json:"keyword"`
	EnvName          entity.AppEnvName     `form:"env_name" json:"env_name"`
	ProjectIDs       []string              `form:"project_ids" json:"project_ids"`
	AppIDs           string                `form:"app_ids" json:"app_ids"`
}

// GetAppDetailReq 获取应用详情请求参数
type GetAppDetailReq struct {
	// 当前公共库解析 config 的时候对应用详情 API 有调用, 并且没有传递 cluster_name, 故只有这里的 cluster_name 参数需要支持空值转换为默认集群
	ClusterName entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"omitempty"`
	EnvName     entity.AppEnvName  `form:"env_name" json:"env_name" binding:"required"`
	DataType    AppDetailDataType  `form:"data_type" json:"data_type"`
}

// CorrectAppNameReq 规范应用名称请求参数(tars迁移服务)
type CorrectAppNameReq struct {
	Name string `json:"name" binding:"required"`
}

// GetAppTipsReq 获取应用资源配置提示请求参数
type GetAppTipsReq struct {
	EnvName     entity.AppEnvName  `form:"env_name" json:"env_name" binding:"required"`
	ClusterName entity.ClusterName `form:"cluster_name" json:"cluster_name" binding:"required"`
}

// CalculateAppRecommendReq 计算应用资源推荐配置请求参数
type CalculateAppRecommendReq struct {
	EnvName           entity.AppEnvName `form:"env_name" json:"env_name" binding:"required"`
	DailyMaxTotalCPU  float64           `form:"daily_max_total_cpu" json:"daily_max_total_cpu" binding:"required"`
	DailyMinTotalCPU  float64           `form:"daily_min_total_cpu" json:"daily_min_total_cpu" binding:"required"`
	WeeklyMaxTotalCPU float64           `form:"weekly_max_total_cpu" json:"weekly_max_total_cpu" binding:"required"`
	DailyMaxTotalMem  float64           `form:"daily_max_total_mem" json:"daily_max_total_mem" binding:"required"`
	DailyMinTotalMem  float64           `form:"daily_min_total_mem" json:"daily_min_total_mem" binding:"required"`
	WeeklyMaxTotalMem float64           `form:"weekly_max_total_mem" json:"weekly_max_total_mem" binding:"required"`
}

type DeleteAppReq struct {
	DeleteSentry bool `form:"delete_sentry" json:"delete_sentry"`
}

// SetAppClusterQDNSWeightsReq 设置应用环境所有集群在 QDNS 统一接入规则中的权重
type SetAppClusterQDNSWeightsReq struct {
	Env                   entity.AppEnvName            `json:"env" binding:"required"`
	ClusterWeights        []*AppClusterKongWeight      `json:"cluster_weights" binding:"required"`
	UpstreamTargetWeights []*KongUpstreamTargetsWeight `json:"traffic_weights" binding:"omitempty"`
	Stage                 entity.TaskAction            `json:"stage"`
	// 是否强制将所有关联的人工配置的权重(Kong上标记有该服务tag)统一成k8s通配域名的权重设置
	// 该值设置为 false 时, 如果存在人工配置的权重, 将会返回特殊错误, 由前端进行提示
	ForceUpdateAll bool `json:"force_update_all" binding:"omitempty"`

	// 功能字段, 服务端自动获取
	OperatorID       string             `json:"-"` // 操作人(从 auth 信息获取)
	DomainController string             `json:"-"` // 域名负责人(当前用第一个项目负责人, 如果没有则为空, TODO: 与运维商议如何规范化)
	HealthCheckPath  string             `json:"-"` // 健康检查路径(从 task param 获取)
	EnableIstio      null.Bool          `json:"-"` // istio 是否启用(从 project 获取)
	ClusterName      entity.ClusterName `json:"-"` // 集群信息, 用于获取指定集群 istio upstream
}

// AppClusterKongWeight 应用环境集群在 Kong 转发规则中的权重
// TODO: binding 没起作用
type AppClusterKongWeight struct {
	ClusterName    entity.ClusterName `json:"cluster_name" binding:"required"`
	TargetHostPort string             `json:"target_host_port"`
	Weight         int                `json:"weight" binding:"min=0,max=100"`
}

// AppKongGrayWeight kong 灰度权重
type KongUpstreamTargetsWeight struct {
	ClusterName    entity.ClusterName `json:"cluster_name"`
	TargetHostPort string             `json:"target_host_port"`
	Weight         int                `json:"weight"`
}

// GetAppClustersWithWorkloadReq 获取应用在指定环境下有工作负载的所有集群请求参数
type GetAppClustersWithWorkloadReq struct {
	models.BaseListRequestWithUnifier
	EnvName entity.AppEnvName `form:"env_name" json:"env_name" binding:"required"`
}
