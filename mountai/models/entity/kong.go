package entity

import (
	"net/http"

	"gitlab.shanhai.int/sre/library/base/null"
)

// KongUpstreamAlgorithm 负载均衡算法
type KongUpstreamAlgorithm string

const (
	KongUpstreamAlgorithmRoundRobin        KongUpstreamAlgorithm = "round-robin" // 免费版只支持轮询
	KongUpstreamAlgorithmConsistentHashing KongUpstreamAlgorithm = "consistent-hashing"
	KongUpstreamAlgorithmLeastConnections  KongUpstreamAlgorithm = "least-connections"
)

// KongUpstreamHash 哈希方案-决定请求如何分配至 upstream 的各个 target
type KongUpstreamHash string

const (
	KongUpstreamHashNone     KongUpstreamHash = "none"
	KongUpstreamHashConsumer KongUpstreamHash = "consumer"
	KongUpstreamHashIP       KongUpstreamHash = "ip"
	KongUpstreamHashHeader   KongUpstreamHash = "header"
	KongUpstreamHashCookie   KongUpstreamHash = "cookie"
)

// KongUpstreamHealthCheckType 健康检查类型
type KongUpstreamHealthCheckType string

const (
	KongUpstreamHealthCheckTypeHTTP  KongUpstreamHealthCheckType = "http"
	KongUpstreamHealthCheckTypeHTTPS KongUpstreamHealthCheckType = "https"
	KongUpstreamHealthCheckTypeTCP   KongUpstreamHealthCheckType = "tcp"
)

// KongProtocol service 使用的协议
type KongProtocol string

const (
	KongProtocolHTTP  KongProtocol = "http"
	KongProtocolHTTPS KongProtocol = "https"
)

// KongPathHandleBehavior 处理 service_path, route_path, request_path 的行为准则
type KongPathHandleBehavior string

const (
	KongPathHandleBehaviorV0 KongPathHandleBehavior = "v0"
	KongPathHandleBehaviorV1 KongPathHandleBehavior = "v1"
)

// KongUpstream Kong 网关 upstream 模型
// 官方文档(缺少参数, 以 API 实际返回为准): https://docs.konghq.com/enterprise/1.5.x/admin-api/#upstream-object
type KongUpstream struct {
	ID                 string                  `json:"id"`
	CreatedAt          int64                   `json:"created_at"`
	Name               string                  `json:"name"`
	Algorithm          KongUpstreamAlgorithm   `json:"algorithm"`
	HashOn             KongUpstreamHash        `json:"hash_on"`
	HashOnHeader       null.String             `json:"hash_on_header"`
	HashOnCookie       null.String             `json:"hash_on_cookie"`
	HashOnCookiePath   string                  `json:"hash_on_cookie_path"`
	HashFallback       KongUpstreamHash        `json:"hash_fallback"`
	HashFallbackHeader null.String             `json:"hash_fallback_header"`
	HostHeader         null.String             `json:"host_header"`
	Slots              int                     `json:"slots"`
	HealthChecks       KongUpstreamHealthCheck `json:"healthchecks"`
	Tags               []string                `json:"tags"`
}

// KongTarget Kong 网关 target 模型
// 官方文档: https://docs.konghq.com/enterprise/1.5.x/admin-api/#target-object
// 注意文档里没有体现出 created_at 是浮点数类型
type KongTarget struct {
	ID        string   `json:"id"`
	CreatedAt float64  `json:"created_at"`
	Upstream  *KongID  `json:"upstream"`
	Target    string   `json:"target"`
	Weight    int      `json:"weight"`
	Tags      []string `json:"tags"`
}

// KongService Kong 网关 service 模型
// 官方文档: https://docs.konghq.com/enterprise/1.5.x/admin-api/#service-object
type KongService struct {
	ID                string       `json:"id"`
	CreatedAt         int64        `json:"created_at"`
	UpdatedAt         int64        `json:"updated_at"`
	Name              string       `json:"name"`
	Retries           int          `json:"retries"`
	Protocol          KongProtocol `json:"protocol"`
	Host              string       `json:"host"`
	Port              int          `json:"port"`
	Path              null.String  `json:"path"`
	ConnectTimeout    int64        `json:"connect_timeout"`
	WriteTimeout      int64        `json:"write_timeout"`
	ReadTimeout       int64        `json:"read_timeout"`
	Tags              []string     `json:"tags"`
	ClientCertificate *KongID      `json:"client_certificate"`
}

