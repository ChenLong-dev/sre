// 运维 QDNS 系统打通
// QDNS API文档：https://git2.qingtingfm.com/devops/qdns/-/blob/develop/doc/api.md
package req

import (
	"net/http"

	"rulai/models/entity"
)

const (
	GetQDNSRecordsPageSizeLimit  = 50 // GetQDNSRecordsPageSizeLimit 从QDNS获取解析记录分页大小上限
	GetQDNSBusinessPageSizeLimit = 50 // GetQDNSBusinessPageSizeLimit 从QDNS获取统一接入规则分页大小上限
)

// CreateQDNSRecordReq 增加QDNS解析记录请求参数
type CreateQDNSRecordReq struct {
	DomainType       entity.DomainRecordType `json:"domain_type" binding:"required"`
	DomainRecordName string                  `json:"domain_record_name" binding:"required"`
	DomainValue      string                  `json:"domain_value" binding:"required"`
	DomainName       string                  `json:"domain_name" binding:"required"`
	LineType         string                  `json:"line_type,omitempty"`
	TTL              string                  `json:"ttl,omitempty"`
	Remark           string                  `json:"remark,omitempty"`
	DomainUpdater    string                  `json:"domain_updater" binding:"required"`
	DomainController string                  `json:"domain_controller,omitempty"`
}

// DeleteQDNSRecordReq 删除QDNS解析记录请求参数
type DeleteQDNSRecordReq struct {
	DomainType       entity.DomainRecordType `json:"domain_type" binding:"required"`
	DomainRecordName string                  `json:"domain_record_name" binding:"required"`
	DomainValue      string                  `json:"domain_value" binding:"required"`
	DomainName       string                  `json:"domain_name" binding:"required"`
	LineType         string                  `json:"line_type,omitempty"`
	Remark           string                  `json:"remark,omitempty"`
	DomainUpdater    string                  `json:"domain_updater" binding:"required"`
}

// GetQDNSRecordsReq 从QDNS获取解析记录请求参数
type GetQDNSRecordsReq struct {
	DomainType       entity.DomainRecordType `form:"domain_type" json:"domain_type"`
	DomainRecordName string                  `form:"domain_record_name" json:"domain_record_name"`
	DomainValue      string                  `form:"domain_value" json:"domain_value"`
	PrivateZone      string                  `form:"private_zone" json:"private_zone"`
	PageNumber       int                     `form:"page_number" json:"page_number"`
	PageSize         int                     `form:"page_size" json:"page_size"`
}

// CreateQDNSBusinessReq 创建QDNS统一接入规则请求参数(目前没有插件需求, 不接入 plugins 字段)
type CreateQDNSBusinessReq struct {
	Env         entity.QDNSEnvName         `form:"env" json:"env"`
	UserName    string                     `form:"username" json:"username"`
	Business    string                     `form:"business,omitempty" json:"business,omitempty"` // default: ""
	Upstream    *CreateQDNSKongUpstreamReq `form:"upstream" json:"upstream"`
	Service     *CreateQDNSKongServiceReq  `form:"service,omitempty" json:"service,omitempty"`
	Routes      []*CreateQDNSKongRouteReq  `form:"routes" json:"routes"`
	Targets     []*UpsertQDNSKongTargetReq `form:"targets" json:"targets"`
	ClientToken string                     `form:"client_token" json:"client_token" binding:"required"`
}

// DeleteQDNSBusinessReq 删除QDNS统一接入规则请求参数
type DeleteQDNSBusinessReq struct {
	Env      entity.QDNSEnvName `form:"env" json:"env"`
	ID       int                `form:"id" json:"id"`
	UserName string             `form:"username" json:"username"`
}

// PatchQDNSBusinessReq 更新QDNS统一接入规则请求参数
type PatchQDNSBusinessReq struct {
	Env         entity.QDNSEnvName        `form:"env" json:"env"`
	Upstream    *PatchQDNSKongUpstreamReq `form:"upstream,omitempty" json:"upstream,omitempty"`
	Routes      []*PatchQDNSKongRouteReq  `form:"routes,omitempty" json:"routes,omitempty"`
	Service     *PatchQDNSKongServiceReq  `form:"service,omitempty" json:"service,omitempty"`
	UserName    string                    `form:"username" json:"username"`
	Business    string                    `form:"business" json:"business"`
	ClientToken string                    `form:"client_token" json:"client_token" binding:"required"`
}

// GetQDNSBusinessListReq 从QDNS获取统一接入信息列表请求参数
type GetQDNSBusinessListReq struct {
	Targets    []string `form:"targets" json:"targets"`
	PageNumber int      `form:"page_number" json:"page_number"`
	PageSize   int      `form:"page_size" json:"page_size"`
}

// UpdateQDNSKongUpstreamsTargetsHealthyReq 更新QDNS统一接入多个 Kong Upstream 下多个 Target 健康检查状态请求参数
type UpdateQDNSKongUpstreamsTargetsHealthyReq struct {
	Env       entity.QDNSEnvName                         `form:"env" json:"env"`
	Upstreams []*UpdateQDNSKongUpstreamTargetsHealthyReq `form:"upstreams" json:"upstreams"`
}

