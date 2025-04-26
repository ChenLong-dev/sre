// AMS 与蜻蜓 Kong 网关集成
// 当前使用的 Kong 网关版本: 1.5.1
// 对应 admin API 文档：https://docs.konghq.com/enterprise/1.5.x/admin-api
// NOTE: Kong admin API 在接收 JSON 数据时, {"key": null} 是非法参数(必须不传), 故暂时不能使用 library 的 null 类型
package req

import (
	"net/http"

	"rulai/models/entity"
)

// Kong 默认值
const (
	// KongServiceUnUsefulPort Kong Service 使用 upstream 作为 Host 时其实并不需要设置 Port, 但它必填
	KongServiceUnUsefulPort = 80
)

var (
	// KongDefaultRoutePaths Kong Route 默认路径参数
	KongDefaultRoutePaths = []string{"/"}
	// TODO 后面通过 kong api 查询并绑定kong upstream 名称到这个变量上
	// KongServiceDefaultIstioGatewayUpstream 默认 Istio kong upstream
	// KongServiceDefaultIstioGatewayUpstream string
)

// 一些默认值和公共只读变量地址配置
var (
	zeroInt                                        = 0
	zeroInt64                                int64 = 0
	kongaDefaultSlot                               = 1000 // Konga 定义的默认 slot 数量, Kong 默认值 100 有点小, 容易负载不均衡
	aMSKongDefaultHealthCheckTimeoutInSecond int64 = 2
	aMSKongDefaultHealthCheckSuccess               = 2
	aMSKongDefaultHealthCheckInterval        int64 = 2
	aMSKongDefaultHealthCheckFailures              = 3
	aMSKongDefaultHealthCheckTimeouts              = 2
	aMSKongServiceDefaultRetries                   = 0
	aMSKongDefaultStripPath                        = false

	ZeroIntPtr                                  = &zeroInt
	ZeroInt64Ptr                                = &zeroInt64
	KongaDefaultSlotPtr                         = &kongaDefaultSlot
	AMSKongDefaultHealthCheckTimeoutInSecondPtr = &aMSKongDefaultHealthCheckTimeoutInSecond
	AMSKongDefaultHealthCheckSuccessPtr         = &aMSKongDefaultHealthCheckSuccess
	AMSKongDefaultHealthCheckIntervalPtr        = &aMSKongDefaultHealthCheckInterval
	AMSKongDefaultHealthCheckFailuresPtr        = &aMSKongDefaultHealthCheckFailures
	AMSKongDefaultHealthCheckTimeoutsPtr        = &aMSKongDefaultHealthCheckTimeouts
	AMSKongServiceDefaultRetriesPtr             = &aMSKongServiceDefaultRetries
	AMSKongDefaultStripPath                     = &aMSKongDefaultStripPath
)

// CreateKongUpstreamReq 创建 Kong 网关 upstream 请求参数
type CreateKongUpstreamReq struct {
	Name string `json:"name" binding:"required"`
	// 负载均衡算法 Default: round-robin
	Algorithm entity.KongUpstreamAlgorithm `json:"algorithm,omitempty" binding:"round-robin consistent-hashing least-connections"`
	// 轮询哈希使用的字段 Default: none
	HashOn entity.KongUpstreamHash `json:"hash_on,omitempty" binding:"none consumer ip header cookie"`
	// (兜底方案)轮询哈希使用的字段 Default: none
	HashFallback       entity.KongUpstreamHash  `json:"hash_fallback,omitempty" binding:"none consumer ip header cookie"`
	HashOnHeader       string                   `json:"hash_on_header,omitempty" binding:"omitempty"`
	HashFallbackHeader string                   `json:"hash_fallback_header,omitempty" binding:"omitempty"`
	HashOnCookie       string                   `json:"hash_on_cookie,omitempty" binding:"omitempty"`
	HashOnCookiePath   string                   `json:"hash_on_cookie_path,omitempty" binding:"omitempty"` // Default: "/"
	Slots              *int                     `json:"slots,omitempty" binding:"omitempty"`               // Default: 10000
	HealthChecks       *KongUpstreamHealthCheck `json:"healthchecks,omitempty" binding:"omitempty"`
	Tags               []string                 `json:"tags,omitempty" binding:"omitempty"`
}

