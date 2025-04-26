package config

import (
	framework "gitlab.shanhai.int/sre/app-framework"
	infraJenkins "gitlab.shanhai.int/sre/gojenkins"
	"gitlab.shanhai.int/sre/library/base/ctime"
	"gitlab.shanhai.int/sre/library/database/mongo"
	"gitlab.shanhai.int/sre/library/database/redis"
	"gitlab.shanhai.int/sre/library/database/sql"
	"gitlab.shanhai.int/sre/library/kafka"
	"gitlab.shanhai.int/sre/library/net/httpclient"
	"gitlab.shanhai.int/sre/library/net/redlock"
)

var (
	// Conf 全局配置文件
	Conf *Config
)

// AliConfig 阿里云配置
type AliConfig struct {
	PrivateZonePrefix string `yaml:"privateZonePrefix"`
}

type GitConfig struct {
	Host string `yaml:"host"`
	// gitlab基础token，用于获取项目，分支，成员等信息
	Token string `yaml:"token"`
	// CI流程gitlab token，用于跑CI流程添加hook
	CIToken string `yaml:"ciToken"`
}

// JWTConfig JWT鉴权配置
type JWTConfig struct {
	SignKey             string   `yaml:"signKey"`
	K8sSystemUserTokens []string `yaml:"k8sSystemUserTokens"`
}

// QTConfigManagerConfig 蜻蜓配置中心管理器配置
type QTConfigManagerConfig struct {
	Host  string `yaml:"host"`
	Token string `yaml:"token"`
}

// OtherConfig 其他配置(杂项)
type OtherConfig struct {
	K8sStgClusterID string `yaml:"k8sStgClusterID"`
	K8sPrdClusterID string `yaml:"k8sPrdClusterID"`

	AliSLSConsoleURL     string `yaml:"aliSlsConsoleUrl"`
	AliLogProjectStgName string `yaml:"aliLogProjectStgName"`
	AliLogProjectPrdName string `yaml:"aliLogProjectPrdName"`

	AliSLBConsoleURL string `yaml:"aliSlbConsoleUrl"`
	AliK8sConsoleURL string `yaml:"aliK8sConsoleUrl"`

	AMSHost string `yaml:"amsHost"`

	ConfigCenter QTConfigCenterConfig `yaml:"configCenter"`

	AmsFrontendHost string `yaml:"amsFrontendHost"`

	InternalUserHost string `yaml:"internalUserHost"`

	BuildJobTimeout                int            `yaml:"buildJobTimeout"`
	InClusterDNSChangeWaitDuration ctime.Duration `yaml:"inClusterDNSChangeWaitDuration"`
	IngressChangeWaitDuration      ctime.Duration `yaml:"ingressChangeWaitDuration"`

	HWConsoleURL string `yaml:"hwConsoleUrl"`
}

// QTConfigCenterConfig 蜻蜓配置中心配置
type QTConfigCenterConfig struct {
	ProjectID string `yaml:"projectID"`
	Branch    string `yaml:"branch"`
}

// PrometheusConfig 普罗米修斯监控服务配置
type PrometheusConfig struct {
	StgHost               string `yaml:"stgHost"`
	PrdHost               string `yaml:"prdHost"`
	StgContainerLabelName string `yaml:"stgContainerLabelName"`
	PrdContainerLabelName string `yaml:"prdContainerLabelName"`
}

// K8sConfig Kubernetes配置
type K8sConfig struct {
	KubeConfigPath string `yaml:"kubeConfigPath"`
	StgContextName string `yaml:"stgContextName"`
	PrdContextName string `yaml:"prdContextName"`
}

// K8sClusterConfig Kubernetes集群配置
type K8sClusterConfig struct {
	Name                string `yaml:"name"`
	ContextName         string `yaml:"contextName"`
	KubeConfigPath      string `yaml:"kubeConfigPath"`
	TLSSecretName       string `yaml:"tlsSecretName"`
	Vendor              string `yaml:"vendor"`
	ClusterID           string `yaml:"clusterID"`
	ClusterNameInVendor string `yaml:"clusterNameInVendor"`
	// 日志桶配置(统一命名为 log_bucket), 实际对应: 阿里-log_project, 华为-log_group
	LogBucketID       string   `yaml:"logBucketID"`
	LogBucketName     string   `yaml:"logBucketName"`
	Region            string   `yaml:"region"`
	IngressClass      string   `yaml:"ingressClass"`
	GrafanaPodPath    string   `yaml:"grafanaPodPath"`
	GrafanaAppPath    string   `yaml:"grafanaAppPath"`
	GrafanaHost       string   `yaml:"grafanaHost"`
	LocalDNS          string   `yaml:"localDNS"`
	VisibleProjectIDs []string `yaml:"visibleProjectIds"`
}

// KongConfig Kong网关配置
type KongConfig struct {
	Envs  map[string]*KongEnvConfig `yaml:"envs"`  // 留意线上环境实际不止一个 Kong 集群，目前 AMS 仅使用 int 集群
	Hosts []string                  `yaml:"hosts"` // kong 生产集群多个管理地址
}

// KongEnvConfig Kong网关环境配置
type KongEnvConfig struct {
	AdminHost []string `yaml:"adminHost"`
	Address   string   `yaml:"address"` // 不一定是 LB
	KongLB    string   `yaml:"kongLB"`
}

// QDNSConfig QDNS动态域名中心配置
type QDNSConfig struct {
	Host string `yaml:"host"`
}

// ApolloConfig Apollo配置中心配置
type ApolloConfig struct {
	StgHost string `yaml:"stgHost"`
	PrdHost string `yaml:"prdHost"`
}

