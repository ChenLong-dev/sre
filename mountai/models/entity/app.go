package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AppType 应用类型
type AppType string

// 已支持的应用类型枚举值
const (
	// 服务类型
	AppTypeService AppType = "Service"
	// 常驻任务类型
	AppTypeWorker AppType = "Worker"
	// 定时任务类型
	AppTypeCronJob AppType = "CronJob"
	// 一次性任务类型
	AppTypeOneTimeJob AppType = "OneTimeJob"
)

// AppServiceType 应用的服务类型
// 只有在类型为服务(kubernetes Service)时才存在
type AppServiceType string

// 已支持的应用的服务类型枚举值
const (
	// Restful服务类型
	AppServiceTypeRestful AppServiceType = "Restful"
	// GRPC服务类型
	AppServiceTypeGRPC AppServiceType = "GRPC"
)

// AppServiceExposeType 服务暴露方式
type AppServiceExposeType string

const (
	// AppServiceExposeTypeInternal 不暴露至外部
	AppServiceExposeTypeInternal AppServiceExposeType = "Internal"
	// AppServiceExposeTypeIngress Ingress 暴露方式
	AppServiceExposeTypeIngress AppServiceExposeType = "Ingress"
	// AppServiceExposeTypeLB LB 暴露方式
	AppServiceExposeTypeLB AppServiceExposeType = "LB"
)

// AppEnvName 应用的环境名
type AppEnvName string

// 已支持的应用的环境名枚举值
const (
	// 单元测试环境
	AppEnvFat   AppEnvName = "fat"
	IstioEnvFat            = IstioNamespacePrefix + AppEnvFat

	// 测试环境
	AppEnvStg   AppEnvName = "test"
	IstioEnvStg            = IstioNamespacePrefix + AppEnvStg

	// 预发布环境
	AppEnvPre   AppEnvName = "pre"
	IstioEnvPre            = IstioNamespacePrefix + AppEnvPre

	// 线上环境
	AppEnvPrd   AppEnvName = "prod"
	IstioEnvPrd            = IstioNamespacePrefix + AppEnvPrd
)

// 已支持的负载均衡协议枚举值
const (
	// 七层负载均衡协议
	LoadBalancerProtocolHTTP = "http"
	// 四层负载均衡协议
	LoadBalancerProtocolTCP = "tcp"
)

// AppInClusterDNSStatus describe
type AppInClusterDNSStatus string

const (
	// AppInClusterDNSStatusUnSupported 不支持集群内DNS的变更
	AppInClusterDNSStatusUnSupported AppInClusterDNSStatus = "unsupported"
	// AppInClusterDNSStatusDisabled DNS处于禁用状态
	AppInClusterDNSStatusDisabled AppInClusterDNSStatus = "disabled"
	// AppInClusterDNSStatusEnabled DNS处于启用状态
	AppInClusterDNSStatusEnabled AppInClusterDNSStatus = "enabled"
	// AppInClusterDNSStatusUpdating DNS更新中
	AppInClusterDNSStatusUpdating AppInClusterDNSStatus = "updating"
)

// 已支持的环境名组
var (
	// AppNormalEnvNames 常见应用的环境名组
	// AppNormalEnvNames = []AppEnvName{AppEnvFat, AppEnvStg, AppEnvPre, AppEnvPrd, IstioEnvStg, IstioEnvFat,
	// 	IstioEnvPre, IstioEnvPrd}
	AppNormalEnvNames = []AppEnvName{AppEnvFat, AppEnvStg, AppEnvPrd}
)

// LogTTLInDaysEmpty 未填写日志保存天数默认值
const LogTTLInDaysEmpty = 0