// UpdateKongUpstreamReq 更新 Kong 网关 upstream 请求参数
// (Patch方式, eg: 传递 {"tags": ["tag-1"]} 将会把 tags 字段全量更新为 ["tag-1"], 保持其他字段不变化)
type UpdateKongUpstreamReq struct {
	Name               string                       `json:"name,omitempty" binding:"omitempty"`
	Algorithm          entity.KongUpstreamAlgorithm `json:"algorithm,omitempty" binding:"round-robin consistent-hashing least-connections"`
	HashOn             entity.KongUpstreamHash      `json:"hash_on,omitempty" binding:"none consumer ip header cookie"`
	HashFallback       entity.KongUpstreamHash      `json:"hash_fallback,omitempty" binding:"none consumer ip header cookie"`
	HashOnHeader       string                       `json:"hash_on_header,omitempty" binding:"omitempty"`
	HashFallbackHeader string                       `json:"hash_fallback_header,omitempty" binding:"omitempty"`
	HashOnCookie       string                       `json:"hash_on_cookie,omitempty" binding:"omitempty"`
	HashOnCookiePath   string                       `json:"hash_on_cookie_path,omitempty" binding:"omitempty"` // Default: "/"
	Slots              *int                         `json:"slots,omitempty" binding:"omitempty"`               // Default: 10000
	HealthChecks       *KongUpstreamHealthCheck     `json:"healthchecks,omitempty" binding:"omitempty"`
	Tags               []string                     `json:"tags,omitempty" binding:"omitempty"`
}

// UpsertKongTargetReq 创建/更新/删除 Kong 网关指定 upstream 下的 target 请求参数
type UpsertKongTargetReq struct {
	Target string   `json:"target" binding:"required"`
	Weight *int     `json:"weight,omitempty" binding:"omitempty"` // Default: 100, range: [0-1000]
	Tags   []string `json:"tags,omitempty" binding:"omitempty"`
}

// CreateKongServiceReq 创建 Kong 网关 service 请求参数
type CreateKongServiceReq struct {
	Name              string              `json:"name" binding:"required"`
	Retries           *int                `json:"retries,omitempty" binding:"omitempty"`   // Default: 5
	Protocol          entity.KongProtocol `json:"protocol,omitempty" binding:"http https"` // Default: http
	Host              string              `json:"host" binding:"required"`
	Port              int                 `json:"port" binding:"required"` // 文档有坑, port 必须指定, 没有默认值
	Path              string              `json:"path,omitempty" binding:"omitempty"`
	ConnectTimeout    *int64              `json:"connect_timeout,omitempty" binding:"omitempty"` // Default: 60000
	WriteTimeout      *int64              `json:"write_timeout,omitempty" binding:"omitempty"`   // Default: 60000
	ReadTimeout       *int64              `json:"read_timeout,omitempty" binding:"omitempty"`    // Default: 60000
	Tags              []string            `json:"tags,omitempty" binding:"omitempty"`
	ClientCertificate *entity.KongID      `json:"client_certificate,omitempty" binding:"omitempty"`
	URL               string              `json:"url,omitempty" binding:"omitempty"` // 快速设置 protocol://host:port/path
}

