package resp

import (
	"fmt"

	"rulai/config"
	"rulai/models/entity"
)

// AppDetailResp 应用详情返回值
type AppDetailResp struct {
	ID                     string                                 `json:"id" deepcopy:"method:GenerateObjectIDString"`
	Name                   string                                 `json:"name"`
	Type                   entity.AppType                         `json:"type"`
	ServiceType            entity.AppServiceType                  `json:"service_type"`
	ServiceExposeType      entity.AppServiceExposeType            `json:"service_expose_type"`
	LoadBalancerInfo       []entity.ServiceLoadBalancer           `json:"load_balancer_info" deepcopy:"method:DecodeLoadBalancerInfo"`
	ServiceName            string                                 `json:"service_name"`
	AliLogConfigName       string                                 `json:"ali_log_config_name"`
	ProjectID              string                                 `json:"project_id"`
	Env                    map[entity.AppEnvName]AppEnvDetailResp `json:"env"`
	SentryProjectPublicDsn string                                 `json:"sentry_project_public_dsn"`
	SentryProjectSlug      string                                 `json:"sentry_project_slug"`
	Description            string                                 `json:"description"`
	CreateTime             string                                 `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime             string                                 `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	LogTTLInDays           int                                    `json:"log_ttl_in_days"`
	EnableIstio            bool                                   `json:"enable_istio"`
}

// FormatSentryProjectURL 格式化输出 sentry 项目链接地址
func (item *AppDetailResp) FormatSentryProjectURL(args map[string]interface{}) string {
	if item.SentryProjectSlug == "" {
		return ""
	}

	return fmt.Sprintf(
		"%s/%s/%s/",
		config.Conf.SentrySystem.Host,
		config.Conf.SentrySystem.Organization,
		item.SentryProjectSlug,
	)
}

// AppEnvDetailResp 应用环境信息详情返回值
type AppEnvDetailResp struct {
	AliAlarmName string `json:"ali_alarm_name"`
	//Deprecated: 由于更换日志仓库名会导致日志丢失，所以需在生成应用时初始化完毕（即将废弃：使用entity.project中定义的LogStoreName替代以减少logstore数量）
	LogStoreName                   string `json:"log_store_name"`
	LogTailName                    string `json:"log_tail_name"`
	ServiceProtocol                string `json:"service_protocol"`
	EnableBranchChangeNotification bool   `json:"enable_branch_change_notification"`
	EnableHotReload                bool   `json:"enable_hot_reload"`
}