// CreateQDNSKongUpstreamReq 创建QDNS统一接入规则的 Kong Upstream 请求参数
type CreateQDNSKongUpstreamReq struct {
	// QDNS 封装后, name 自动生成, 故不支持该 field
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

// CreateQDNSKongServiceReq 创建QDNS统一接入规则的 Kong Service 请求参数
type CreateQDNSKongServiceReq struct {
	// QDNS 封装后, name 自动生成, host(upstream) 会在同一个请求中一起创建, 故不支持这两个 field
	//  host 在接入 istio 之后,需要进行改造
	Retries           *int                `json:"retries,omitempty" binding:"omitempty"`   // Default: 5
	Protocol          entity.KongProtocol `json:"protocol,omitempty" binding:"http https"` // Default: http
	Port              int                 `json:"port" binding:"required"`                 // 文档有坑, port 必须指定, 没有默认值
	Path              string              `json:"path,omitempty" binding:"omitempty"`
	ConnectTimeout    *int64              `json:"connect_timeout,omitempty" binding:"omitempty"` // Default: 60000
	WriteTimeout      *int64              `json:"write_timeout,omitempty" binding:"omitempty"`   // Default: 60000
	ReadTimeout       *int64              `json:"read_timeout,omitempty" binding:"omitempty"`    // Default: 60000
	Tags              []string            `json:"tags,omitempty" binding:"omitempty"`
	ClientCertificate *entity.KongID      `json:"client_certificate,omitempty" binding:"omitempty"`
	URL               string              `json:"url,omitempty" binding:"omitempty"`  // 快速设置 protocol://host:port/path
	Host              string              `json:"host,omitempty" binding:"omitempty"` // 指定 kong istio upstream
}

// CreateQDNSKongRouteReq 创建QDNS统一接入规则的 Kong Route 请求参数
type CreateQDNSKongRouteReq struct {
	// QDNS 封装后, name 自动生成, service 会在同一个请求中一起创建, 所以没有这两个 field
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
}

// PatchQDNSKongUpstreamReq 更新QDNS统一接入规则的 Kong Upstream 请求参数
type PatchQDNSKongUpstreamReq struct {
	ID string `json:"id" binding:"required"` // Kong Upstream ID, QDNS business 与 upstream 一一对应, 故必传 upstream_id
	// 负载均衡算法 Default: round-robin
	Algorithm entity.KongUpstreamAlgorithm `json:"algorithm,omitempty" binding:"round-robin consistent-hashing least-connections"`
	// 轮询哈希使用的字段 Default: none
	HashOn entity.KongUpstreamHash `json:"hash_on,omitempty" binding:"none consumer ip header cookie"`
	// (兜底方案)轮询哈希使用的字段 Default: none
	HashFallback       entity.KongUpstreamHash    `json:"hash_fallback,omitempty" binding:"none consumer ip header cookie"`
	HashOnHeader       string                     `json:"hash_on_header,omitempty" binding:"omitempty"`
	HashFallbackHeader string                     `json:"hash_fallback_header,omitempty" binding:"omitempty"`
	HashOnCookie       string                     `json:"hash_on_cookie,omitempty" binding:"omitempty"`
	HashOnCookiePath   string                     `json:"hash_on_cookie_path,omitempty" binding:"omitempty"` // Default: "/"
	Slots              *int                       `json:"slots,omitempty" binding:"omitempty"`               // Default: 10000
	HealthChecks       *KongUpstreamHealthCheck   `json:"healthchecks,omitempty" binding:"omitempty"`
	Tags               []string                   `json:"tags,omitempty" binding:"omitempty"`
	Targets            []*UpsertQDNSKongTargetReq `form:"targets,omitempty" json:"targets,omitempty"`
}

// PatchQDNSKongServiceReq 更新QDNS统一接入规则的 Kong Service 请求参数
type PatchQDNSKongServiceReq struct {
	// AMS 对应的k8s域名解析不应当修改 host(upstream), 故不支持这两个 field
	ID                string              `json:"id,omitempty" binding:"omitempty"`        // Kong Service ID
	Retries           *int                `json:"retries,omitempty" binding:"omitempty"`   // Default: 5
	Protocol          entity.KongProtocol `json:"protocol,omitempty" binding:"http https"` // Default: http
	Port              int                 `json:"port" binding:"required"`                 // 文档有坑, port 必须指定, 没有默认值
	Path              string              `json:"path,omitempty" binding:"omitempty"`
	ConnectTimeout    *int64              `json:"connect_timeout,omitempty" binding:"omitempty"` // Default: 60000
	WriteTimeout      *int64              `json:"write_timeout,omitempty" binding:"omitempty"`   // Default: 60000
	ReadTimeout       *int64              `json:"read_timeout,omitempty" binding:"omitempty"`    // Default: 60000
	Tags              []string            `json:"tags,omitempty" binding:"omitempty"`
	ClientCertificate *entity.KongID      `json:"client_certificate,omitempty" binding:"omitempty"`
	URL               string              `json:"url,omitempty" binding:"omitempty"` // 快速设置 protocol://host:port/path
}

// PatchQDNSKongRouteReq 更新QDNS统一接入规则的 Kong Route 请求参数
type PatchQDNSKongRouteReq struct {
	ID                      string                        `json:"id,omitempty" binding:"omitempty"`         // Kong Route ID
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
}

// UpsertQDNSKongTargetReq 创建/更新QDNS统一接入规则的 Kong Target 参数
type UpsertQDNSKongTargetReq struct {
	Target string   `json:"target" binding:"required"`
	Weight *int     `json:"weight,omitempty" binding:"omitempty"` // Default: 100, range: [0-1000]
	Tags   []string `json:"tags,omitempty" binding:"omitempty"`
}

// UpdateQDNSKongUpstreamsTargetsHealthyReq 更新QDNS统一接入单个 Kong Upstream 下多个 Target 健康检查状态请求参数
type UpdateQDNSKongUpstreamTargetsHealthyReq struct {
	ID      string           `form:"id" json:"id" binding:"required"` // Kong Upstream ID
	Targets []*entity.KongID `form:"targets" json:"targets" binding:"required"`
}