// KongRoute Kong 网关 route 模型
// 官方文档: https://docs.konghq.com/enterprise/1.5.x/admin-api/#route-object
type KongRoute struct {
	ID                      string                 `json:"id"`
	CreatedAt               int64                  `json:"created_at"`
	UpdatedAt               int64                  `json:"updated_at"`
	Name                    string                 `json:"name"`
	Protocols               []KongProtocol         `json:"protocols"`
	Methods                 []string               `json:"methods"`
	Hosts                   []string               `json:"hosts"`
	Paths                   []string               `json:"paths"`
	Headers                 http.Header            `json:"headers"`
	HTTPSRedirectStatusCode int                    `json:"https_redirect_status_code"`
	RegexPriority           int                    `json:"regex_priority"`
	StripPath               bool                   `json:"strip_path"`
	PathHandling            KongPathHandleBehavior `json:"path_handling"`
	PreserveHost            bool                   `json:"preserve_host"`
	SNIs                    []string               `json:"snis"`
	Sources                 []string               `json:"sources"`
	Destinations            []string               `json:"destinations"`
	Tags                    []string               `json:"tags"`
	Service                 *KongID                `json:"service"`
}

// KongUpstreamHealthCheck 健康检查配置
type KongUpstreamHealthCheck struct {
	Active  KongUpstreamActiveHealthCheck  `json:"active"`
	Passive KongUpstreamPassiveHealthCheck `json:"passive"`
}

// KongUpstreamActiveHealthCheck 主动健康检查配置(Kong定时发起请求进行健康检查)
type KongUpstreamActiveHealthCheck struct {
	Healthy                KongUpstreamActiveHealthCheckHealthyConfig   `json:"healthy"`
	Unhealthy              KongUpstreamActiveHealthCheckUnhealthyConfig `json:"unhealthy"`
	HTTPSVerifyCertificate bool                                         `json:"https_verify_certificate"`
	HTTPPath               string                                       `json:"http_path"`
	Timeout                int64                                        `json:"timeout"`
	HTTPSSNI               null.String                                  `json:"https_sni"`
	Concurrency            int                                          `json:"concurrency"`
	Type                   KongUpstreamHealthCheckType                  `json:"type" binding:"oneof=http https tcp"`
}

// KongUpstreamActiveHealthCheckHealthyConfig 主动健康检查健康状态配置
type KongUpstreamActiveHealthCheckHealthyConfig struct {
	HTTPStatuses []int `json:"http_statuses"`
	Interval     int64 `json:"interval"`
	Successes    int   `json:"successes"`
}

// KongUpstreamActiveHealthCheckUnhealthyConfig 主动健康检查非健康状态配置
type KongUpstreamActiveHealthCheckUnhealthyConfig struct {
	HTTPStatuses []int `json:"http_statuses"`
	TCPFailures  int   `json:"tcp_failures"`
	Timeouts     int   `json:"timeouts"`
	HTTPFailures int   `json:"http_failures"`
	Interval     int64 `json:"interval"`
}

// KongUpstreamPassiveHealthCheck 被动健康检查配置(按照响应进行健康检查)
type KongUpstreamPassiveHealthCheck struct {
	Healthy   KongUpstreamPassiveHealthCheckHealthyConfig   `json:"healthy"`
	Unhealthy KongUpstreamPassiveHealthCheckUnhealthyConfig `json:"unhealthy"`
	Type      KongUpstreamHealthCheckType                   `json:"type" binding:"oneof=http https tcp"`
}

// KongUpstreamPassiveHealthCheckHealthyConfig 被动健康检查健康状态配置
type KongUpstreamPassiveHealthCheckHealthyConfig struct {
	Successes    int   `json:"successes"`
	HTTPStatuses []int `json:"http_statuses"`
}

// KongUpstreamPassiveHealthCheckUnhealthyConfig 被动健康检查非健康状态配置
type KongUpstreamPassiveHealthCheckUnhealthyConfig struct {
	HTTPFailures int   `json:"http_failures"`
	HTTPStatuses []int `json:"http_statuses"`
	TCPFailures  int   `json:"tcp_failures"`
	Timeouts     int   `json:"timeouts"`
}

// KongID Kong 网关资源通用 ID 信息
type KongID struct {
	ID string `json:"id"`
}