// AppListResp 应用列表返回值
type AppListResp struct {
	ID                string                              `json:"id" deepcopy:"method:GenerateObjectIDString"`
	Name              string                              `json:"name"`
	Type              entity.AppType                      `json:"type"`
	ServiceType       entity.AppServiceType               `json:"service_type"`
	ServiceName       string                              `json:"service_name"`
	ServiceExposeType entity.AppServiceExposeType         `json:"service_expose_type"`
	LoadBalancerInfo  []entity.ServiceLoadBalancer        `json:"load_balancer_info" deepcopy:"method:DecodeLoadBalancerInfo"`
	ProjectID         string                              `json:"project_id"`
	Description       string                              `json:"description"`
	Env               map[entity.AppEnvName]entity.AppEnv `json:"env"`
	CreateTime        string                              `json:"create_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	UpdateTime        string                              `json:"update_time" deepcopy:"timeformat:2006-01-02 15:04:05"`
	LogTTLInDays      int                                 `json:"log_ttl_in_days"`
	Namespace         string                              `json:"namespace"`
	EnableIstio       bool                                `json:"enable_istio"`
}

// AppDisplayDetailResp 应用显示详情返回值
type AppDisplayDetailResp struct {
	ID                     string                       `json:"id"`
	Name                   string                       `json:"name"`
	Type                   entity.AppType               `json:"type"`
	ServiceType            entity.AppServiceType        `json:"service_type"`
	ServiceExposeType      entity.AppServiceExposeType  `json:"service_expose_type"`
	LoadBalancerInfo       []entity.ServiceLoadBalancer `json:"load_balancer_info"`
	ServiceName            string                       `json:"service_name"`
	ClusterName            entity.ClusterName           `json:"cluster_name"`
	AliLogConfigName       string                       `json:"ali_log_config_name"`
	ProjectID              string                       `json:"project_id"`
	ProjectName            string                       `json:"project_name"`
	SentryProjectPublicDsn string                       `json:"sentry_project_public_dsn"`
	SentryProjectURL       string                       `json:"sentry_project_url" deepcopy:"method:FormatSentryProjectURL"`
	Description            string                       `json:"description"`
	CreateTime             string                       `json:"create_time"`
	UpdateTime             string                       `json:"update_time"`
	AliAlarmName           string                       `json:"ali_alarm_name"`
	EnableIstio            bool                         `json:"enable_istio"`
	// Deprecated: LogStoreName（即将废弃：减少日志仓库数量，使用LogStoreNameBasedProject替代）表示基于 项目-应用-环境建立的log store
	LogStoreName string `json:"log_store_name"`
	// LogStoreNameBasedProject 表示基于 项目建立的 log store
	LogStoreNameBasedProject       string `json:"log_store_name_based_project"`
	LogTTLInDays                   int    `json:"log_ttl_in_days"`
	ServiceProtocol                string `json:"service_protocol"`
	EnableBranchChangeNotification bool   `json:"enable_branch_change_notification"`
	EnableHotReload                bool   `json:"enable_hot_reload"`
	AppEnvExtraDetailResp

	RunningStatus      []*RunningStatusListResp     `json:"running_status"`
	Subscriptions      []string                     `json:"subscriptions"`
	InClusterDNSStatus entity.AppInClusterDNSStatus `json:"in_cluster_dns_status"`
}

// AppEnvExtraDetailResp 应用环境额外信息返回值
type AppEnvExtraDetailResp struct {
	MonitorURL string `json:"monitor_url"`
	// Deprecated: LogStoreURL 日志仓库连接（即将废弃：使用LogStoreURLBasedProject替代）
	LogStoreURL string `json:"log_store_url"`
	// LogStoreURLBasedProject 基于项目分类的日志仓库连接
	LogStoreURLBasedProject string                  `json:"log_store_url_based_project"`
	ServiceIP               string                  `json:"service_ip"`
	SLBUrl                  string                  `json:"slb_url"`
	AccessHosts             []string                `json:"access_hosts"`
	KongFrontendInfo        []*QDNSFrontendInfoResp `json:"kong_frontend_info"`
}

// AppTipsResp 应用资源配置提示返回值
type AppTipsResp struct {
	RecommendCPURequest   entity.CPUResourceType `json:"recommend_cpu_request"`
	RecommendMemRequest   entity.MemResourceType `json:"recommend_mem_request"`
	RecommendMinPodCount  string                 `json:"recommend_min_pod_count"`
	RecommendMaxPodCount  string                 `json:"recommend_max_pod_count"`
	WastedMaxCPUUsageRate string                 `json:"wasted_max_cpu_usage_rate"`
	WastedMaxMemUsageRate string                 `json:"wasted_max_mem_usage_rate"`
}

// CalculateAppRecommendResp 应用资源消耗统计返回值
type CalculateAppRecommendResp struct {
	RecommendCPURequest  entity.CPUResourceType `json:"recommend_cpu_request"`
	RecommendMemRequest  entity.MemResourceType `json:"recommend_mem_request"`
	RecommendMinPodCount string                 `json:"recommend_min_pod_count"`
	RecommendMaxPodCount string                 `json:"recommend_max_pod_count"`
}

// AppClusterKongWeightResp 应用环境集群在 QDNS 统一接入规则中的权重
type AppClusterQDNSWeightResp struct {
	ClusterName entity.ClusterName `json:"cluster_name"`
	Weight      int                `json:"weight"`
}