// CreateKongRouteReq 创建 Kong 网关 route 请求参数
type CreateKongRouteReq struct {
	Name                    string                        `json:"name" binding:"required"`
	Protocols               []entity.KongProtocol         `json:"protocols,omitempty" binding:"http https"` // Default: ["http", "https"]
	Methods                 []string                      `json:"methods,omitempty" binding:"omitempty"`    // 不填代表不限制
	Hosts                   []string                      `json:"hosts"  binding:"omitempty"`
	Paths                   []string                      `json:"paths,omitempty"  binding:"omitempty"`
	Headers                 http.Header                   `json:"headers,omitempty"  binding:"omitempty"`
	HTTPSRedirectStatusCode *int                          `json:"https_redirect_status_code,omitempty" binding:"omitempty"` // Default: 426
	RegexPriority           *int                          `json:"regex_priority,omitempty" binding:"omitempty"`             // Default: 0
	StripPath               *bool                         `json:"strip_path,omitempty" binding:"omitempty"`                 // Default: true
	PathHandling            entity.KongPathHandleBehavior `json:"path_handling,omitempty" binding:"v0 v1"`                  // Default: v1
	PreserveHost            *bool                         `json:"preserve_host,omitempty" binding:"omitempty"`              // Default: false
	SNIs                    []string                      `json:"snis,omitempty" binding:"omitempty"`
	Sources                 []string                      `json:"sources,omitempty" binding:"omitempty"`
	Destinations            []string                      `json:"destinations,omitempty" binding:"omitempty"`
	Tags                    []string                      `json:"tags,omitempty" binding:"omitempty"`
	Service                 *entity.KongID                `json:"service,omitempty" binding:"omitempty"` // 当前在 url path 中填 service_name
}

// KongUpstreamHealthCheck 健康检查配置
type KongUpstreamHealthCheck struct {
	Active  *KongUpstreamActiveHealthCheck  `json:"active,omitempty" binding:"omitempty"`
	Passive *KongUpstreamPassiveHealthCheck `json:"passive,omitempty" binding:"omitempty"`
}

// KongUpstreamActiveHealthCheck 主动健康检查配置(Kong定时发起请求进行健康检查)
type KongUpstreamActiveHealthCheck struct {
	Healthy   *KongUpstreamActiveHealthCheckHealthyConfig   `json:"healthy,omitempty" binding:"omitempty"`
	Unhealthy *KongUpstreamActiveHealthCheckUnhealthyConfig `json:"unhealthy,omitempty" binding:"omitempty"`

	HTTPSVerifyCertificate *bool  `json:"https_verify_certificate,omitempty" binding:"omitempty"` // Default: true
	HTTPPath               string `json:"http_path,omitempty" binding:"omitempty"`                // Default: "/"
	Timeout                *int64 `json:"timeout,omitempty" binding:"omitempty"`                  //Default: 1
	HTTPSSNI               string `json:"https_sni,omitempty" binding:"omitempty"`
	Concurrency            *int   `json:"concurrency,omitempty" binding:"omitempty"` // Default: 10

	Type entity.KongUpstreamHealthCheckType `json:"type,omitempty" binding:"oneof=http https tcp"` // Default: http
}

// KongUpstreamActiveHealthCheckHealthyConfig 主动健康检查健康状态配置
type KongUpstreamActiveHealthCheckHealthyConfig struct {
	HTTPStatuses []int  `json:"http_statuses,omitempty" binding:"omitempty"` // Default: [200,302]
	Interval     *int64 `json:"interval,omitempty" binding:"omitempty"`      // Default: 0, 单位: 秒
	Successes    *int   `json:"successes,omitempty" binding:"omitempty"`     // Default: 0
}

// KongUpstreamActiveHealthCheckUnhealthyConfig 主动健康检查非健康状态配置
type KongUpstreamActiveHealthCheckUnhealthyConfig struct {
	HTTPStatuses []int  `json:"http_statuses,omitempty" binding:"omitempty"` // Default: [429,404,500,501,502,503,504,505]
	TCPFailures  *int   `json:"tcp_failures,omitempty" binding:"omitempty"`  // Default: 0
	Timeouts     *int   `json:"timeouts,omitempty" binding:"omitempty"`      // Default: 0
	HTTPFailures *int   `json:"http_failures,omitempty" binding:"omitempty"` // Default: 0
	Interval     *int64 `json:"interval,omitempty" binding:"omitempty"`      // Default: 0, 单位: 秒
}