// SentrySystemConfig Sentry报警服务配置
type SentrySystemConfig struct {
	Organization string `yaml:"organization"`
	AuthToken    string `yaml:"authToken"`
	Host         string `yaml:"host"`
}

// Ding approval process config.
type DingApprovalConfig struct {
	ProcessCode string `yaml:"processCode"`
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	Host        string              `yaml:"host"`
	App         string              `yaml:"app"`
	Token       string              `yaml:"token"`
	GroupTokens map[string]string   `yaml:"groupTokens"`
	Approval    *DingApprovalConfig `yaml:"approval"`
}

// AppOpConsumerConfig 应用操作kafka配置
type AppOpConsumerConfig struct {
	Topic   string `yaml:"topic"`
	GroupID string `yaml:"groupID"`
}

// create shipper config
type ColdStorageConfig struct {
	OSSStgBucket   string `yaml:"ossStgBucket"`
	OSSPrdBucket   string `yaml:"ossPrdBucket"`
	RoleArn        string `yaml:"roleArn"`
	CompressType   string `yaml:"compressType"`
	BufferInterval int    `yaml:"bufferInterval"`
	BufferSize     int    `yaml:"bufferSize"`
	PathFormat     string `yaml:"pathFormat"`
	Format         string `yaml:"format"`
}

type JenkinsCIConfig struct {
	GoJenkins             *infraJenkins.Config `yaml:"goJenkins"`
	GitlabSecretToken     string               `yaml:"gitlabSecretToken"`
	PipelineBranch        string               `yaml:"pipelineBranch"`
	ScriptPath            string               `yaml:"scriptPath"`
	GitLabConnection      string               `yaml:"gitLabConnection"`
	PipelineURL           string               `yaml:"pipelineURL"`
	PipelineCredentialsID string               `yaml:"pipelineCredentialsID"`
}

// Ding approval callback consumer config.
type ApprovalCallbackConsumerConfig struct {
	Topic   string `yaml:"topic"`
	GroupID string `yaml:"groupID"`
}

// VendorConfig 云服务商配置项(与集群关联)
type VendorConfig struct {
	Name                string               `yaml:"name"`
	DisableLogConfig    bool                 `yaml:"disableLogConfig"`    // 云服务商日志接入开关(未来可能弃用, 所以使用 disable 条件)
	ImageRegistryConfig *ImageRegistryConfig `yaml:"imageRegistryConfig"` // 云服务商对应的镜像仓库配置
	AccessKeyID         string               `yaml:"accessKeyID"`
	AccessKeySecret     string               `yaml:"accessKeySecret"`
	LogEndpoint         string               `yaml:"logEndpoint"`
	RegionID            string               `yaml:"regionID"` // 各服务商对区域的命名其实不同: 阿里-region_id, 华为-project_id
}

// ImageRegistryConfig 云服务商对应的镜像仓库配置
type ImageRegistryConfig struct {
	Host      string `yaml:"host"`
	Namespace string `yaml:"namespace"`
}

type Feishu struct {
	Host             string `yaml:"host"`
	DeployNotiChatID string `yaml:"deployNotiChatID"`
}

// Config 配置文件
type Config struct {
	// 基础配置文件
	*framework.Config `yaml:",inline"`

	HTTPClient         *httpclient.Config              `yaml:"httpClient"`
	Redis              *redis.Config                   `yaml:"redis"`
	ApolloPrdMysql     *sql.Config                     `yaml:"apolloPrdMysql"`
	ApolloStgMysql     *sql.Config                     `yaml:"apolloStgMysql"`
	Redlock            *redlock.Config                 `yaml:"redlock"`
	Mongo              *mongo.Config                   `yaml:"mongo"`
	Ali                *AliConfig                      `yaml:"ali"`
	Jenkins            *infraJenkins.Config            `yaml:"jenkins"`
	JenkinsCI          *JenkinsCIConfig                `yaml:"jenkinsCI"`
	Git                *GitConfig                      `yaml:"git"`
	JWT                *JWTConfig                      `yaml:"jwt"`
	K8s                *K8sConfig                      `yaml:"k8s"`
	K8sClusters        map[string][]*K8sClusterConfig  `yaml:"k8sClusters"`
	Kong               *KongConfig                     `yaml:"kong"`
	ConfigManager      *QTConfigManagerConfig          `yaml:"cm"`
	Other              *OtherConfig                    `yaml:"other"`
	Prometheus         *PrometheusConfig               `yaml:"prometheus"`
	Apollo             *ApolloConfig                   `yaml:"apollo"`
	SentrySystem       *SentrySystemConfig             `yaml:"sentrySystem"`
	KafkaProducer      *kafka.Config                   `yaml:"kafkaProducer"`
	DingTalk           *DingTalkConfig                 `yaml:"dingtalk"`
	AppOpConsumer      *AppOpConsumerConfig            `yaml:"appOpConsumer"`
	ColdStorage        *ColdStorageConfig              `yaml:"coldStorage"`
	QDNS               *QDNSConfig                     `yaml:"qdns"`
	ApprovalConsumer   *ApprovalCallbackConsumerConfig `yaml:"approvalConsumer"`
	Vendors            []*VendorConfig                 `yaml:"vendors"`
	OneTimeJobMaxCount int                             `yaml:"oneTimeJobMaxCount"`
	IstioOnEnv         []string                        `yaml:"istioOnEnv"` // 用于控制哪些环境中已经可以 istio 部署
	Feishu             *Feishu                         `yaml:"feishu"`
}

// Read 读取并加载配置文件
func Read(param string) *Config {
	Conf = new(Config)

	// 解码yaml配置文件
	framework.DecodeConfig(param, Conf)

	return Conf
}