// App 应用
type App struct {
	// 应用id
	ID primitive.ObjectID `bson:"_id" json:"_id"`
	// 应用名
	Name string `bson:"name" json:"name"`
	// 应用类型
	Type AppType `bson:"type" json:"type"`
	// 项目id
	ProjectID string `bson:"project_id" json:"project_id"`
	// 环境相关配置
	Env map[AppEnvName]AppEnv `bson:"env" json:"env"`

	// 应用的服务类型
	// 仅当类型为服务时，才有效
	ServiceType AppServiceType `bson:"service_type" json:"service_type"`
	// 服务暴露方式
	// 仅当服务是 Restful 才有效
	ServiceExposeType AppServiceExposeType `bson:"service_expose_type" json:"service_expose_type"`
	// 仅当服务是 Restful 且通过 LB 方式暴露服务时的 LB 信息
	LoadBalancerInfo []ServiceLoadBalancer `bson:"load_balancer_info" json:"load_balancer_info"`
	// 应用的K8s服务名
	// 由于改变服务名会导致服务ip变化，所以需在生成应用时初始化完毕
	ServiceName string `bson:"service_name" json:"service_name"`
	// 阿里云日志采集CRD的名称
	// 由于更换CRD会导致日志丢失，所以需在生成应用时初始化完毕
	AliLogConfigName string `bson:"ali_log_config_name" json:"ali_log_config_name"`
	// 日志存储时间（天）
	LogTTLInDays int `bson:"log_ttl_in_days" json:"log_ttl_in_days"`

	CreateTime *time.Time `bson:"create_time" json:"create_time"`
	UpdateTime *time.Time `bson:"update_time" json:"update_time"`
	// 软删除
	DeleteTime *time.Time `bson:"delete_time" json:"delete_time"`
	// 应用的sentry public dsn
	SentryProjectPublicDsn string `bson:"sentry_project_public_dsn" json:"sentry_project_public_dsn"`
	// 应用的sentry项目名称
	SentryProjectSlug string `bson:"sentry_project_slug" json:"sentry_project_slug"`
	// 应用的描述
	Description string `bson:"description" json:"description"`
	// 是否启用 istio
	EnableIstio bool `json:"enable_istio" bson:"enable_istio"`
}

// TableName 应用数据库表名
func (*App) TableName() string {
	return "app"
}

// GenerateObjectIDString 生成应用 MongoID
func (item *App) GenerateObjectIDString(args map[string]interface{}) string {
	return item.ID.Hex()
}

func (item *App) DecodeLoadBalancerInfo(_ map[string]interface{}) []ServiceLoadBalancer {
	if item.LoadBalancerInfo == nil {
		return make([]ServiceLoadBalancer, 0)
	}
	return item.LoadBalancerInfo
}

// AppEnv 应用与环境有关元数据
type AppEnv struct {
	// 阿里云告警名
	// 由于更换告警过于复杂，并可能会丢失，所以需在生成应用时初始化完毕
	AliAlarmName string `bson:"ali_alarm_name" json:"ali_alarm_name"`
	// 负载均衡协议类型
	// 由于无法更换服务的负载均衡协议，所以需在生成应用时初始化完毕
	ServiceProtocol string `bson:"service_protocol" json:"service_protocol"`
	// 阿里云日志仓库名
	// Deprecated: 由于更换日志仓库名会导致日志丢失，所以需在生成应用时初始化完毕（即将废弃：使用entity.project中定义的LogStoreName替代以减少logstore数量）
	LogStoreName string `bson:"log_store_name" json:"log_store_name"`
	// LogTailName 应用首次创建时确定，修改应用名不会更改
	LogTailName string `bson:"log_tail_name" json:"log_tail_name"`
	// 分支改变时发送通知
	EnableBranchChangeNotification bool `bson:"enable_branch_change_notification" json:"enable_branch_change_notification"`
	// 是否支持热部署
	EnableHotReload bool `bson:"enable_hot_reload" json:"enable_hot_reload"`
}

// ServiceLoadBalancer Restful 应用，如果通过 LB 方式暴露服务
type ServiceLoadBalancer struct {
	Env                AppEnvName  `bson:"env" json:"env"`
	Cluster            ClusterName `bson:"cluster" json:"cluster"`
	LoadBalancerID     string      `bson:"load_balancer_id" json:"load_balancer_id"`
	LoadbalancerCertId string      `bson:"load_balancer_cert_id" json:"load_balancer_cert_id"`
}

const (
	Istio                      = "istio"
	IstioNamespacePrefix       = Istio + "-"
	IstioUpstreamName          = "k8s-cluster-istio-upstream"
	IstioNamespace             = "istio-system"
	IstioServiceIngressGateway = "istio-ingressgateway"

	KongDefaultCluster                 = KongClusterDev
	KongClusterDev     KongClusterName = "dev"
	KongClusterStg     KongClusterName = "stg"
	KongClusterPrd     KongClusterName = "prd"
)

type KongClusterName string