// KongUpstreamPassiveHealthCheck 被动健康检查配置(按照响应进行健康检查)
type KongUpstreamPassiveHealthCheck struct {
	Healthy   *KongUpstreamPassiveHealthCheckHealthyConfig   `json:"healthy,omitempty" binding:"omitempty"`
	Unhealthy *KongUpstreamPassiveHealthCheckUnhealthyConfig `json:"unhealthy,omitempty" binding:"omitempty"`
	Type      entity.KongUpstreamHealthCheckType             `json:"type,omitempty" binding:"oneof=http https tcp"` // Default: http
}

// KongUpstreamPassiveHealthCheckHealthyConfig 被动健康检查健康状态配置
type KongUpstreamPassiveHealthCheckHealthyConfig struct {
	Successes *int `json:"successes,omitempty" binding:"omitempty"` // Default: 0
	// Default: [200,201,202,203,204,205,206,207,208,226,300,301,302,303,304,305,306,307,308]
	HTTPStatuses []int `json:"http_statuses,omitempty" binding:"omitempty"`
}

// KongUpstreamPassiveHealthCheckUnhealthyConfig 被动健康检查非健康状态配置
type KongUpstreamPassiveHealthCheckUnhealthyConfig struct {
	HTTPFailures *int  `json:"http_failures,omitempty" binding:"omitempty"` // Default: 0
	HTTPStatuses []int `json:"http_statuses,omitempty" binding:"omitempty"` // Default: [429,500,503]
	TCPFailures  *int  `json:"tcp_failures,omitempty" binding:"omitempty"`  // Default: 0
	Timeouts     *int  `json:"timeouts,omitempty" binding:"omitempty"`      // Default: 0
}

// GetKongUpstreamsReq  通过 tags 获取 upstreams 列表
type GetKongUpstreamsReq struct {
	EnvName entity.AppEnvName `json:"env_name,omitempty" binding:"omitempty"` // kong 环境
	Tags    string            `json:"tags,omitempty" binding:"omitempty"`     // 检索 tags
	Next    string            `json:"next,omitempty" binding:"omitempty"`     // 分页 token
	Offset  string            `json:"offset,omitempty" binding:"omitempty"`   // 分页偏移
	Size    string            `json:"size,omitempty" binding:"size"`          // 每页记录数
}

// GetKongUpstreamTargetsReq 获取某个 upstream 下的所有 target
type GetKongUpstreamTargetsReq struct {
	EnvName      entity.AppEnvName `json:"env_name,omitempty" binding:"omitempty"`      // kong 环境
	UpstreamID   string            `json:"upstream_id,omitempty" binding:"omitempty"`   // upstream ID
	UpstreamName string            `json:"upstream_name,omitempty" binding:"omitempty"` // Upstream Name
	Next         string            `json:"next,omitempty" binding:"omitempty"`          //  分页 token
	Offset       string            `json:"offset,omitempty" binding:"omitempty"`        // 分页偏移
	Size         string            `json:"size,omitempty" binding:"size"`               // 每页记录数
}

// DeleteKongTargetReq 删除指定 target
type DeleteKongTargetReq struct {
	ID           string `json:"id,omitempty" binding:"omitempty"`            // target ID
	Host         string `json:"host,omitempty" binding:"omitempty"`          // ip:port
	EnvName      string `json:"env_name,omitempty" binding:"omitempty"`      // kong 环境
	UpstreamID   string `json:"upstream_id,omitempty" binding:"omitempty"`   // upstream ID
	UpstreamName string `json:"upstream_name,omitempty" binding:"omitempty"` // Upstream Name
}

// CreateKongTargetReq 创建 kong target
type CreateKongTargetReq struct {
	UpstreamID   string `json:"upstream_id,omitempty" binding:"omitempty"`   // upstream ID
	Target       string `json:"target,omitempty" binding:"omitempty"`        // target
	EnvName      string `json:"env_name,omitempty" binding:"omitempty"`      // kong 环境
	UpstreamName string `json:"upstream_name,omitempty" binding:"omitempty"` // Upstream Name
}

type GetKongServicesReq struct {
	Tags    string `json:"tags" binding:"omitempty"`
	Offset  string `json:"offset" binding:"omitempty"`
	Size    string `json:"size" binding:"omitempty"`
	EnvName string `json:"env_name" binding:"omitempty"`
}
